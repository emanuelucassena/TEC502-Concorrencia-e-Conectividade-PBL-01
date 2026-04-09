# 🌐 Rota das Coisas 

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![HTML5](https://img.shields.io/badge/HTML5-E34F26?style=for-the-badge&logo=html5&logoColor=white)
![JavaScript](https://img.shields.io/badge/JavaScript-323330?style=for-the-badge&logo=javascript&logoColor=F7DF1E)

Projeto desenvolvido para a disciplina **TEC502 — Concorrência e Conectividade** (Engenharia de Computação — UEFS). Simula uma malha industrial de refrigeração com telemetria contínua, controle remoto de atuadores e processamento concorrente.

---

## 📋 Sumário

1. [Arquitetura do Sistema](#1-arquitetura-do-sistema)
2. [Comunicação e Protocolos](#2-comunicação-e-protocolos)
3. [Encapsulamento de Dados](#3-encapsulamento-de-dados)
4. [Concorrência](#4-concorrência)
5. [Estrutura de Arquivos](#5-estrutura-de-arquivos)
6. [Como Executar no Linux](#6-como-executar-no-linux)
7. [Equipamentos Simulados](#7-equipamentos-simulados)
8. [Interface do Cliente](#8-interface-do-cliente)

---

## 1. Arquitetura do Sistema

O sistema utiliza o padrão **Broker** (Serviço de Integração) para eliminar o alto acoplamento ponto-a-ponto entre produtores de dados (sensores) e consumidores (clientes). Nenhum sensor conhece o cliente, e nenhum cliente conhece os sensores.

```
┌─────────────┐  UDP :9090   ┌──────────────────┐  TCP :8081   ┌──────────────┐
│   SENSORES  │ ───────────▶ │   INTEGRADOR     │ ───────────▶ │   ATUADOR    │
│ (sensor_bin)│              │  (Broker / Go)   │              │ (atuador_bin)│
└─────────────┘              └────────┬─────────┘              └──────┬───────┘
                                      │ SSE :8080                     │
                                      ▼                               │ (lê/escreve)
                             ┌─────────────────┐               ┌──────┴──────┐
                             │  CLIENTE WEB    │               │ fisica/*.txt│
                             │ (Nginx :80)     │               └─────────────┘
                             └─────────────────┘
         HTTP POST :8080 ───────────────────────────────────────────▶ (via Integrador)
```

**Camadas:**
- **Dispositivos** — Sensores leem arquivos de estado físico (`fisica/ambiente_N.txt`) e enviam telemetria. O Atuador modifica esses arquivos simulando um compressor.
- **Integração** — O Broker recebe telemetria UDP, aplica regras de limite térmico automaticamente e expõe API HTTP com streaming SSE.
- **Apresentação** — Dashboard web servido por Nginx, com gráficos em tempo real e controle manual por botões.

---

## 2. Comunicação e Protocolos

| Fluxo | Protocolo | Porta | Justificativa |
|---|---|---|---|
| Sensor → Integrador | **UDP** | 9090 | Telemetria volumosa e contínua. Baixa latência, sem overhead de handshake. A perda pontual de um pacote não compromete o histórico. |
| Integrador → Atuador | **TCP** | 8081 | Comandos críticos de controle físico. Entrega garantida obrigatória. |
| Cliente → Integrador | **HTTP POST** | 8080 | Comando manual do operador com resposta imediata de sucesso/erro. |
| Integrador → Cliente | **SSE (HTTP)** | 8080 | Canal unidirecional persistente. O servidor empurra atualizações a cada 2s, eliminando o overhead de polling do navegador. |

---

## 3. Encapsulamento de Dados

Todos os dados trafegam encapsulados em **JSON**, serializados pelas structs do pacote `shared`.

**Telemetria** (Sensor → Integrador via UDP):
```json
{
  "tipo": "telemetria",
  "payload": {
    "equipamento_id": 1,
    "tipo_equipamento": "geladeira",
    "temperatura": 4.10,
    "temp_min": 2.0,
    "temp_max": 4.5,
    "timestamp": "2026-04-09T14:00:00Z"
  }
}
```

**Comando** (Integrador → Atuador via TCP):
```json
{
  "equipamento_id": 1,
  "tipo": "diminuir_temperatura",
  "timestamp": "2026-04-09T14:00:01Z"
}
```

**Tipos de comando disponíveis:**

| Valor | Efeito |
|---|---|
| `diminuir_temperatura` | Liga o compressor (resfriamento ativo) |
| `aumentar_temperatura` | Desliga o compressor (aquecimento natural) |
| `desligar_equipamento` | Desliga o compressor |

---

## 4. Concorrência

| Mecanismo | Onde | Função |
|---|---|---|
| **Goroutines** | Integrador | Cada pacote UDP recebido é processado em uma goroutine independente — multiplexação massiva de I/O sem bloqueio. |
| **`sync.Mutex`** | Integrador | Protege o mapa de histórico compartilhado entre a goroutine UDP e o handler HTTP de streaming, prevenindo race conditions. |
| **Goroutine de física** | Atuador | `simularFisicaDoAmbiente()` roda em background, alterando os arquivos `.txt` a cada segundo independente do servidor TCP. |
| **`sync.Mutex`** | Atuador | Protege o mapa `estadoCompressores` acessado tanto pela goroutine de física quanto pelo handler de comandos. |

---

## 5. Estrutura de Arquivos

```
rota-das-coisas/
│
├── docker-compose.yml          # Orquestração completa
├── go.mod                      # Módulo Go (rota-das-coisas)
│
├── shared/                     # Tipos compartilhados entre todos os serviços
│   ├── mensagem.go             # Structs Leitura e Mensagem
│   ├── comando.go              # Struct Comando e constantes TipoComando
│   └── equipamento.go          # Struct Equipamento
│
├── integration/
│   ├── main.go                 # Broker: UDP + HTTP (SSE + Comandos)
│   └── Dockerfile
│
├── device/
│   ├── sensor/
│   │   └── sensor.go           # Lê arquivo físico e envia UDP
│   ├── atuador/
│   │   └── atuador.go          # Recebe TCP e altera arquivos físicos
│   └── Dockerfile
│
├── client/
│   └── index.html              # Dashboard Web (Chart.js + SSE + controles)
│
└── fisica/
    ├── ambiente_1.txt           # Estado térmico atual de cada equipamento
    └── ...
```

---

## 6. Como Executar no Linux

### Pré-requisitos

```bash
# Verificar se Docker está instalado
docker --version

# Verificar se Docker Compose está instalado
docker compose version
```

Caso não esteja instalado:
```bash
sudo apt update && sudo apt install -y docker.io docker-compose-plugin
sudo systemctl start docker
```

### Subir o sistema

```bash
# Acesse a pasta do projeto
cd rota-das-coisas

# Build e inicialização de todos os containers
sudo docker compose up --build
```

Aguarde todos os serviços subirem. O sistema está pronto quando aparecer:
```
integrador-1  | [UDP] Ouvindo telemetria na porta :9090...
integrador-1  | [HTTP] API (Comandos) e Streaming (Gráficos) operantes na porta :8080
atuador-1     | 🦾 Atuador 1 pronto e controlando a física na porta 8081...
sensor1-1     | 📡 Sensor 1 (Geladeira_1) ligado...
```

### Acessar o Dashboard

**Na mesma máquina:**
```
http://localhost
```

**Em outra máquina na mesma rede** (celular, outro notebook etc.):
```bash
# Descubra o IP da máquina que está rodando o Docker
hostname -I | awk '{print $1}'
```
Acesse `http://[IP-DA-MAQUINA]` — o frontend detecta automaticamente o endereço do servidor via `window.location.hostname`, sem necessidade de alterar nenhum arquivo.

### Encerrar

```bash
sudo docker compose down
```

### Problemas comuns no Linux

| Problema | Solução |
|---|---|
| `permission denied` ao rodar Docker | `sudo usermod -aG docker $USER` e reinicie a sessão |
| Porta 80 já em uso | `sudo lsof -i :80` para identificar e encerrar o processo conflitante |
| Containers não se comunicam | Garanta que todos os serviços sobem com o mesmo `docker compose up` (mesma rede `rede_iot`) |
| Build falha sem internet | A compilação é feita localmente a partir do `go.mod` presente na raiz — não requer acesso externo |

---

## 7. Equipamentos Simulados

| ID | Nome | Tipo | Temp. Mín | Temp. Máx |
|---|---|---|---|---|
| 1 | Geladeira_1 | geladeira | 2.0 °C | 4.5 °C |
| 2 | Geladeira_2 | geladeira | 1.0 °C | 3.0 °C |
| 3 | Geladeira_3 | geladeira | 3.0 °C | 5.0 °C |
| 4 | Freezer_4 | freezer | -18.0 °C | -15.0 °C |
| 5 | Freezer_5 | freezer | -20.0 °C | -10.0 °C |

**Comportamento automático do Integrador:**
- Temperatura acima do limite máximo → `diminuir_temperatura` → compressor ligado
- Temperatura abaixo do limite mínimo → `aumentar_temperatura` → compressor desligado

---

## 8. Interface do Cliente

O Dashboard exibe um card por equipamento contendo:

- **Gráfico de linha** em tempo real com as últimas 15 leituras
- **Temperatura atual** em destaque numérico
- **Badge de status** dinâmica:
  - ✅ `Sistema Estável` — temperatura dentro dos limites
  - ⚠️ `Resfriando (Limite Máx)` — compressor acionado automaticamente
  - ⚠️ `Aquecendo (Limite Mín)` — compressor desligado automaticamente
- **Controle manual:**
  - ❄️ **Resfriar** — envia comando para ligar o compressor
  - 🔥 **Aquecer** — envia comando para desligar o compressor

---

## Portas Utilizadas

| Serviço | Porta | Protocolo |
|---|---|---|
| Integrador — Telemetria | 9090 | UDP |
| Integrador — API/Stream | 8080 | TCP/HTTP |
| Atuador — Comandos | 8081 | TCP |
| Cliente — Dashboard | 80 | HTTP |

---

*Emanuel — Engenharia de Computação, UEFS (2026)*