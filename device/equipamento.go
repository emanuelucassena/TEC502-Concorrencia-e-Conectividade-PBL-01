package main
import(

 "rota-das-coisas/shared"
 "math/rand"
)


func NovoEquipamento(id int, nome string, tipo shared.TipoEquipamento, tempMin float64, tempMax float64) shared.Equipamento  {
	equip := shared.Equipamento{
	ID: id,
	Nome: nome,
	Tipo: tipo,
	TempAtual: tempMin + (tempMax-tempMin)/2,
	TempMin: tempMin,
	TempMax: tempMax,
	Status: shared.Normal,
	Ligado: true,
}
	return equip
}

func AtualizarStatus(equip *shared.Equipamento){
	if !equip.Ligado {
		equip.Status = shared.Desligado
	}else if equip.TempAtual > equip.TempMax || equip.TempAtual < equip.TempMin {
    equip.Status = shared.Critico
	}else if equip.TempAtual >= equip.TempMax * 0.9 {
    equip.Status = shared.Alerta
	}else{
	equip.Status = shared.Normal
} 	
}

func SimularTemperatura(equip *shared.Equipamento){
	
	if equip.Ligado {
		chance := rand.Float64()
		if chance < 0.70 {
			equip.TempAtual += 0.0
		}else if chance < 0.90{
			equip.TempAtual += 0.1
		}else{
			equip.TempAtual -= 0.1
		}
	}else{
		equip.TempAtual += 0.1
	}
	
}