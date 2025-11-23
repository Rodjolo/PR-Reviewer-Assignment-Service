FROM golang:1.24.0-alpine AS builder

# Обновляем пакеты
RUN apk update && apk upgrade --no-cache && apk add --no-cache git

WORKDIR /app

# 1. Сначала копируем описание зависимостей и скачиваем их
COPY go.mod go.sum* ./
RUN go mod download

# Устанавливаем swag
RUN go install github.com/swaggo/swag/cmd/swag@latest

# 2. Теперь копируем весь остальной код
COPY . .

# Генерируем Swagger
# Если swag упадет — значит, в коде есть ошибки комментариев, лучше увидеть это сейчас
RUN swag init -g cmd/server/main.go -o docs

# УДАЛЕНО: RUN go mod tidy
# (Эта команда не нужна внутри Docker, она пытается скачать зависимости для тестов/моков, которых нет)

# Собираем приложение (go build игнорирует _test.go файлы, поэтому отсутствие моков не помешает)
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/migrate ./cmd/migrate

# --- Финальный этап ---
FROM alpine:3.19

# Оставляем dos2unix, чтобы скрипт точно работал
RUN apk update && apk upgrade --no-cache && \
    apk add --no-cache ca-certificates tzdata dos2unix && \
    rm -rf /var/cache/apk/* /tmp/* /var/tmp/*

WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/migrate .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs

COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN dos2unix /docker-entrypoint.sh && chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]