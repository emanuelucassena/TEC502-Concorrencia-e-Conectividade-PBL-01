package main

import (

	"fmt"
	"net"
	"time"
    "math/rand"
)

func sensor(id int) {
	time.Sleep(1 * time.Second)

    conn, err := net.Dial("udp", "localhost:9090")
    if err != nil{
        fmt.Printf("Sensor %d -> erro ao conectar: %v\n", id, err)
        return
    }
    defer conn.Close()

    fmt.Printf("Sensor %d ligado !\n", id)

    for{
        temperatura := 2.0 + rand.Float64()*6.0
        mensagem := fmt.Sprintf("sensor_id:%d temperatura:%.1f", id, temperatura)
        conn.Write([]byte(mensagem))
        fmt.Printf("Sensor %d -> %.1f°C enviado\n", id, temperatura)
        time.Sleep(1 * time.Second)
    }
}

    func servidor()  {
        conn, err := net.ListenPacket("udp", ":9090")
        if err != nil {
            fmt.Println("Erro ao iniciar o servidor:", err)
            return
        }
        defer conn.Close()

        fmt.Println("Servidor UDP rodando na porta 9090...")
        buffer := make([]byte, 1024)

        for{
            n, addr, err := conn.ReadFrom(buffer)
            if err != nil {
                fmt.Println("Erro ao ler os dados:", err)
                continue
            }
            fmt.Printf("Recebido de %s → %s\n", addr, string(buffer[:n]))
        }
    }

    func main(){
        
        go sensor(1)
        go sensor(2)
        go sensor(3)

        servidor()
    }



