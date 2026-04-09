package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
	"rota-das-coisas/shared"
)

var historico = make(map[int][]shared.Leitura)
var mutex sync.Mutex

func salvarHistorico() {
	mutex.Lock()
	dados, err := json.MarshalIndent(historico, "", " ")
	mutex.Unlock()

	if err == nil {
		os.WriteFile("historico.json", dados, 0644)
	}
}

func main() {
	fmt.Println("=== Iniciando Serviço de Integração ===")
	
	// Ligamos a API para o Dashboard Web
	go iniciarServidorHTTP(":8080")

	// Ligamos o servidor UDP para ouvir os sensores
	iniciarServidorUDP(":9090")
}

func iniciarServidorHTTP(porta string) {
	// =========================================================
	// ROTA 1: O "Tubo" de streaming (A que você acabou de mandar)
	// =========================================================
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming não suportado pelo navegador", http.StatusInternalServerError)
			return
		}

		for {
			mutex.Lock()
			dados, _ := json.Marshal(historico)
			mutex.Unlock()

			fmt.Fprintf(w, "data: %s\n\n", dados)
			flusher.Flush() 
			time.Sleep(2 * time.Second)
		}
	})

	// =========================================================
	// ROTA 2: A central de comandos (A nova, blindada contra CORS)
	// =========================================================
	http.HandleFunc("/comando", func(w http.ResponseWriter, r *http.Request) {
		// 1. CARIMBANDO O PASSAPORTE (CORS) - Tem que ser a primeira coisa!
		w.Header().Set("Access-Control-Allow-Origin", "*") // Aceita cliques do localhost:3000 ou qualquer outro
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// 2. O LEÃO DE CHÁCARA: Responde à requisição fantasma do navegador
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 3. COMANDO REAL DO BOTÃO
		var req struct {
			EquipamentoID int    `json:"equipamento_id"`
			Acao          string `json:"acao"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { 
			http.Error(w, err.Error(), http.StatusBadRequest)
			return 
		}

		// Traduz o comando do Front para a Física
		tipoCmd := shared.AumentarTemperatura 
		if req.Acao == "resfriar" {
			tipoCmd = shared.DiminuirTemperatura 
		}

		// Envia a ordem para o Atuador operar o motor!
		hostAtuador := os.Getenv("HOST_ATUADOR")
		if hostAtuador == "" { hostAtuador = "atuador:8081" }
		enviarComandoParaAtuador(hostAtuador, req.EquipamentoID, tipoCmd)
		
		// Responde para o Frontend que deu tudo certo
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "sucesso"}`)
	})

	// =========================================================
	// INICIA O SERVIDOR COM AS DUAS ROTAS
	// =========================================================
	fmt.Printf("[HTTP] API (Comandos) e Streaming (Gráficos) operantes na porta %s\n", porta)
	http.ListenAndServe(porta, nil)
}

func iniciarServidorUDP(porta string) {
	addr, err := net.ResolveUDPAddr("udp", porta)
	if err != nil { return }
	conn, err := net.ListenUDP("udp", addr)
	if err != nil { return }
	defer conn.Close()

	fmt.Printf("[UDP] Ouvindo telemetria na porta %s...\n", porta)
	buffer := make([]byte, 1024)

	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil { continue }
		payload := make([]byte, n)
		copy(payload, buffer[:n])
		go processarTelemetria(payload, addr)
	}
}

func processarTelemetria(payload []byte, addr *net.UDPAddr) {
	var msg shared.Mensagem
	if err := json.Unmarshal(payload, &msg); err != nil { return }

	payloadBytes, _ := json.Marshal(msg.Payload)
	var leitura shared.Leitura
	json.Unmarshal(payloadBytes, &leitura)

	mutex.Lock()
	historico[leitura.EquipamentoID] = append(historico[leitura.EquipamentoID], leitura)
	if len(historico[leitura.EquipamentoID]) > 30 {
		historico[leitura.EquipamentoID] = historico[leitura.EquipamentoID][1:]
	}
	mutex.Unlock()
	
	salvarHistorico()

	fmt.Printf("[INTEGRADOR] Recebido do Sensor %d | Temp: %.1f°C \n", leitura.EquipamentoID, leitura.Temperatura)

	hostAtuador := os.Getenv("HOST_ATUADOR")
	if hostAtuador == "" { hostAtuador = "localhost:8081" }

	if leitura.Temperatura > leitura.TempMax {
		fmt.Printf("[ALERTA] geladeira %d muito quente! Acionando resfriamento...\n", leitura.EquipamentoID)
		enviarComandoParaAtuador(hostAtuador, leitura.EquipamentoID, shared.DiminuirTemperatura)
	} else if leitura.Temperatura < leitura.TempMin {
		fmt.Printf("[ALERTA] geladeira %d muito frio! Acionando aquecimento...\n", leitura.EquipamentoID)
		enviarComandoParaAtuador(hostAtuador, leitura.EquipamentoID, shared.AumentarTemperatura)
	}
}

func enviarComandoParaAtuador(enderecoTCP string, id int, acao shared.TipoComando) {
	conn, err := net.Dial("tcp", enderecoTCP)
	if err != nil { return }
	defer conn.Close()

	comando := shared.Comando{ EquipamentoID: id, Tipo: acao, Timestamp: time.Now() }
	json.NewEncoder(conn).Encode(comando)
	fmt.Printf("[ALERTA CRÍTICO] Comando '%s' enviado com sucesso para o Equipamento %d!\n", acao, id)
}