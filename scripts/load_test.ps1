# Скрипт для нагрузочного тестирования с помощью bombardier (PowerShell)
# Требует установки: go install github.com/codesenberg/bombardier@latest

$API_URL = if ($env:API_URL) { $env:API_URL } else { "http://localhost:8080" }
$CONCURRENT = if ($env:CONCURRENT) { $env:CONCURRENT } else { 50 }
$REQUESTS = if ($env:REQUESTS) { $env:REQUESTS } else { 20000 }

Write-Host "=========================================="
Write-Host "Load Testing PR Reviewer Service"
Write-Host "=========================================="
Write-Host "API URL: $API_URL"
Write-Host "Concurrent connections: $CONCURRENT"
Write-Host "Total requests: $REQUESTS"
Write-Host "=========================================="
Write-Host ""

# Проверяем наличие bombardier
$bombardierPath = Get-Command bombardier -ErrorAction SilentlyContinue
if (-not $bombardierPath) {
    Write-Host "bombardier not found. Installing..."
    go install github.com/codesenberg/bombardier@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error installing bombardier"
        exit 1
    }
    # Добавляем GOPATH/bin в PATH
    $gopath = go env GOPATH
    $env:PATH = "$gopath\bin;$env:PATH"
}

# Используем один флаг --print с перечислением через запятую
$printFlags = "r,p"

Write-Host "1. Testing GET /stats (statistics)"
Write-Host "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/stats" --print $printFlags
Write-Host ""

Write-Host "2. Testing GET /users (list of users)"
Write-Host "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/users" --print $printFlags
Write-Host ""

Write-Host "3. Testing GET /teams (list of teams)"
Write-Host "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/teams" --print $printFlags
Write-Host ""

Write-Host "4. Testing GET /prs (list of PRs)"
Write-Host "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/prs" --print $printFlags
Write-Host ""

Write-Host "5. Testing GET /prs?user_id=1 (PRs by user)"
Write-Host "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/prs?user_id=1" --print $printFlags
Write-Host ""

Write-Host "=========================================="
Write-Host "Load testing completed"
Write-Host "=========================================="
