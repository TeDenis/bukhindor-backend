# System Patterns - Bukhindor Backend

## Архитектурные принципы

### Чистая гексагональная архитектура
```
┌─────────────────────────────────────────────────────────────┐
│                        Web Layer                            │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │   API Service   │  │  Pages Service  │  │   Server    │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │  User Service   │  │  Auth Service   │  │  Crypto     │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Adapters Layer                            │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │   Storage       │  │  External API   │  │  Cache      │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                     Domain Layer                            │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │     User        │  │     Session     │  │   Types     │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Направление зависимостей
- **Web → Service**: Web слои используют сервисы через интерфейсы
- **Service → Adapters**: Сервисы используют адаптеры через интерфейсы
- **Adapters → Domain**: Адаптеры работают с доменными структурами
- **Запрещено**: Обратные зависимости между слоями

## Ключевые паттерны

### 1. Dependency Injection через интерфейсы
```go
// external.go - интерфейсы определяются в потребляющем пакете
type UserRepository interface {
    CreateUser(ctx context.Context, user *domain.User) error
    GetUserByID(ctx context.Context, id string) (*domain.User, error)
}

// service.go - сервис принимает интерфейсы
type UserService struct {
    repo UserRepository
    logger *zap.Logger
}

func NewUserService(repo UserRepository, logger *zap.Logger) *UserService {
    return &UserService{repo: repo, logger: logger}
}
```

### 2. Структура файлов в пакетах
```
internal/service/users/
├── external.go      # Внешние интерфейсы (зависимости)
├── service.go       # Структура, конструктор, типы
├── users.go         # Основная бизнес-логика
├── validation.go    # Валидация данных
└── users_test.go    # Тесты
```

### 3. Обработка ошибок
```go
// Единообразная обработка ошибок
func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (*domain.User, error) {
    if err := s.validateCreateUser(input); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    user := &domain.User{
        ID:        uuid.New().String(),
        Email:     input.Email,
        Username:  input.Username,
        CreatedAt: time.Now(),
    }
    
    if err := s.repo.CreateUser(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}
```

### 4. Middleware паттерн
```go
// Логирование, CORS, Recovery middleware
app.Use(recover.New())
app.Use(cors.New(cors.Config{
    AllowOrigins: cfg.CORSAllowedOrigins,
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
}))
```

## Технические решения

### HTTP Framework: Fiber
- **Выбор**: Fiber v2 для высокой производительности
- **Преимущества**: 
  - Быстрый HTTP сервер
  - Middleware поддержка
  - JSON обработка
  - Graceful shutdown
- **Альтернативы**: Gin, Echo (отклонены в пользу Fiber)

### База данных: SQLite + PostgreSQL
- **Разработка**: SQLite для простоты
- **Продакшен**: PostgreSQL для масштабируемости
- **Query Builder**: Squirrel для построения SQL
- **Миграции**: Goose для управления схемой

### Логирование: Zap
- **Структурированное логирование**: JSON и console форматы
- **Производительность**: Высокая скорость записи
- **Конфигурируемость**: Уровни логирования через переменные окружения

### Аутентификация: JWT
- **Stateless**: Не требует хранения сессий на сервере
- **Безопасность**: Подпись токенов секретным ключом
- **Управление**: Создание, валидация, инвалидация токенов

### Конфигурация: Environment Variables
- **Гибкость**: Настройки через переменные окружения
- **Безопасность**: Секреты не в коде
- **Развертывание**: Легкая настройка для разных окружений

## Паттерны тестирования

### 1. Табличные тесты
```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserInput
        want    *domain.User
        wantErr bool
        setup   func(*MockUserRepository)
    }{
        // тест-кейсы
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // выполнение теста
        })
    }
}
```

### 2. Мокирование зависимостей
```go
//go:generate mockgen -source=external.go -destination=./mock/external.go -package mock
type UserRepository interface {
    CreateUser(ctx context.Context, user *domain.User) error
}

// В тестах
ctrl := gomock.NewController(t)
defer ctrl.Finish()
repo := NewMockUserRepository(ctrl)
```

### 3. Изоляция тестов
- Каждый тест независим
- Моки для всех внешних зависимостей
- Очистка состояния после каждого теста

## Паттерны безопасности

### 1. Валидация входных данных
```go
func (s *UserService) validateCreateUser(input CreateUserInput) error {
    if input.Email == "" {
        return errors.New("email is required")
    }
    if !isValidEmail(input.Email) {
        return errors.New("invalid email format")
    }
    return nil
}
```

### 2. Хеширование паролей
```go
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}
```

### 3. CORS настройки
```go
app.Use(cors.New(cors.Config{
    AllowOrigins: cfg.CORSAllowedOrigins,
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
}))
```

## Паттерны развертывания

### 1. Graceful Shutdown
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

go func() {
    if err := app.Listen(":" + cfg.ServerPort); err != nil {
        logger.Error("Server failed to start", zap.Error(err))
    }
}()

sig := <-sigChan
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := app.ShutdownWithContext(ctx); err != nil {
    logger.Error("Server shutdown failed", zap.Error(err))
}
```

### 2. Конфигурация через переменные окружения
```go
type Config struct {
    ServerPort         string `env:"SERVER_PORT" envDefault:"8080"`
    DatabaseURL        string `env:"DATABASE_URL" envDefault:"sqlite://./data/bukhindor.db"`
    LogLevel           string `env:"LOG_LEVEL" envDefault:"info"`
    JWTSecret          string `env:"JWT_SECRET" envDefault:"your-secret-key"`
    CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"*"`
}
``` 