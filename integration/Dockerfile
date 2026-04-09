FROM golang:1.20-alpine
WORKDIR /app

# 1. Copia a casa inteira
COPY . .

# 2. A MÁGICA: Força o Go a reconhecer a pasta "shared"
RUN rm -f go.mod go.sum
RUN go mod init rota-das-coisas
RUN go mod tidy

# 3. Compilação blindada
RUN go build -o integrador_bin ./integration/main.go

RUN chmod +x integrador_bin