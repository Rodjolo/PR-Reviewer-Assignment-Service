#!/bin/bash

# Скрипт для нагрузочного тестирования с помощью bombardier
# Требует установки: go install github.com/codesenberg/bombardier@latest

API_URL="${API_URL:-http://localhost:8080}"
CONCURRENT="${CONCURRENT:-50}"
REQUESTS="${REQUESTS:-20000}"

echo "=========================================="
echo "Нагрузочное тестирование PR Reviewer Service"
echo "=========================================="
echo "API URL: $API_URL"
echo "Concurrent connections: $CONCURRENT"
echo "Total requests: $REQUESTS"
echo "=========================================="
echo ""

# Проверяем наличие bombardier
if ! command -v bombardier &> /dev/null; then
    echo "bombardier не найден. Устанавливаю..."
    go install github.com/codesenberg/bombardier@latest
    if [ $? -ne 0 ]; then
        echo "Ошибка установки bombardier"
        exit 1
    fi
fi

echo "1. Тестирование GET /stats (статистика)"
echo "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/stats" \
    --print r --print p --print h
echo ""

echo "2. Тестирование GET /users (список пользователей)"
echo "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/users" \
    --print r --print p --print h
echo ""

echo "3. Тестирование GET /teams (список команд)"
echo "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/teams" \
    --print r --print p --print h
echo ""

echo "4. Тестирование GET /prs (список PR)"
echo "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/prs" \
    --print r --print p --print h
echo ""

echo "5. Тестирование GET /prs?user_id=1 (PR по пользователю)"
echo "----------------------------------------"
bombardier -c $CONCURRENT -n $REQUESTS "$API_URL/prs?user_id=1" \
    --print r --print p --print h
echo ""

echo "=========================================="
echo "Нагрузочное тестирование завершено"
echo "=========================================="


