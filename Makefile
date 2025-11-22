.PHONY: build run test migrate-up migrate-down docker-up docker-down clean

# Переменные
# Используйте переменные окружения или создайте .env файл
# DB_URL должен быть установлен через переменную окружения DATABASE_URL
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

# Запуск E2E тестов
test-e2e:
	@echo "Running E2E tests..."
	@echo "Make sure PostgreSQL is running and database 'pr_reviewer_test' exists"
	@echo "Or set TEST_DATABASE_URL to use a different database"
	@go test -v ./test/e2e/... -timeout 30s

# Создание тестовой БД (PostgreSQL)
test-db-create:
	@echo "Creating test database..."
	@psql -U postgres -c "CREATE DATABASE pr_reviewer_test;" || echo "Database already exists or connection error"

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
		echo "Installing bombardier..."; \
		go install github.com/codesenberg/bombardier@latest; \
		echo "Running tests..."; \
		bombardier -c 50 -n 20000 http://localhost:8080/stats --print intro,progress,result; \
	fi

# Нагрузочное тестирование (Windows PowerShell)
load-test-win:
	@powershell -ExecutionPolicy Bypass -File scripts/load_test.ps1

# Быстрое нагрузочное тестирование (только stats)
load-test-quick:
	@echo "Installing bombardier (if needed)..."
	@go install github.com/codesenberg/bombardier@latest
	@echo "Testing GET /stats..."
	@bombardier -c 50 -n 20000 http://localhost:8080/stats --print intro,progress,result

# Упрощенное нагрузочное тестирование (PowerShell, без внешних зависимостей)
load-test-simple:
	@powershell -ExecutionPolicy Bypass -File scripts/load_test_simple.ps1

# Установка golangci-lint (рекомендуется использовать binary installation для Windows)
install-linter:
	@echo "Installing golangci-lint..."
	@echo "Note: If 'go install' fails on Windows, download binary from:"
	@echo "https://github.com/golangci/golangci-lint/releases"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Установка golangci-lint через бинарник (Windows PowerShell)
install-linter-win:
	@powershell -Command "Invoke-WebRequest -OutFile golangci-lint.zip -Uri https://github.com/golangci/golangci-lint/releases/download/v1.61.0/golangci-lint-1.61.0-windows-amd64.zip; Expand-Archive -Path golangci-lint.zip -DestinationPath .; Move-Item -Path golangci-lint-1.61.0-windows-amd64\golangci-lint.exe -Destination $(GOPATH)\bin\; Remove-Item -Recurse -Force golangci-lint-1.61.0-windows-amd64, golangci-lint.zip"

# Запуск линтера
lint:
	@which golangci-lint > /dev/null 2>&1 || where golangci-lint > nul 2>&1 || (echo "golangci-lint not found. Run 'make install-linter' or 'make install-linter-win' first" && exit 1)
	golangci-lint run ./...

# Запуск линтера с автоисправлением
lint-fix:
	@which golangci-lint > /dev/null 2>&1 || where golangci-lint > nul 2>&1 || (echo "golangci-lint not found. Run 'make install-linter' or 'make install-linter-win' first" && exit 1)
	golangci-lint run --fix ./...

# Базовая проверка через встроенные инструменты Go (альтернатива golangci-lint)
lint-basic:
	@echo "Running go vet..."
	@go vet ./internal/... ./pkg/...
	@echo "Checking code formatting..."
	@gofmt -l internal pkg || echo "Some files need formatting. Run 'gofmt -w internal pkg' to fix"
	@echo "Building to check for errors..."
	@go build ./...
	@echo "OK - Basic lint checks passed!"

