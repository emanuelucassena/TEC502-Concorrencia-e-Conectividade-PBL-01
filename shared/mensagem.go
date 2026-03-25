package shared

import "time"

type Leitura struct{
	EquipamentoID   int             `json:"equipamento_id"`
	TipoEquipamento TipoEquipamento `json:"tipo_equipamento"` 
	Temperatura     float64         `json:"temperatura"`
	TempMin         float64         `json:"temp_min"`         
	TempMax         float64         `json:"temp_max"`         
	Umidade         float64         `json:"umidade"`
	Timestamp       time.Time       `json:"timestamp"`
}

type Mensagem struct {
	Tipo    string      `json:"tipo"`
	Payload interface{} `json:"payload"`
}

