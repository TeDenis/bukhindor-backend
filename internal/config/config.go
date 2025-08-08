package config

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

// Config содержит все настройки приложения
type Config struct {
	// Сервер
	ServerPort string `env:"SERVER_PORT" envDefault:"8080"`
	ServerHost string `env:"SERVER_HOST" envDefault:"localhost"`

	// База данных PostgreSQL
	PostgresHost     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	PostgresPort     string `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresUser     string `env:"POSTGRES_USER" envDefault:"bukhindor"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" envDefault:"password"`
	PostgresDB       string `env:"POSTGRES_DB" envDefault:"bukhindor"`
	PostgresSSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`

	// Redis
	RedisURL      string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	// Логирование
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`     // debug, info, warn, error
	LogFormat string `env:"LOG_FORMAT" envDefault:"console"` // console, json

	// JWT
	JWTSecret              string        `env:"JWT_SECRET" envDefault:"your-secret-key"`
	JWTExpiration          time.Duration `env:"JWT_EXPIRATION" envDefault:"15m"`
	RefreshTokenExpiration time.Duration `env:"REFRESH_TOKEN_EXPIRATION" envDefault:"7d"`

	// CORS
	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"*"`

	// Пароли
	MinPasswordLength int `env:"MIN_PASSWORD_LENGTH" envDefault:"6"`

	// Метрики
	MetricsPort string `env:"METRICS_PORT" envDefault:"9090"`
}

// New создает новую конфигурацию из переменных окружения
func New() *Config {
	cfg := &Config{
		ServerPort:             getEnv("SERVER_PORT", "7080"),
		ServerHost:             getEnv("SERVER_HOST", "localhost"),
		PostgresHost:           getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:           getEnv("POSTGRES_PORT", "5433"),
		PostgresUser:           getEnv("POSTGRES_USER", "bukhindor"),
		PostgresPassword:       getEnv("POSTGRES_PASSWORD", ""),
		PostgresDB:             getEnv("POSTGRES_DB", "bukhindor"),
		PostgresSSLMode:        getEnv("POSTGRES_SSLMODE", "disable"),
		RedisURL:               getEnv("REDIS_URL", "redis://localhost:6380"),
		RedisPassword:          getEnv("REDIS_PASSWORD", ""),
		RedisDB:                getEnvAsInt("REDIS_DB", 0),
		LogLevel:               getEnv("LOG_LEVEL", "info"),
		LogFormat:              getEnv("LOG_FORMAT", "console"),
		JWTSecret:              getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiration:          getEnvAsDuration("JWT_EXPIRATION", 15*time.Minute),
		RefreshTokenExpiration: getEnvAsDuration("REFRESH_TOKEN_EXPIRATION", 7*24*time.Hour),
		CORSAllowedOrigins:     getEnv("CORS_ALLOWED_ORIGINS", "*"),
		MinPasswordLength:      getEnvAsInt("MIN_PASSWORD_LENGTH", 6),
		MetricsPort:            getEnv("METRICS_PORT", "9090"),
	}

	return cfg
}

// GetPostgresDSN возвращает строку подключения к PostgreSQL
func (c *Config) GetPostgresDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresDB, c.PostgresSSLMode)
}

// NewLogger создает новый логгер на основе конфигурации
func NewLogger(cfg *Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	switch cfg.LogFormat {
	case "json":
		zapConfig = zap.NewProductionConfig()
	case "console":
		zapConfig = zap.NewDevelopmentConfig()
	default:
		return nil, fmt.Errorf("unknown log format: %s", cfg.LogFormat)
	}

	level, err := zap.ParseAtomicLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	zapConfig.Level = level

	return zapConfig.Build()
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает значение переменной окружения как int или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

// getEnvAsDuration получает значение переменной окружения как Duration или возвращает значение по умолчанию
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
