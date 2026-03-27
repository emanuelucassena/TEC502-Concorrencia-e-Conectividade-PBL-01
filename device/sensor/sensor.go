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

func escutarAtuador(porta string, equip *shared.Equipamento){
	listener, err := net.Listen("tcp", ":"+ porta)
		if err != nil {
			fmt.Printf("[ERRO!] NÃO FOI POSSÍVEL ABRIR A PORTA %s: %v\n", porta, err)
			return
		}
		defer listener.Close()
		fmt.Printf("[INFO] Sensor ouvindo o Atuador na porta %s...\n", porta)

		for{
			conn, err := listener.Accept()
				if err != nil {
					continue
				}
			var cmd shared.Comando
			json.NewDecoder(conn).Decode(&cmd)
			conn.Close()

			fmt.Printf("\n [COMANDO] ATUADOR ORDENOU: %s\n", cmd.Tipo)

			if cmd.Tipo == "diminuir_temperatura"  {
				equip.Ligado = true
				fmt.Println("[STATUS] Compressor LIGADO. Esfriando...")

			}else if cmd.Tipo == "aumentar_temperatura" || cmd.Tipo == "resetar_alarme" {
			equip.Ligado = false
			fmt.Println("[STATUS] Compressor DESLIGADO.")
		}
		}
	}




func main() {
	idStr := os.Getenv("ID_SENSOR")
	idSensor, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("[ERRO FATAL] ID_SENSOR inválido ou não informado:", err)
		return
	}

	nome := os.Getenv("NOME")
	tipoSensor := os.Getenv("TIPO_SENSOR")

	tempMinStr := os.Getenv("TEMP_MIN")
	tempMin, err := strconv.ParseFloat(tempMinStr, 64)
	if err != nil {
		fmt.Println("[ERRO FATAL] TEMP_MIN inválido ou não informado:", err)
		return
	}

	tempMaxStr := os.Getenv("TEMP_MAX")
	tempMax, err := strconv.ParseFloat(tempMaxStr, 64)
	if err != nil {
		fmt.Println("[ERRO FATAL] TEMP_MAX inválido ou não informado:", err)
		return
	}

	equip := shared.NovoEquipamento(idSensor, nome, shared.TipoEquipamento(tipoSensor), tempMin, tempMax)
	equip.Ligado = false

	go escutarAtuador("7001", &equip)

	
	
	hostIntegrador := os.Getenv("HOST_INTEGRADOR")
	if hostIntegrador == "" {
		hostIntegrador = "localhost:9090"
	}


	
	conn, err := net.Dial("udp", hostIntegrador)
	if err != nil {
		fmt.Printf("Sensor %d erro ao conectar: %v\n", idSensor, err)
		return
	}
	defer conn.Close()

	fmt.Printf("Sensor %d (%s do tipo %s) ligado e enviando para %s!\n", idSensor, nome, tipoSensor, hostIntegrador)


	for {
		
		
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
		
		status := "DESLIGADO"
		if equip.Ligado {
			status = "LIGADO"
		}
		fmt.Printf("Sensor %d -> %.1f°C | Compressor: [%s]\n", idSensor, equip.TempAtual, status)
		

		time.Sleep(1 * time.Second)
	}


}