FROM golang:1.20-alpine
WORKDIR /app

# 1. Copia a casa inteira para dentro do container
COPY . .

# 2. A MÁGICA: Apaga arquivos velhos, força a criação do módulo e mapeia as pastas
RUN rm -f go.mod go.sum
RUN go mod init rota-das-coisas
RUN go mod tidy

# 3. Compilação blindada
RUN go build -o atuador_bin ./device/atuador/atuador.go
RUN go build -o sensor_bin ./device/sensor/sensor.go

RUN chmod +x atuador_bin sensor_bin