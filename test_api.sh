#!/bin/bash

# Тестирование API Bukhindor Backend

BASE_URL="http://localhost:8080"
HEADERS="-H Content-Type: application/json -H X-App-Version: 1.0.0 -H X-App-Type: ios -H X-Device-ID: test-device-123"

echo "🧪 Тестирование API Bukhindor Backend"
echo "======================================"

# Тест 1: Health check
echo "1. Health check..."
curl -s -X GET "$BASE_URL/health" | jq .
echo ""

# Тест 2: Регистрация нового пользователя
echo "2. Регистрация пользователя..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" $HEADERS -d '{
  "name": "Test User 2",
  "email": "test2@example.com",
  "password": "password123"
}')
echo "$REGISTER_RESPONSE" | jq .
echo ""

# Тест 3: Вход пользователя
echo "3. Вход пользователя..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" $HEADERS -d '{
  "email": "test2@example.com",
  "password": "password123"
}' -c cookies.txt)
echo "$LOGIN_RESPONSE" | jq .
echo ""

# Тест 4: Получение информации о пользователе (с авторизацией)
echo "4. Получение информации о пользователе..."
ME_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/auth/me" $HEADERS -b cookies.txt)
echo "$ME_RESPONSE" | jq .
echo ""

# Тест 5: Попытка доступа без авторизации
echo "5. Попытка доступа без авторизации..."
UNAUTHORIZED_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/auth/me" $HEADERS)
echo "$UNAUTHORIZED_RESPONSE" | jq .
echo ""

# Тест 6: Сброс пароля
echo "6. Сброс пароля..."
RESET_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/reset-password" $HEADERS -d '{
  "email": "test2@example.com"
}')
echo "$RESET_RESPONSE" | jq .
echo ""

# Тест 7: Обновление токенов
echo "7. Обновление токенов..."
# Извлекаем refresh токен из предыдущего входа
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.user.refresh_token // empty')
if [ -n "$REFRESH_TOKEN" ]; then
    REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/refresh" $HEADERS -d "{
      \"refresh_token\": \"$REFRESH_TOKEN\"
    }")
    echo "$REFRESH_RESPONSE" | jq .
else
    echo "Refresh token not found in login response"
fi
echo ""

# Тест 8: Попытка входа с неверными данными
echo "8. Попытка входа с неверными данными..."
INVALID_LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" $HEADERS -d '{
  "email": "test2@example.com",
  "password": "wrongpassword"
}')
echo "$INVALID_LOGIN_RESPONSE" | jq .
echo ""

# Тест 9: Попытка доступа без обязательных заголовков
echo "9. Попытка доступа без обязательных заголовков..."
NO_HEADERS_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" -H "Content-Type: application/json" -d '{
  "email": "test2@example.com",
  "password": "password123"
}')
echo "$NO_HEADERS_RESPONSE" | jq .
echo ""

echo "✅ Тестирование завершено!" 