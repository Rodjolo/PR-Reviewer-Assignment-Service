# Упрощенный скрипт нагрузочного тестирования
# Использует встроенные возможности PowerShell для тестирования

$API_URL = if ($env:API_URL) { $env:API_URL } else { "http://localhost:8081" }
$CONCURRENT = if ($env:CONCURRENT) { $env:CONCURRENT } else { 50 }
$REQUESTS = if ($env:REQUESTS) { $env:REQUESTS } else { 200 }

Write-Host "=========================================="
Write-Host "Нагрузочное тестирование PR Reviewer Service"
Write-Host "=========================================="
Write-Host "API URL: $API_URL"
Write-Host "Concurrent requests: $CONCURRENT"
Write-Host "Total requests: $REQUESTS"
Write-Host "=========================================="
Write-Host ""

# Функция для выполнения запроса
function Test-Endpoint {
    param(
        [string]$Url,
        [string]$Name
    )
    
    Write-Host "Тестирование: $Name"
    Write-Host "URL: $Url"
    
    $results = @()
    $errors = 0
    $success = 0
    $totalTime = 0
    
    $jobs = @()
    for ($i = 0; $i -lt $CONCURRENT; $i++) {
        $jobs += Start-Job -ScriptBlock {
            param($url, $requestsPerJob)
            $results = @()
            for ($j = 0; $j -lt $requestsPerJob; $j++) {
                try {
                    $start = Get-Date
                    $response = Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
                    $end = Get-Date
                    $duration = ($end - $start).TotalMilliseconds
                    $results += @{
                        StatusCode = $response.StatusCode
                        Duration = $duration
                        Success = $true
                    }
                } catch {
                    $results += @{
                        StatusCode = 0
                        Duration = 0
                        Success = $false
                        Error = $_.Exception.Message
                    }
                }
            }
            return $results
        } -ArgumentList $Url, ([math]::Ceiling($REQUESTS / $CONCURRENT))
    }
    
    Write-Host "Выполнение запросов..."
    $allResults = $jobs | Wait-Job | Receive-Job
    $jobs | Remove-Job
    
    $allResults | ForEach-Object {
        if ($_.Success) {
            $success++
            $totalTime += $_.Duration
        } else {
            $errors++
        }
    }
    
    $avgTime = if ($success -gt 0) { $totalTime / $success } else { 0 }
    $rps = if ($avgTime -gt 0) { [math]::Round(1000 / $avgTime, 2) } else { 0 }
    
    Write-Host "Результаты:"
    Write-Host "  Успешных: $success"
    Write-Host "  Ошибок: $errors"
    Write-Host "  Среднее время ответа: $([math]::Round($avgTime, 2)) мс"
    Write-Host "  Примерная RPS: $rps"
    Write-Host ""
}

# Тестируем эндпоинты
Test-Endpoint -Url "$API_URL/stats" -Name "GET /stats"
Test-Endpoint -Url "$API_URL/users" -Name "GET /users"
Test-Endpoint -Url "$API_URL/teams" -Name "GET /teams"
Test-Endpoint -Url "$API_URL/prs" -Name "GET /prs"

Write-Host "=========================================="
Write-Host "Тестирование завершено"
Write-Host "=========================================="
Write-Host ""
Write-Host "Примечание: Для более точного тестирования рекомендуется использовать bombardier:"
Write-Host "  go install github.com/codesenberg/bombardier@latest"
Write-Host "  bombardier -c 50 -n 20000 http://localhost:8081/stats"


