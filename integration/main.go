package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
	"rota-das-coisas/shared" 
)


// Esse mapa guarda a última leitura de cada sensor pelo seu ID.
var mapaSensores = make(map[int]shared.Leitura)


// precisamos trancar o mapa na hora de escrever para o Go não dar crash.
var mu sync.Mutex

func main() {
	fmt.Println("=== Iniciando Serviço de Integração ===")

	go iniciarServidorUDP(":9090")

	iniciarServidorTCP(":8080")
}


func iniciarServidorUDP(porta string) {
	addr, err := net.ResolveUDPAddr("udp", porta)
	if err != nil {
		fmt.Println("Erro ao resolver UDP:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Erro ao iniciar UDP:", err)
		return
	}
	defer conn.Close()

	fmt.Printf("[UDP] Ouvindo telemetria dos sensores na porta %s...\n", porta)

	buffer := make([]byte, 1024)

	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Erro ao ler UDP:", err)
			continue
		}

		// Copia os bytes recebidos para não serem sobrescritos no próximo loop
		payload := make([]byte, n)
		copy(payload, buffer[:n])

		go processarTelemetria(payload, addr)
	}
}

func processarTelemetria(payload []byte, addr *net.UDPAddr) {
	var msg shared.Mensagem


	err := json.Unmarshal(payload, &msg)
	if err != nil {
		fmt.Printf("Erro ao decodificar JSON de %s: %v\n", addr.String(), err)
		return
	}

	
	// Recriando os bytes apenas do payload para converter limpo para Leitura
	payloadBytes, _ := json.Marshal(msg.Payload)
	var leitura shared.Leitura
	json.Unmarshal(payloadBytes, &leitura)

	// fechando o mapa para atualizar 
	mu.Lock()
	mapaSensores[leitura.EquipamentoID] = leitura
	mu.Unlock() 

	fmt.Printf("[INTEGRADOR] Recebido do Sensor %d | Temp: %.1f°C | Status da Memória Atualizado\n", 
		leitura.EquipamentoID, leitura.Temperatura)

	if leitura.Temperatura > leitura.TempMax {
		
		fmt.Printf("[ALERTA] %s %d muito quente! Acionando resfriamento...\n", leitura.TipoEquipamento, leitura.EquipamentoID)
		enviarComandoParaAtuador("localhost:8081", leitura.EquipamentoID, shared.DiminuirTemperatura)
		
	} else if leitura.Temperatura < leitura.TempMin {
		
		fmt.Printf("[ALERTA] %s %d muito frio! Acionando aquecimento...\n", leitura.TipoEquipamento, leitura.EquipamentoID)
		enviarComandoParaAtuador("localhost:8081", leitura.EquipamentoID, shared.AumentarTemperatura)
		
	}
}


//servidor TCP para se comunicar com o atuador e o cliente
func iniciarServidorTCP(porta string) {
	listener, err := net.Listen("tcp", porta)
	if err != nil {
		fmt.Println("Erro ao iniciar TCP:", err)
		return
	}
	defer listener.Close()

	fmt.Printf("[TCP] Ouvindo conexões de clientes e comandos na porta %s...\n", porta)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão TCP:", err)
			continue
		}

		go func(c net.Conn) {
			fmt.Printf("[TCP] Novo cliente conectado: %s\n", c.RemoteAddr().String())
			c.Close()
		}(conn)
	}

	
}

func enviarComandoParaAtuador(enderecoTCP string, id int, acao shared.TipoComando) {
	conn, err := net.Dial("tcp", enderecoTCP)
	if err != nil {
		fmt.Printf("[ERRO] Integrador não conseguiu alcançar o Atuador: %v\n", err)
		return
	}
	defer conn.Close()

	comando := shared.Comando{
		EquipamentoID: id,
		Tipo:          acao, 
		Timestamp:     time.Now(),
	}

	json.NewEncoder(conn).Encode(comando)
	fmt.Printf("[ALERTA CRÍTICO] Comando '%s' enviado com sucesso para o Equipamento %d!\n", acao, id)
}