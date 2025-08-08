# Build stage
FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарники API и CLI
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/bin/cli ./cmd/cli
COPY .env ./cmd/.env

# Final stage
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates tzdata curl

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем бинарные файлы из builder stage
COPY --from=builder /app/bin/api /usr/local/bin/api
COPY --from=builder /app/bin/cli /usr/local/bin/cli

# Копируем миграции в ожидаемый путь CLI
COPY --from=builder /app/deployments/postgres/migrations /root/deployments/postgres/migrations

# Меняем владельца файлов
RUN chown -R appuser:appgroup /root/

# Переключаемся на непривилегированного пользователя
USER appuser

# Открываем порты
EXPOSE 8080 9091

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Запускаем приложение по умолчанию (API)
CMD ["api"]