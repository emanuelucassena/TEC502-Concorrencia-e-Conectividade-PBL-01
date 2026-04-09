FROM golang:1.20-alpine
WORKDIR /app
ENV GOPROXY=off
COPY . .
RUN rm -f go.mod go.sum
RUN go mod init rota-das-coisas
RUN go build -o integrador_bin ./integration/main.go
RUN chmod +x integrador_bin

