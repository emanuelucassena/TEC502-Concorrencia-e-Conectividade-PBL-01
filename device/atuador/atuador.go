package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"math/rand"
	"rota-das-coisas/shared"
)

var estadoCompressores = make(map[int]bool)

// Goroutine que altera o arquivo simulando o mundo real
func simularFisicaDoAmbiente() {
	os.MkdirAll("fisica", os.ModePerm)
	
	for {
		// Varre todos os equipamentos que o atuador conhece e simula a física neles
		for id, ligado := range estadoCompressores {
			nomeArquivo := fmt.Sprintf("fisica/ambiente_%d.txt", id)
			data, err := os.ReadFile(nomeArquivo)
			if err != nil { continue }

			tempAtual, errParse := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
			if errParse != nil { continue }

			
			chance := rand.Float64()
			if ligado {
				if chance < 0.60 { tempAtual -= 0.3 } else { tempAtual -= 0.1 }
			} else {
				if chance < 0.60 { tempAtual += 0.3 } else { tempAtual += 0.1 }
			}

			
			os.WriteFile(nomeArquivo, []byte(fmt.Sprintf("%.2f", tempAtual)), 0644)
		}
		time.Sleep(1 * time.Second)
	}
}

func handleComando(conn net.Conn, idAtuador string) {
	defer conn.Close()
	var cmd shared.Comando
	
	if err := json.NewDecoder(conn).Decode(&cmd); err != nil {
		fmt.Printf("Erro ao decodificar comando: %v\n", err)
		return
	}

	fmt.Printf("\n>>> ATUADOR %s RECEBEU COMANDO: %s para Equip %d\n", idAtuador, cmd.Tipo, cmd.EquipamentoID)

	// Executa a ação fisicamente (atualiza o mapa para a Goroutine processar)
	if cmd.Tipo == shared.DiminuirTemperatura {
		estadoCompressores[cmd.EquipamentoID] = true
		fmt.Println("[FÍSICA] Ligando motor para resfriar...")
	} else if cmd.Tipo == shared.AumentarTemperatura || cmd.Tipo == shared.DesligarEquipamento {
		estadoCompressores[cmd.EquipamentoID] = false
		fmt.Println("[FÍSICA] Desligando motor. Aquecimento natural...")
	}
}

func main() {
	idAtuador := os.Getenv("ID_ATUADOR")
	porta := os.Getenv("PORTA_ATUADOR")
	if porta == "" { porta = "8081" }

	
	totalEquipamentos, err := strconv.Atoi(os.Getenv("TOTAL_EQUIPAMENTOS"))
	if err != nil || totalEquipamentos == 0 {
		totalEquipamentos = 2 
	}

	for i := 1; i <= totalEquipamentos; i++ {
		estadoCompressores[i] = false
	}

	go simularFisicaDoAmbiente()

	listener, err := net.Listen("tcp", ":" + porta)
	if err != nil {
		fmt.Printf("Erro ao iniciar o Atuador %s: %v\n", idAtuador, err)
		return
	}
	defer listener.Close()

	fmt.Printf("Atuador %s pronto e controlando a física na porta %s...\n", idAtuador, porta)

	for {
		conn, err := listener.Accept()
		if err != nil { continue }
		go handleComando(conn, idAtuador)
	}
}