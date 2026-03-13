package shared




// define se é geladeira ou freezer, que são os dois equipamentos para serem simulados
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