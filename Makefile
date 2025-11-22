.PHONY: build run test migrate-up migrate-down docker-up docker-down clean

# Переменные
DB_URL ?= postgres://postgres:postgres@localhost:5432/pr_reviewer?sslmode=disable
PORT ?= 8080

# Сборка приложения
build:
	go build -o bin/server ./cmd/server
	go build -o bin/migrate ./cmd/migrate

# Запуск сервера локально
run:
	go run ./cmd/server

# Запуск тестов
test:
	go test -v ./...

# Миграции вверх
migrate-up:
	go run ./cmd/migrate -up

# Миграции вниз
migrate-down:
	go run ./cmd/migrate -down

# Заполнить БД тестовыми данными (seed)
seed:
	go run ./cmd/migrate -up

# Запуск через docker-compose
docker-up:
	docker-compose up --build

# Остановка docker-compose
docker-down:
	docker-compose down

# Очистка
clean:
	rm -rf bin/
	docker-compose down -v

# Установка зависимостей
deps:
	go mod download
	go mod tidy

# Генерация Swagger документации
swagger:
	@which swag > /dev/null || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g cmd/server/main.go -o docs

# Установка bombardier для нагрузочного тестирования
install-bombardier:
	go install github.com/codesenberg/bombardier@latest

# Нагрузочное тестирование (Linux/Mac)
load-test:
	@if [ -f scripts/load_test.sh ]; then \
		chmod +x scripts/load_test.sh && \
		./scripts/load_test.sh; \
	else \
		echo "Установка bombardier..."; \
		go install github.com/codesenberg/bombardier@latest; \
		echo "Запуск тестов..."; \
		bombardier -c 50 -n 20000 http://localhost:8080/stats --print r --print p --print h; \
	fi

# Нагрузочное тестирование (Windows PowerShell)
load-test-win:
	@powershell -ExecutionPolicy Bypass -File scripts/load_test.ps1

# Быстрое нагрузочное тестирование (только stats)
load-test-quick:
	@echo "Установка bombardier (если нужно)..."
	@go install github.com/codesenberg/bombardier@latest
	@echo "Тестирование GET /stats..."
	@bombardier -c 50 -n 20000 http://localhost:8080/stats --print r --print p --print h

# Упрощенное нагрузочное тестирование (PowerShell, без внешних зависимостей)
load-test-simple:
	@powershell -ExecutionPolicy Bypass -File scripts/load_test_simple.ps1

