package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"rota-das-coisas/shared"
	"strconv"
	"sync"
	"time"
)

// Estrutura para manter o estado interno de cada geladeira
type EstadoGeladeira struct {
	Temperatura float64
	MotorLigado bool
	Modo        shared.TipoComando // resfriar ou aquecer
}

var (
	estados = make(map[int]*EstadoGeladeira)
	mutex   sync.Mutex
)

func main() {
	porta := os.Getenv("PORTA_ATUADOR")
	if porta == "" {
		porta = "8081"
	}

	totalStr := os.Getenv("TOTAL_EQUIPAMENTOS")
	total, _ := strconv.Atoi(totalStr)
	if total == 0 {
		total = 5
	}

	// Inicializa a física para cada equipamento
	for i := 1; i <= total; i++ {
		estados[i] = &EstadoGeladeira{
			Temperatura: 25.0, // Temperatura inicial ambiente
			MotorLigado: false,
		}
		// Garante que o diretório física existe
		os.MkdirAll("fisica", 0755)
	}

	// Loop da Física: Roda em background atualizando os arquivos .txt
	go loopFisica(total)

	// Servidor TCP: Escuta comandos do Integrador
	ln, err := net.Listen("tcp", ":"+porta)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao iniciar Atuador: %v\n", err)
		return
	}
	fmt.Printf("[ATUADOR] Ouvindo comandos na porta %s...\n", porta)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go lidarComComando(conn)
	}
}

func lidarComComando(conn net.Conn) {
	defer conn.Close()
	var cmd shared.Comando
	if err := json.NewDecoder(conn).Decode(&cmd); err != nil {
		return
	}

	mutex.Lock()
	if est, ok := estados[cmd.EquipamentoID]; ok {
		est.MotorLigado = true
		est.Modo = cmd.Tipo
		fmt.Printf(">>> ATUADOR: Equip %d agora está em modo %s\n", cmd.EquipamentoID, cmd.Tipo)
	}
	mutex.Unlock()
}

func loopFisica(total int) {
	for {
		mutex.Lock()
		for id, est := range estados {
			// Lógica da Física
			var mudanca float64

			if est.MotorLigado {
				// Se o motor está ligado, a mudança é forte
				if est.Modo == shared.DiminuirTemperatura {
					mudanca = -1.5 // Resfriamento rápido
				} else {
					mudanca = 1.5 // Aquecimento rápido
				}
				// Chance de desligar o motor se chegar perto de um valor neutro
				// (Opcional, o integrador deve controlar isso)
			} else {
				// Calor ambiente (Inércia térmica: tende a subir para 30°C lentamente)
				if est.Temperatura < 30.0 {
					mudanca = 0.1
				}
			}

			est.Temperatura += mudanca

			// Escreve no arquivo .txt para os sensores lerem
			filename := filepath.Join("fisica", fmt.Sprintf("ambiente_%d.txt", id))
			os.WriteFile(filename, []byte(fmt.Sprintf("%.2f", est.Temperatura)), 0644)
		}
		mutex.Unlock()

		time.Sleep(1 * time.Second) // Ciclo da física
	}
}
