package shared

import (
	"math/rand"
)

// Define se é geladeira ou freezer, que são os dois equipamentos para serem simulados
type TipoEquipamento string

const(
	Geladeira TipoEquipamento = "geladeira"
	Freezer TipoEquipamento = "freezer"

)

type StatusEquipamento string

const(
	Normal StatusEquipamento = "normal"
	Alerta StatusEquipamento = "alerta"
	Critico StatusEquipamento = "critico"
	Desligado StatusEquipamento = "desligado"
)

type Equipamento struct{
	ID int	`json:"id"`
	Nome string	`json:"nome"`
	Tipo TipoEquipamento	 `json:"tipo"`
	TempAtual float64	`json:"temp_atual"`
	TempMin float64		`json:"temp_min"`
	TempMax float64		`json:"temp_max"`
	Status StatusEquipamento	`json:"status"`
	Ligado bool		`json:"ligado"`
}


func NovoEquipamento(id int, nome string, tipo TipoEquipamento, tempMin float64, tempMax float64) Equipamento  {
	equip := Equipamento{
		ID: id,
		Nome: nome,
		Tipo: tipo,
		TempAtual: tempMin + (tempMax-tempMin)/2,
		TempMin: tempMin,
		TempMax: tempMax,
		Status: Normal,
		Ligado: true,
	}
	return equip
}

func AtualizarStatus(equip *Equipamento){
	if !equip.Ligado {
		equip.Status = Desligado
	} else if equip.TempAtual > equip.TempMax || equip.TempAtual < equip.TempMin {
		equip.Status = Critico
	} else if equip.TempAtual >= equip.TempMax * 0.9 {
		equip.Status = Alerta
	} else {
		equip.Status = Normal
	}   
}

func SimularTemperatura(equip *Equipamento){
	if equip.Ligado {
		chance := rand.Float64()
		if chance < 0.70 {
			equip.TempAtual += 0.0
		} else if chance < 0.90 {
			equip.TempAtual += 0.1
		} else {
			equip.TempAtual -= 0.1
		}
	} else {
		// Se desligado, a temperatura sobe
		equip.TempAtual += 0.1
	}
}