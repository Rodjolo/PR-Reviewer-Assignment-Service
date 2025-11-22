# Инструкция по нагрузочному тестированию

## Требования

1. **Docker Desktop должен быть запущен**
   - Убедитесь, что Docker Desktop запущен перед тестированием
   - Проверьте: `docker ps` должен работать без ошибок

2. **Сервис должен быть запущен**
   ```bash
   docker-compose up -d
   ```
   Проверьте доступность: `curl http://localhost:8080/stats`

## Способы тестирования

### 1. С помощью bombardier (рекомендуется)

**Установка:**
```bash
go install github.com/codesenberg/bombardier@latest
```

Если возникают проблемы с сетью при установке:
- Проверьте интернет-соединение
- Попробуйте установить через прокси
- Или используйте упрощенный вариант (см. ниже)

**Запуск:**
```bash
# Windows
make load-test-win

# Linux/Mac
make load-test

# Быстрый тест
make load-test-quick
```

**Ручной запуск:**
```bash
bombardier -c 50 -n 20000 http://localhost:8080/stats --print intro,progress,result
```

### 2. Упрощенный вариант (PowerShell, без внешних зависимостей)

Если bombardier не устанавливается, используйте встроенный скрипт:

```bash
make load-test-simple
```

Или напрямую:
```powershell
powershell -ExecutionPolicy Bypass -File scripts/load_test_simple.ps1
```

Этот скрипт использует только встроенные возможности PowerShell и не требует внешних зависимостей.

### 3. Ручное тестирование с помощью curl

Для быстрой проверки производительности:

```bash
# Windows PowerShell
Measure-Command { 1..100 | ForEach-Object { Invoke-WebRequest -Uri "http://localhost:8080/stats" -UseBasicParsing } }
```

## Настройка параметров

### Для bombardier:
```bash
# Через переменные окружения
$env:CONCURRENT=100
$env:REQUESTS=50000
make load-test-win

# Или напрямую
bombardier -c 100 -n 50000 http://localhost:8080/stats
```

### Для упрощенного скрипта:
```powershell
$env:CONCURRENT=50
$env:REQUESTS=200
powershell -ExecutionPolicy Bypass -File scripts/load_test_simple.ps1
```

## Интерпретация результатов

### Bombardier выводит:
- **Reqs/sec** - запросов в секунду (RPS)
- **Latency** - задержка (latency)
- **HTTP codes** - распределение кодов ответа

### Целевые показатели:
- **RPS**: минимум 5 (требование), ожидается больше
- **Latency**: средняя задержка должна быть < 300 мс (SLI)
- **Success rate**: 99.9% запросов должны быть успешными (2xx коды)

## Решение проблем

### Проблема: "Docker Desktop не запущен"
**Решение:** Запустите Docker Desktop и подождите его полной загрузки

### Проблема: "Сервис недоступен"
**Решение:**
```bash
docker-compose up -d
# Подождите 5-10 секунд
curl http://localhost:8080/stats
```

### Проблема: "bombardier не устанавливается"
**Решение:** Используйте упрощенный скрипт `make load-test-simple`

### Проблема: "TLS handshake timeout"
**Решение:** 
- Проверьте интернет-соединение
- Попробуйте установить позже
- Используйте упрощенный вариант тестирования

## Пример успешного теста

```
Bombardier 1.2.6
Running 20000 request(s) @ http://localhost:8080/stats
100% |████████████████████████████████| [20000/20000] [00:04<00:00, 4500 req/s]

Statistics        Avg      Stdev        Max
  Reqs/sec      4500.00    200.00    4800.00
  Latency        11.00ms    3.00ms    45.00ms
  HTTP codes:
    1xx - 0, 2xx - 20000, 3xx - 0, 4xx - 0, 5xx - 0
```

