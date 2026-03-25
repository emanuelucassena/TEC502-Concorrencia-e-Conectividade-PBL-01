package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"rota-das-coisas/shared" 
)

func main() {
	idSensorStr := os.Getenv("ID_SENSOR")
	tipoSensor := os.Getenv("TIPO_SENSOR")
	nome := os.Getenv("NOME")


	idSensor, err := strconv.Atoi(idSensorStr)
	if err != nil {
		fmt.Println("erro ao converter ID_SENSOR:", err)
		return
	}

	
	tempMinStr := os.Getenv("TEMP_MIN")
	tempMin, err := strconv.ParseFloat(tempMinStr, 64)
	if err != nil {
		fmt.Println("erro ao converter TEMP_MIN:", err)
		return
	}

	tempMaxStr := os.Getenv("TEMP_MAX")
	tempMax, err := strconv.ParseFloat(tempMaxStr, 64)
	if err != nil {
		fmt.Println("erro ao converter TEMP_MAX:", err)
		return
	}

	
	hostIntegrador := os.Getenv("HOST_INTEGRADOR")
	if hostIntegrador == "" {
		hostIntegrador = "localhost:9090"
	}

	ligado := true

	
	conn, err := net.Dial("udp", hostIntegrador)
	if err != nil {
		fmt.Printf("Sensor %d erro ao conectar: %v\n", idSensor, err)
		return
	}
	defer conn.Close()

	fmt.Printf("Sensor %d (%s do tipo %s) ligado e enviando para %s!\n", idSensor, nome, tipoSensor, hostIntegrador)

	equip := shared.NovoEquipamento(idSensor, nome, shared.TipoEquipamento(tipoSensor), tempMin, tempMax)

	for {
		if !ligado {
			break
		}

		
		shared.SimularTemperatura(&equip)

		
		leitura := shared.Leitura{
			EquipamentoID: idSensor,
			TipoEquipamento: shared.TipoEquipamento(tipoSensor),
			Temperatura:   equip.TempAtual,
			TempMin: tempMin,
			TempMax: tempMax,
			Umidade:       50.0, 
			Timestamp:     time.Now(),
		}

		
		mensagem := shared.Mensagem{
			Tipo:    "telemetria",
			Payload: leitura,
		}

		
		jsonBytes, err := json.Marshal(mensagem)
		if err != nil {
			fmt.Println("Erro ao converter para JSON:", err)
			continue
		}

		
		conn.Write(jsonBytes)
		fmt.Printf("Sensor %d -> %.1f°C enviado\n", idSensor, equip.TempAtual)

		time.Sleep(1 * time.Second)
	}
}