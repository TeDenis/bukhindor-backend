.PHONY: build run test lint clean migrate-up migrate-down migrate-status deps generate dev start lint-install

# Переменные
BINARY_NAME=bukhindor-backend
CLI_NAME=bukhindor-cli
BUILD_DIR=build
GOLANGCI_BIN := $(shell go env GOPATH)/bin/golangci-lint
GOLANGCI_VERSION := v2.3.0
TOOLCHAIN := $(shell awk '/^toolchain/{print $$2}' go.mod)

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
	GOTOOLCHAIN=$(TOOLCHAIN) go test -v ./...

# Установка golangci-lint нужной версии
lint-install:
	@echo "Installing golangci-lint $(GOLANGCI_VERSION)..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_VERSION)

# Линтер (как в CI: v1.60.3, timeout 5m)
lint:
	@echo "Running linter (golangci-lint $(GOLANGCI_VERSION))..."
	@if ! $(GOLANGCI_BIN) version 2>/dev/null | grep -q "version 2.3.0"; then \
		echo "Installing/Updating golangci-lint to $(GOLANGCI_VERSION)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_VERSION); \
	fi
	@echo "Ensuring modules are downloaded..."; \
	GOTOOLCHAIN=$(TOOLCHAIN) go mod download; \
	GOTOOLCHAIN=$(TOOLCHAIN) $(GOLANGCI_BIN) run --modules-download-mode=mod --timeout=5m

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