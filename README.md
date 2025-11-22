# Сервис назначения ревьюеров для Pull Request'ов

Микросервис для автоматического назначения ревьюеров на Pull Request'ы с управлением командами и участниками.

## Описание

Сервис предоставляет REST API для:
- Автоматического назначения ревьюеров на PR из команды автора
- Переназначения ревьюверов
- Получения списка PR'ов по пользователю
- Управления командами и активностью пользователей

## Технологии

- **Язык**: Go 1.21
- **База данных**: PostgreSQL 15
- **Роутинг**: gorilla/mux
- **Миграции**: golang-migrate

## Быстрый старт

### Через Docker Compose (рекомендуется)

```bash
docker-compose up --build
```

Сервис будет доступен на `http://localhost:8080`

### Локальный запуск

1. Установите зависимости:
```bash
make deps
```

2. Запустите PostgreSQL (или используйте docker-compose только для БД):
```bash
docker-compose up postgres -d
```

3. Примените миграции:
```bash
make migrate-up
```

4. (Опционально) Заполните БД тестовыми данными:
```bash
make seed
```

5. Запустите сервер:
```bash
make run
```

## Структура проекта

```
.
├── cmd/
│   ├── migrate/      # Утилита для миграций
│   └── server/        # Основной сервер
├── internal/
│   ├── database/     # Подключение к БД
│   ├── handlers/     # HTTP handlers
│   ├── models/       # Модели данных
│   ├── repository/   # Репозитории для работы с БД
│   ├── router/       # Роутинг
│   └── service/      # Бизнес-логика
├── migrations/       # SQL миграции
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── openapi.yaml      # OpenAPI спецификация
```

## API Endpoints

### Pull Requests

- `POST /prs` - Создать PR (автоматически назначает до 2 ревьюверов)
- `GET /prs` - Список всех PR'ов
- `GET /prs?user_id={id}` - PR'ы пользователя (как автор или ревьюер)
- `GET /prs/{id}` - Получить PR по ID
- `PATCH /prs/{id}/reassign` - Переназначить ревьювера
- `POST /prs/{id}/merge` - Мержить PR

### Пользователи

- `POST /users` - Создать пользователя
- `GET /users` - Список всех пользователей
- `GET /users/{id}` - Получить пользователя по ID
- `PATCH /users/{id}` - Обновить пользователя

### Команды

- `POST /teams` - Создать команду
- `GET /teams` - Список всех команд
- `GET /teams/{name}` - Получить команду по имени
- `POST /teams/{name}/members` - Добавить участника в команду
- `DELETE /teams/{name}/members?user_id={id}` - Удалить участника из команды

### Статистика

- `GET /stats` - Получить статистику (количество пользователей, команд, PR'ов и т.д.)

### Swagger документация

- `GET /swagger/index.html` - Интерактивная Swagger UI документация

## Правила назначения ревьюверов

### При создании PR:
1. Автоматически выбираются до 2 активных пользователей из команды автора
2. Автор исключается из кандидатов
3. Если активных меньше 2, назначаются только доступные (0/1)
4. Назначение случайное (ORDER BY RANDOM() в PostgreSQL)

### Переназначение:
- Заменяет одного ревьювера на случайного активного участника из **команды заменяемого ревьювера** (не из команды автора!)
- Автор PR также исключается из кандидатов

### После MERGED:
- Любые изменения ревьюверов запрещены
- Операции переназначения возвращают ошибку 409 Conflict

## Допущения и решения

1. **Один пользователь может быть только в одной команде** - упрощает логику и соответствует типичным сценариям использования
2. **Merge - идемпотентная операция** - повторный merge уже мерженного PR возвращает успешный ответ с актуальным состоянием PR (200 OK)
3. **При переназначении новый ревьювер выбирается из команды старого ревьювера** - это ключевое бизнес-правило из требований
4. **Если в команде нет доступных ревьюверов, PR создается без ревьюверов** - это допустимо согласно требованиям
5. **Статус PR хранится как строка** ('OPEN' или 'MERGED') с CHECK constraint в БД

## Переменные окружения

- `DATABASE_URL` - строка подключения к PostgreSQL (по умолчанию: `postgres://postgres:postgres@localhost:5432/pr_reviewer?sslmode=disable`)
- `PORT` - порт для HTTP сервера (по умолчанию: `8080`)

## Makefile команды

- `make build` - собрать приложение
- `make run` - запустить сервер локально
- `make test` - запустить тесты
- `make migrate-up` - применить миграции
- `make migrate-down` - откатить миграции
- `make seed` - заполнить БД тестовыми данными (пользователи и команды)
- `make docker-up` - запустить через docker-compose
- `make docker-down` - остановить docker-compose
- `make clean` - очистить бинарники и volumes
- `make deps` - установить зависимости
- `make swagger` - сгенерировать Swagger документацию
- `make install-bombardier` - установить bombardier для нагрузочного тестирования
- `make load-test` - запустить нагрузочное тестирование (Linux/Mac)
- `make load-test-win` - запустить нагрузочное тестирование (Windows)
- `make load-test-quick` - быстрое тестирование только /stats
- `make load-test-simple` - упрощенное тестирование (PowerShell, без внешних зависимостей)

## Заполнение базы данных

### Автоматическое заполнение (рекомендуется)

После применения миграций можно заполнить БД тестовыми данными:

```bash
make seed
```

Это создаст:
- **6 пользователей** (5 активных: Alice, Bob, Charlie, David, Frank; 1 неактивный: Eve)
- **3 команды**: backend, frontend, devops
- **Распределение пользователей по командам**:
  - `backend`: Alice, Bob, Charlie
  - `frontend`: David, Frank
  - `devops`: Bob

### Ручное создание данных

Или создайте данные вручную через API:

```bash
# Создать пользователей
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "is_active": true}'

curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "is_active": true}'

curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Charlie", "is_active": true}'

# Создать команду
curl -X POST http://localhost:8080/teams \
  -H "Content-Type: application/json" \
  -d '{"name": "backend"}'

# Добавить участников в команду
curl -X POST http://localhost:8080/teams/backend/members \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1}'

curl -X POST http://localhost:8080/teams/backend/members \
  -H "Content-Type: application/json" \
  -d '{"user_id": 2}'

curl -X POST http://localhost:8080/teams/backend/members \
  -H "Content-Type: application/json" \
  -d '{"user_id": 3}'
```

### Создание PR

```bash
curl -X POST http://localhost:8080/prs \
  -H "Content-Type: application/json" \
  -d '{"title": "Add new feature", "author_id": 1}'
```

PR автоматически получит до 2 случайных ревьюверов из команды автора.

### Переназначение ревьювера

```bash
curl -X PATCH http://localhost:8080/prs/1/reassign \
  -H "Content-Type: application/json" \
  -d '{"old_reviewer_id": 2}'
```

### Merge PR

```bash
curl -X POST http://localhost:8080/prs/1/merge
```

## Swagger документация

После запуска сервиса Swagger UI доступен по адресу:
- **http://localhost:8080/swagger/index.html**

Документация генерируется автоматически при сборке Docker образа. Для локальной разработки используйте:
```bash
make swagger
```

## Нагрузочное тестирование

Для проведения нагрузочного тестирования используется [bombardier](https://github.com/codesenberg/bombardier).

### Установка bombardier

```bash
make install-bombardier
# или
go install github.com/codesenberg/bombardier@latest
```

### Запуск тестов

**Важно:** Перед тестированием убедитесь, что Docker Desktop запущен и сервис доступен:
```bash
docker-compose up -d
curl http://localhost:8080/stats
```

**Linux/Mac:**
```bash
make load-test
```

**Windows:**
```bash
make load-test-win
```

**Быстрый тест (только /stats):**
```bash
make load-test-quick
```

**Упрощенный вариант (без внешних зависимостей, PowerShell):**
```bash
make load-test-simple
```

Если возникают проблемы, см. подробную инструкцию в [LOAD_TESTING.md](LOAD_TESTING.md)

### Настройка параметров

Можно настроить параметры через переменные окружения:

```bash
# Linux/Mac
CONCURRENT=100 REQUESTS=50000 make load-test

# Windows PowerShell
$env:CONCURRENT=100; $env:REQUESTS=50000; make load-test-win
```

Параметры по умолчанию:
- Concurrent connections: 50
- Total requests: 20000

### Пример результатов

```
Statistics        Avg      Stdev        Max
  Reqs/sec      5000.00    500.00    5500.00
  Latency       10.00ms    2.00ms    50.00ms
  HTTP codes:
    1xx - 0, 2xx - 20000, 3xx - 0, 4xx - 0, 5xx - 0
```

## Производительность

Сервис рассчитан на:
- До 20 команд
- До 200 пользователей
- 5 RPS
- SLI 300 мс
- Доступность 99.9%


