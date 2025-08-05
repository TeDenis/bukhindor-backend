.PHONY: build run test lint clean migrate-up migrate-down migrate-status

# Переменные
BINARY_NAME=bukhindor-backend
CLI_NAME=bukhindor-cli
BUILD_DIR=build

# Сборка
build:
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api
	go build -o $(BUILD_DIR)/$(CLI_NAME) ./cmd/cli
	@echo "Build completed!"

# Запуск
run:
	@echo "Starting application..."
	go run ./cmd/api

# Тестирование
test:
	@echo "Running tests..."
	go test -v ./...

# Линтер
lint:
	@echo "Running linter..."
	golangci-lint run

# Очистка
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Миграции
migrate-up:
	@echo "Running migrations up..."
	go run ./cmd/cli migrate up

migrate-down:
	@echo "Running migrations down..."
	go run ./cmd/cli migrate down

migrate-status:
	@echo "Checking migration status..."
	go run ./cmd/cli migrate status

# Установка зависимостей
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Генерация моков
generate:
	@echo "Generating mocks..."
	go generate ./...

# Полный цикл разработки
dev: deps lint test build

# Запуск с миграциями
start: migrate-up run 