# 🌐 Rota das Coisas — Malha IoT Distribuída

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![HTML5](https://img.shields.io/badge/HTML5-E34F26?style=for-the-badge&logo=html5&logoColor=white)
![JavaScript](https://img.shields.io/badge/JavaScript-323330?style=for-the-badge&logo=javascript&logoColor=F7DF1E)

Projeto desenvolvido para a disciplina de **Concorrência e Conectividade** (Engenharia de Computação — UEFS). Simula uma malha de Internet das Coisas (IoT) distribuída com foco em telemetria contínua, controle remoto de atuadores e processamento concorrente.

---

## 📋 Sumário

1. [Arquitetura do Sistema](#1-arquitetura-do-sistema)
2. [Comunicação e Protocolos](#2-comunicação-e-protocolos)
3. [Encapsulamento de Dados](#3-encapsulamento-de-dados)
4. [Concorrência](#4-concorrência)
5. [Estrutura de Arquivos](#5-estrutura-de-arquivos)
6. [Como Executar](#6-como-executar)
7. [Equipamentos Simulados](#7-equipamentos-simulados)
8. [Interface do Cliente](#8-interface-do-cliente)

---

## 1. Arquitetura do Sistema

O sistema foi desenhado para eliminar o alto acoplamento (ponto-a-ponto) utilizando o padrão **Broker**. A arquitetura é composta por três camadas isoladas que se comunicam exclusivamente via rede:

```
┌──────────────┐   UDP (telemetria)   ┌─────────────────┐   TCP (comandos)   ┌──────────────┐
│   SENSORES   │ ──────────────────▶  │   INTEGRADOR    │ ─────────────────▶ │   ATUADOR    │
│  (sensor_bin)│                      │ (integrador_bin)│                    │ (atuador_bin)│
└──────────────┘                      └────────┬────────┘                    └──────┬───────┘
                                               │ SSE (stream)                       │
                                               ▼                                    │ (física)
                                      ┌──────────────────┐                   ┌──────┴───────┐
                                      │   CLIENTE WEB    │                   │  /fisica/*.txt│
                                      │  (Nginx + HTML)  │                   └──────────────┘
                                      └──────────────────┘
```

**Componentes:**

- **Sensores** — Leem a temperatura atual do arquivo de física do ambiente e enviam telemetria ao Integrador via UDP a cada segundo.
- **Atuador** — Recebe comandos TCP e altera a temperatura dos arquivos físicos simulando o comportamento de um motor (compressor ligado = resfria, desligado = aquece).
- **Integrador** — Centraliza o recebimento de telemetria, aplica regras de negócio (limites de temperatura) e roteia comandos. Expõe uma API HTTP com streaming SSE para o cliente.
- **Cliente** — Dashboard web servido por Nginx que consome o stream de dados e emite comandos de controle ao Integrador via HTTP POST.

---

## 2. Comunicação e Protocolos

A comunicação foi projetada para lidar com dois perfis de tráfego distintos, otimizando a latência de acordo com a criticidade de cada mensagem:

| Fluxo | Protocolo | Porta | Justificativa |
|---|---|---|---|
| Sensor → Integrador | **UDP** | 9090 | Telemetria volumosa e contínua. A perda eventual de um pacote não compromete o sistema; prioriza baixa latência e elimina o overhead do handshake TCP. |
| Integrador → Atuador | **TCP** | 8081 | Comandos críticos (ligar/desligar motor). A entrega garantida é obrigatória. |
| Cliente → Integrador | **HTTP POST** | 8080 | Comando manual do operador. Requer confirmação (resposta HTTP). |
| Integrador → Cliente | **SSE (HTTP)** | 8080 | Server-Sent Events: um único túnel HTTP persistente. O servidor empurra os dados assim que chegam, eliminando completamente o overhead de *polling* (requisições periódicas do navegador). |

---

## 3. Encapsulamento de Dados

Todos os dados trafegam pela rede encapsulados em **JSON**, serializados e desserializados pelas structs do pacote `shared`.

**Payload de Telemetria** (Sensor → Integrador via UDP):
```json
{
  "tipo": "telemetria",
  "payload": {
    "equipamento_id": 1,
    "tipo_equipamento": "geladeira",
    "temperatura": 4.10,
    "temp_min": 2.0,
    "temp_max": 4.5,
    "umidade": 50.0,
    "timestamp": "2026-04-09T10:00:00Z"
  }
}
```

**Payload de Comando** (Integrador → Atuador via TCP):
```json
{
  "equipamento_id": 1,
  "tipo": "diminuir_temperatura",
  "timestamp": "2026-04-09T10:00:01Z"
}
```

**Tipos de Comando disponíveis** (definidos em `shared/comando.go`):

| Constante | Valor | Efeito no Atuador |
|---|---|---|
| `LigarEquipamento` | `ligar_equipamento` | — |
| `DesligarEquipamento` | `desligar_equipamento` | Desliga o compressor |
| `AumentarTemperatura` | `aumentar_temperatura` | Desliga o compressor (aquecimento natural) |
| `DiminuirTemperatura` | `diminuir_temperatura` | **Liga o compressor** (resfriamento ativo) |

---

## 4. Concorrência

O sistema faz uso intenso de concorrência para lidar com múltiplos dispositivos simultâneos:

- **Goroutines no Integrador** — Cada pacote UDP recebido é processado em uma goroutine independente (`go processarTelemetria(...)`), permitindo que o servidor não bloqueie enquanto processa uma leitura.
- **`sync.Mutex` no Integrador** — O mapa `historico` é compartilhado entre a goroutine de processamento UDP e o handler HTTP de streaming. O `mutex` garante que leituras e escritas nunca ocorram simultaneamente, prevenindo *race conditions*.
- **Goroutine de Física no Atuador** — A função `simularFisicaDoAmbiente()` roda em uma goroutine dedicada, alterando os arquivos `.txt` a cada segundo de forma independente do servidor TCP que recebe comandos.
- **`sync.Mutex` no Atuador** — O mapa `estadoCompressores` é acessado tanto pela goroutine de física quanto pelo handler de comandos TCP, protegido por mutex para garantir consistência.

---

## 5. Estrutura de Arquivos

```
rota-das-coisas/
│
├── docker-compose.yml          # Orquestração completa de todos os serviços
│
├── shared/                     # Tipos compartilhados (pacote Go)
│   ├── mensagem.go             # Structs Leitura e Mensagem
│   ├── comando.go              # Struct Comando e constantes de TipoComando
│   └── equipamento.go          # Struct Equipamento e funções de simulação
│
├── integration/
│   ├── main.go                 # Integrador: servidor UDP + API HTTP (SSE + Comandos)
│   └── Dockerfile
│
├── device/
│   ├── sensor.go               # Sensor: lê arquivo físico e envia UDP
│   ├── atuador.go              # Atuador: recebe TCP e altera o ambiente físico
│   └── Dockerfile
│
├── cliente/
│   └── index.html              # Dashboard Web (Chart.js + SSE + controles)
│
└── fisica/
    ├── ambiente_1.txt           # Estado térmico atual de cada equipamento
    ├── ambiente_2.txt
    └── ...
```

---

## 6. Como Executar

### Pré-requisitos

- [Docker](https://www.docker.com/get-started) instalado
- [Docker Compose](https://docs.docker.com/compose/) instalado

### Subindo o sistema completo

```bash
# Clone o repositório
git clone <url-do-repositório>
cd rota-das-coisas

# Build e inicialização de todos os containers
docker-compose up --build
```

Aguarde todos os serviços subirem. Você verá logs como:
```
integrador-1  | === Iniciando Serviço de Integração ===
integrador-1  | [UDP] Ouvindo telemetria na porta :9090...
integrador-1  | [HTTP] API (Comandos) e Streaming (Gráficos) operantes na porta :8080
atuador-1     | 🦾 Atuador 1 pronto e controlando a física na porta 8081...
sensor1-1     | 📡 Sensor 1 (Geladeira_1) ligado...
```

### Acessando o Dashboard

Abra o navegador em:

```
http://localhost
```

> **Acesso na rede local:** qualquer dispositivo na mesma rede pode acessar pelo IP da máquina que está rodando o Docker (ex: `http://192.168.1.50`). O frontend detecta automaticamente o endereço do servidor via `window.location.hostname`.

### Encerrando

```bash
docker-compose down
```

---

## 7. Equipamentos Simulados

| ID | Nome | Tipo | Temp. Mín | Temp. Máx |
|---|---|---|---|---|
| 1 | Geladeira_1 | geladeira | 2.0 °C | 4.5 °C |
| 2 | Geladeira_2 | geladeira | 1.0 °C | 3.0 °C |
| 3 | Geladeira_3 | geladeira | 3.0 °C | 5.0 °C |
| 4 | Freezer_4 | freezer | -18.0 °C | -15.0 °C |
| 5 | Freezer_5 | freezer | -20.0 °C | -10.0 °C |

A **física do ambiente** é simulada por arquivos de texto (`fisica/ambiente_N.txt`). O Atuador altera esses arquivos diretamente, e os Sensores os leem para gerar a telemetria — desacoplando completamente a simulação física da lógica de rede.

**Comportamento automático do Integrador:**
- Temperatura acima do limite máximo → envia `DiminuirTemperatura` ao Atuador (liga compressor)
- Temperatura abaixo do limite mínimo → envia `AumentarTemperatura` ao Atuador (desliga compressor)

---

## 8. Interface do Cliente

O Dashboard exibe um card por equipamento com:

- **Gráfico de linha** em tempo real com as últimas 15 leituras
- **Temperatura atual** em destaque numérico
- **Badge de status** dinâmica:
  - ✅ `Sistema Estável` — temperatura dentro dos limites
  - ⚠️ `Resfriando (Limite Máx)` — compressor ligado automaticamente
  - ⚠️ `Aquecendo (Limite Mín)` — compressor desligado automaticamente
- **Botões de controle manual:**
  - ❄️ **Resfriar** — força o Atuador a ligar o compressor
  - 🔥 **Aquecer** — força o Atuador a desligar o compressor

O streaming via SSE mantém **uma única conexão HTTP persistente** com o Integrador, recebendo atualizações a cada 2 segundos sem nenhum refresh ou polling da página.

---

## Portas Utilizadas

| Serviço | Porta | Protocolo |
|---|---|---|
| Integrador (telemetria) | 9090 | UDP |
| Integrador (API/Stream) | 8080 | TCP/HTTP |
| Atuador | 8081 | TCP |
| Cliente (Nginx) | 80 | HTTP |

---

*Projeto desenvolvido por Emanuel — UEFS, 2026.*
