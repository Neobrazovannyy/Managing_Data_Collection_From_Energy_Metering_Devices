FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ServerTCP ./ServerTCP.go

# Финальный этап
FROM alpine:latest

WORKDIR /root/

# Копируем бинарник из этапа сборки
COPY --from=builder /app/myapp .

# Копируем конфиги (если есть)
COPY --from=builder /app/config ./config

EXPOSE 8080

CMD ["./myapp"]