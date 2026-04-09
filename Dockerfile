FROM golang:1.25-alpine
WORKDIR /app
COPY . .
RUN go build -o integrador_bin integration/main.go
RUN go build -o sensor_bin device/sensor/sensor.go
RUN go build -o atuador_bin device/atuador/atuador.go