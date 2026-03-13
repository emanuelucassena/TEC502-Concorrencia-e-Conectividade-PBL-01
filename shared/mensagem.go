package shared

import "time"

type Leitura struct{
	EquipamentoID int `json:"equipamento_id"`
	Temperatura float64 `json:"temperatura"`
	Umidade float64 `json:"umidade"`
	Timestamp time.Time `json:"timestamp"`
}

type Mensagem struct {
	Tipo string `json:"tipo"`
	Payload interface{} `json:"payload"`
}

