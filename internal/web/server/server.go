package server

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/TeDenis/bukhindor-backend/internal/adapters/storage"
	"github.com/TeDenis/bukhindor-backend/internal/config"
	"github.com/TeDenis/bukhindor-backend/internal/service/auth"
	"github.com/TeDenis/bukhindor-backend/internal/web/api"
)

// Server представляет HTTP сервер
type Server struct {
	app    *fiber.App
	config *config.Config
	logger *zap.Logger
	db     *pgxpool.Pool
	redis  *redis.Client
}

// New создает новый сервер
func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// Подключаемся к базе данных (pgxpool)
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
	// Безопасная настройка CORS: если указаны wildcard-источники, запрещаем креды
	allowOrigins := cfg.CORSAllowedOrigins
	var allowCredentials bool
	if allowOrigins != "*" {
		allowCredentials = true
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-App-Version, X-App-Type, X-Device-ID, X-Requested-With",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: allowCredentials,
		MaxAge:           300,
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
		_ = s.redis.Close()
	}

	return s.app.ShutdownWithContext(ctx)
}

// connectDB подключается к базе данных PostgreSQL через pgxpool
func connectDB(cfg *config.Config) (*pgxpool.Pool, error) {
	dsn := cfg.GetPostgresDSN()
	var pool *pgxpool.Pool
	var err error
	pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	err = pool.Ping(context.Background())
	if err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
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
