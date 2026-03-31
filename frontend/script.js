function abrirDetalhes(id, nome) {
    
    document.getElementById('visao-geral').classList.add('hidden');
    
    
    document.getElementById('visao-detalhes').classList.remove('hidden');
    
    document.getElementById('detalhe-nome').innerText = nome;
    document.getElementById('detalhe-id').innerText = "ID do Equipamento: " + id;
    
    console.log("Iniciando monitoramento detalhado do Sensor ID:", id);
}


function voltarVisaoGeral() {
    
    document.getElementById('visao-detalhes').classList.add('hidden');
    
    // Mostra a div da Visão Geral novamente
    document.getElementById('visao-geral').classList.remove('hidden');
    
    console.log("Voltando para a visão de todos os sensores.");
}