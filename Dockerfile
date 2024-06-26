# Stage 1: Сборка приложения
FROM golang:1.22 AS builder

WORKDIR /app

COPY . .
RUN go mod tidy

WORKDIR /app/cmd/birthday-notify
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/birthday

# Stage 2: Создание минимального образа для запуска
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/birthday .

RUN pwd && ls

# Команда для запуска приложения
CMD ["./birthday"]
