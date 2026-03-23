package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
	"net"
)
	
func main(){

idSensor := os.Getenv("ID_SENSOR")
tipoSensor := os.Getenv("TIPO_SENSOR")

tempMinStr := os.Getenv("TEMP_MIN")
tempMin, err := strconv.ParseFloat(tempMinStr, 64)
	if err != nil {
		println("erro ao converter TEMP_MIN:", err)
		return
	}

tempMaxStr := os.Getenv("TEMP_MAX")
tempMax, err := strconv.ParseFloat(tempMaxStr, 64)
	if err != nil {
		println("erro ao converter TEM_MAX:", err)
		return
	}

nome := os.Getenv("NOME")



ligado := true

conn, err := net.Dial("udp", "localhost:9090")
		if err != nil {
			fmt.Printf("Sensor %s erro ao conectar: %v\n", idSensor, err)
			return
		}
		defer conn.Close()

		fmt.Printf("Sensor %s ligado !\n", idSensor)


for{
	if !ligado {
		break
		
	}

	temperatura := tempMin + rand.Float64()*(tempMax-tempMin)
	mensagem := fmt.Sprintf(
    `{"tipo":"telemetria", "nome":"%s", "sensor_id":"%s","tipo_equipamento":"%s","temperatura":%.1f, "timestamp":"%s"}`, nome,
    idSensor, tipoSensor, temperatura, time.Now().Format(time.RFC3339))
	conn.Write([]byte(mensagem))
	fmt.Printf("Sensor %s -> %.1f°C enviado\n", idSensor, temperatura)

	time.Sleep(1 * time.Second)
}

}
