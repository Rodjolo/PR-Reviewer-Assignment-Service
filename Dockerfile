FROM golang:1.21-alpine3.19 AS builder

# Обновляем пакеты до последних версий с исправлениями безопасности
RUN apk update && apk upgrade --no-cache && apk add --no-cache git

WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum* ./
RUN go mod download

# Устанавливаем swag для генерации Swagger документации
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Копируем исходный код
COPY . .

# Генерируем Swagger документацию
RUN swag init -g cmd/server/main.go -o docs || true

# Генерируем go.sum на основе всех импортов
RUN go mod tidy

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/migrate ./cmd/migrate

FROM alpine:3.19

# Обновляем все пакеты до последних версий с исправлениями безопасности
RUN apk update && apk upgrade --no-cache && \
    apk add --no-cache ca-certificates tzdata && \
    rm -rf /var/cache/apk/* /tmp/* /var/tmp/*

WORKDIR /root/

# Копируем бинарники
COPY --from=builder /app/server .
COPY --from=builder /app/migrate .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs

# Запускаем миграции и сервер
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]

