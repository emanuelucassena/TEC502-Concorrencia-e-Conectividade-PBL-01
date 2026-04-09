FROM golang:1.20-alpine
WORKDIR /app
ENV GOPROXY=off
COPY . .
RUN rm -f go.mod go.sum
RUN go mod init rota-das-coisas
RUN go build -o atuador_bin ./device/atuador/atuador.go
RUN go build -o sensor_bin ./device/sensor/sensor.go
RUN chmod +x atuador_bin sensor_bin

