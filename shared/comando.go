package shared

import "time"

type TipoComando string

const(
	LigarEquipamento TipoComando = "ligar_equipamento"
	DesligarEquipamento TipoComando = "desligar_equipamento"
	ResetarAlarme TipoComando = "resetar_alarme"
)

type Comando struct{
	EquipamentoID int `json:"equipamento_id"`
	Tipo TipoComando `json:"tipo"`
	Timestamp time.Time `json:"timestamp"`
}