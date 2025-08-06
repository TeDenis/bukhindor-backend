package server

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/TeDenis/bukhindor-backend/internal/adapters/storage"
	"github.com/TeDenis/bukhindor-backend/internal/config"
	"github.com/TeDenis/bukhindor-backend/internal/service/auth"
	"github.com/TeDenis/bukhindor-backend/internal/web/api"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Server представляет HTTP сервер
type Server struct {
	app    *fiber.App
	config *config.Config
	logger *zap.Logger
	db     *sql.DB
	redis  *redis.Client
}

// New создает новый сервер
func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// Подключаемся к базе данных
	db, err := connectDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Подключаемся к Redis
	redisClient, err := connectRedis(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Создаем Fiber приложение
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
	})

	// Добавляем middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-App-Version, X-App-Type, X-Device-ID",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Создаем сервер
	server := &Server{
		app:    app,
		config: cfg,
		logger: logger,
		db:     db,
		redis:  redisClient,
	}

	// Настраиваем роуты
	if err := server.setupRoutes(); err != nil {
		return nil, err
	}

	return server, nil
}

// setupRoutes настраивает все роуты приложения
func (s *Server) setupRoutes() error {
	// Создаем storage сервис
	storageService := storage.NewService(s.db, s.redis, s.config, s.logger)

	// Создаем auth сервис
	authService := auth.NewService(
		storageService,
		storageService,
		storageService,
		storageService,
		s.config,
		s.logger,
	)

	// API роуты
	apiService := api.NewService(s.config, s.logger, authService)
	apiService.SetupRoutes(s.app)

	// Health check
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "bukhindor-backend",
		})
	})

	return nil
}

// Listen запускает сервер
func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

// ShutdownWithContext завершает работу сервера
func (s *Server) ShutdownWithContext(ctx context.Context) error {
	// Закрываем соединения с БД
	if s.db != nil {
		s.db.Close()
	}

	// Закрываем соединение с Redis
	if s.redis != nil {
		s.redis.Close()
	}

	return s.app.ShutdownWithContext(ctx)
}

// connectDB подключается к базе данных PostgreSQL напрямую
func connectDB(cfg *config.Config) (*sql.DB, error) {
	// Используем PostgreSQL напрямую
	dsn := cfg.GetPostgresDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// connectRedis подключается к Redis
func connectRedis(cfg *config.Config) (*redis.Client, error) {
	// Парсим Redis URL
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, err
	}

	// Устанавливаем пароль если есть
	if cfg.RedisPassword != "" {
		opt.Password = cfg.RedisPassword
	}

	// Устанавливаем номер БД
	opt.DB = cfg.RedisDB

	client := redis.NewClient(opt)

	// Проверяем соединение
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

// errorHandler обрабатывает ошибки приложения
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
		"code":  code,
	})
}
