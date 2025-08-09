# Build stage
FROM golang:1.23-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata

# Отключаем автоподкачку toolchain из go.mod
ENV GOTOOLCHAIN=local

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./

# Скачиваем зависимости с таймаутом и кешем
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
RUN go mod download && go mod verify

# Копируем исходный код
COPY . .

# Собираем бинарники API и CLI без CGO для избежания зависаний
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/cli ./cmd/cli

# Создаем .env если его нет
RUN touch .env

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
# Копируем конфиг
COPY --from=builder /app/.env .
RUN ls -l
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