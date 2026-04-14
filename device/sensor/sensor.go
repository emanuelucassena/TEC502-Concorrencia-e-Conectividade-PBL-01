package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"rota-das-coisas/shared"
	"strconv"
	"strings"
	"time"
)

// Função que lê a "física" do arquivo de texto
func lerTemperaturaFisica(idSensor int, tempInicial float64) float64 {
	nomeArquivo := fmt.Sprintf("fisica/ambiente_%d.txt", idSensor)
	data, err := os.ReadFile(nomeArquivo)

	if err != nil {
		// Se o arquivo não existe, cria um com a temperatura média inicial
		os.MkdirAll("fisica", os.ModePerm)
		os.WriteFile(nomeArquivo, []byte(fmt.Sprintf("%.2f", tempInicial)), 0644)
		return tempInicial
	}

	tempStr := strings.TrimSpace(string(data))
	temp, err := strconv.ParseFloat(tempStr, 64)
	if err != nil {
		return tempInicial
	}
	return temp
}

func main() {
	idSensor, _ := strconv.Atoi(os.Getenv("ID_SENSOR"))
	nome := os.Getenv("NOME")
	tipoSensor := os.Getenv("TIPO_SENSOR")
	tempMin, _ := strconv.ParseFloat(os.Getenv("TEMP_MIN"), 64)
	tempMax, _ := strconv.ParseFloat(os.Getenv("TEMP_MAX"), 64)

	hostIntegrador := os.Getenv("HOST_INTEGRADOR")
	if hostIntegrador == "" {
		hostIntegrador = "172.16.103.2:9090"
	}

	conn, err := net.Dial("udp", hostIntegrador)
	if err != nil {
		fmt.Printf("Sensor %d erro ao conectar: %v\n", idSensor, err)
		return
	}
	defer conn.Close()

	fmt.Printf("📡 Sensor %d (%s) ligado. Lendo ambiente físico e enviando para %s!\n", idSensor, nome, hostIntegrador)

	tempInicial := tempMin + (tempMax-tempMin)/2

	for {
		tempAtual := lerTemperaturaFisica(idSensor, tempInicial)

		leitura := shared.Leitura{
			EquipamentoID:   idSensor,
			TipoEquipamento: shared.TipoEquipamento(tipoSensor),
			Temperatura:     tempAtual,
			TempMin:         tempMin,
			TempMax:         tempMax,
			Umidade:         50.0,
			Timestamp:       time.Now(),
		}

		mensagem := shared.Mensagem{
			Tipo:    "telemetria",
			Payload: leitura,
		}

		jsonBytes, _ := json.Marshal(mensagem)
		conn.Write(jsonBytes)

		fmt.Printf("Sensor %d -> Lendo da Física: %.1f°C\n", idSensor, tempAtual)
		time.Sleep(1 * time.Second)
	}
}
