package main

import(
	"encoding/json"
	"fmt"
	"net"
	"os"
	"rota-das-coisas/shared"
)


func handleComando(conn net.Conn, id string) {
	defer conn.Close()
	var cmd shared.Comando
	
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&cmd); err != nil {
		fmt.Printf("Erro ao decodificar comando para Atuador %s: %v\n", id, err)
		return
	}

	
	fmt.Printf(">>> ATUADOR %s RECEBEU COMANDO: %s às %v\n", id, cmd.Tipo, cmd.Timestamp)

	portaDoSensor := fmt.Sprintf("localhost:%d", 7000 + cmd.EquipamentoID)

	connSensor, err := net.Dial("tcp", portaDoSensor)
		if err != nil {
			fmt.Println("[ERRO] Atuador tentou ligar, mas o Sensor não atendeu na porta 7001:", err)
		} else {
			
			
			
			json.NewEncoder(connSensor).Encode(cmd)
			connSensor.Close()
			
			fmt.Println(">>> ATUADOR REPASSOU A ORDEM FISICA PARA O SENSOR COM SUCESSO!")
		}
	
	if cmd.Tipo == shared.DesligarEquipamento {
		fmt.Printf("!!! AVISO: Equipamento %d sendo DESLIGADO !!!\n", cmd.EquipamentoID)
	}
}

func main(){
	
	idAtuador := os.Getenv("ID_ATUADOR")
	porta := os.Getenv("PORTA_ATUADOR")


	listener, err := net.Listen("tcp", ":" + porta)
	if err != nil {
		fmt.Printf("Erro ao iniciar o Atuador %s: %v\n", idAtuador, err)
		return
	}
	defer listener.Close()

	fmt.Printf("Atuador %s pronto e ouvindo na porta %s...\n", idAtuador, porta)

	for{
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}

		go handleComando(conn, idAtuador)
	}


	
}