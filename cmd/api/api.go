package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TeDenis/bukhindor-backend/internal/config"
	"github.com/TeDenis/bukhindor-backend/internal/web/server"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Инициализируем конфигурацию
	cfg := config.New()

	// Создаем логгер
	logger, err := config.NewLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Создаем сервер
	app, err := server.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	// Канал для сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Запускаем сервер в горутине
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.ServerPort))
		if err := app.Listen(":" + cfg.ServerPort); err != nil {
			logger.Error("Server failed to start", zap.Error(err))
		}
	}()

	// Ждем сигнала завершения
	sig := <-sigChan
	logger.Info("Received signal", zap.String("signal", sig.String()))

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("Server shutdown failed", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}
