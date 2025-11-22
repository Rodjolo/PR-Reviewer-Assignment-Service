# Скрипт для нагрузочного тестирования с помощью bombardier (PowerShell)
# Требует установки: go install github.com/codesenberg/bombardier@latest

$API_URL = if ($env:API_URL) { $env:API_URL } else { "http://localhost:8080" }
$CONCURRENT = if ($env:CONCURRENT) { $env:CONCURRENT } else { 50 }
$REQUESTS = if ($env:REQUESTS) { $env:REQUESTS } else { 20000 }

Write-Host "=========================================="
Write-Host "Нагрузочное тестирование PR Reviewer Service"
Write-Host "=========================================="
Write-Host "API URL: $API_URL"
Write-Host "Concurrent connections: $CONCURRENT"
Write-Host "Total requests: $REQUESTS"
Write-Host "=========================================="
Write-Host ""

# Проверяем наличие bombardier
$bombardierPath = Get-Command bombardier -ErrorAction SilentlyContinue
if (-not $bombardierPath) {
    Write-Host "bombardier не найден. Устанавливаю..."
    go install github.com/codesenberg/bombardier@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Ошибка установки bombardier"
        exit 1
    }
    # Добавляем GOPATH/bin в PATH
    $gopath = go env GOPATH
    $env:PATH = "$gopath\bin;$env:PATH"
}

Write-Host "1. Тестирование GET /stats (статистика)"
Write-Host "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/stats" --print r --print p --print h
Write-Host ""

Write-Host "2. Тестирование GET /users (список пользователей)"
Write-Host "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/users" --print r --print p --print h
Write-Host ""

Write-Host "3. Тестирование GET /teams (список команд)"
Write-Host "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/teams" --print r --print p --print h
Write-Host ""

Write-Host "4. Тестирование GET /prs (список PR)"
Write-Host "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/prs" --print r --print p --print h
Write-Host ""

Write-Host "5. Тестирование GET /prs?user_id=1 (PR по пользователю)"
Write-Host "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/prs?user_id=1" --print r --print p --print h
Write-Host ""

Write-Host "=========================================="
Write-Host "Нагрузочное тестирование завершено"
Write-Host "=========================================="
