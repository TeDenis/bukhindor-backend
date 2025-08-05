# Tech Context - Bukhindor Backend

## Технологический стек

### Основные технологии
- **Язык программирования**: Go 1.21+
- **HTTP Framework**: Fiber v2.52.0
- **База данных**: SQLite (разработка), PostgreSQL (продакшен)
- **Query Builder**: Squirrel (планируется)
- **Миграции**: Goose v3 (планируется)
- **Логирование**: Zap v1.26.0
- **CLI**: Cobra v1.8.0
- **Конфигурация**: godotenv v1.5.1

### Зависимости проекта

#### Основные зависимости
```go
require (
    github.com/gofiber/fiber/v2 v2.52.0    // HTTP framework
    github.com/joho/godotenv v1.5.1        // Загрузка .env файлов
    github.com/spf13/cobra v1.8.0          // CLI framework
    go.uber.org/zap v1.26.0                // Логирование
)
```

#### Косвенные зависимости
```go
require (
    github.com/andybalholm/brotli v1.0.6           // Сжатие
    github.com/google/uuid v1.5.0                  // UUID генерация
    github.com/klauspost/compress v1.17.2          // Сжатие
    github.com/mattn/go-colorable v0.1.13          // Цветной вывод
    github.com/mattn/go-isatty v0.0.20             // TTY определение
    github.com/mattn/go-runewidth v0.0.15          // Ширина символов
    github.com/rivo/uniseg v0.2.0                  // Unicode
    github.com/spf13/pflag v1.0.5                  // Флаги CLI
    github.com/stretchr/testify v1.8.3             // Тестирование
    github.com/valyala/bytebufferpool v1.0.0       // Пул буферов
    github.com/valyala/fasthttp v1.51.0            // Fast HTTP
    github.com/valyala/tcplisten v1.0.0            // TCP listener
    go.uber.org/multierr v1.11.0                   // Множественные ошибки
    golang.org/x/sys v0.15.0                       // Системные вызовы
)
```

## Структура проекта

### Корневая структура
```
bukhindor-backend/
├── cmd/                    # Точки входа
│   ├── api/               # API сервер
│   │   └── api.go         # Главный файл API
│   └── cli/               # CLI инструменты
│       └── cli.go         # Главный файл CLI
├── internal/              # Внутренний код
│   ├── app/               # Общие компоненты
│   │   └── const.go       # Константы
│   ├── config/            # Конфигурация
│   │   └── config.go      # Настройки приложения
│   ├── domain/            # Доменные модели
│   │   └── user.go        # Модель пользователя
│   └── web/               # Web слой
│       ├── api/           # API endpoints
│       │   └── service.go # API сервис
│       └── server/        # HTTP сервер
│           └── server.go  # Сервер
├── deployments/           # Развертывание
│   └── sqlite/            # SQLite конфигурация
│       └── migrations/    # Миграции БД
│           └── 00001_init.sql
├── build/                 # Собранные бинарники
├── data/                  # Данные (SQLite файлы)
├── .env                   # Переменные окружения
├── env.example            # Пример .env файла
├── go.mod                 # Go модули
├── go.sum                 # Хеши зависимостей
├── Makefile               # Команды сборки
├── README.md              # Документация
└── .gitignore             # Игнорируемые файлы
```

### Планируемая структура (после развития)
```
bukhindor-backend/
├── cmd/
│   ├── api/
│   └── cli/
├── internal/
│   ├── app/
│   │   ├── const.go
│   │   ├── funcs.go       # Общие функции
│   │   └── errors.go      # Тексты ошибок
│   ├── adapters/
│   │   └── storage/       # Адаптеры БД
│   │       ├── external.go
│   │       ├── service.go
│   │       └── sqlite.go
│   ├── config/
│   ├── domain/
│   │   ├── user.go
│   │   └── session.go     # Модель сессии
│   ├── service/
│   │   ├── users/         # Сервис пользователей
│   │   │   ├── external.go
│   │   │   ├── service.go
│   │   │   ├── users.go
│   │   │   └── validation.go
│   │   ├── auth/          # Сервис аутентификации
│   │   │   ├── external.go
│   │   │   ├── service.go
│   │   │   ├── auth.go
│   │   │   └── jwt.go
│   │   └── crypto/        # Криптографические функции
│   │       ├── service.go
│   │       └── password.go
│   └── web/
│       ├── api/
│       │   ├── external.go
│       │   ├── service.go
│       │   ├── handlers.go
│       │   └── routing.go
│       └── server/
├── deployments/
│   ├── sqlite/
│   │   └── migrations/
│   └── postgresql/
│       └── migrations/
├── docs/                  # Документация
├── swag/                  # Swagger документация
└── tests/                 # Интеграционные тесты
```

## Конфигурация

### Переменные окружения
```bash
# Сервер
SERVER_PORT=8080
SERVER_HOST=localhost

# База данных
DATABASE_URL=sqlite://./data/bukhindor.db

# Логирование
LOG_LEVEL=info          # debug, info, warn, error
LOG_FORMAT=console      # console, json

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# CORS
CORS_ALLOWED_ORIGINS=*
```

### Структура конфигурации
```go
type Config struct {
    // Сервер
    ServerPort string `env:"SERVER_PORT" envDefault:"8080"`
    ServerHost string `env:"SERVER_HOST" envDefault:"localhost"`

    // База данных
    DatabaseURL string `env:"DATABASE_URL" envDefault:"sqlite://./data/bukhindor.db"`

    // Логирование
    LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
    LogFormat string `env:"LOG_FORMAT" envDefault:"console"`

    // JWT
    JWTSecret string `env:"JWT_SECRET" envDefault:"your-secret-key"`

    // CORS
    CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"*"`
}
```

## Команды сборки

### Makefile команды
```makefile
# Основные команды
build          # Сборка приложения
run            # Запуск приложения
test           # Запуск тестов
lint           # Проверка линтером
clean          # Очистка артефактов

# Миграции
migrate-up     # Применить миграции
migrate-down   # Откатить миграции
migrate-status # Статус миграций

# Разработка
deps           # Установка зависимостей
generate       # Генерация моков
dev            # Полный цикл разработки
start          # Запуск с миграциями
```

### CLI команды
```bash
# Миграции
bukhindor-cli migrate up      # Применить миграции
bukhindor-cli migrate down    # Откатить миграции
bukhindor-cli migrate status  # Статус миграций
```

## База данных

### Схема БД (текущая)
```sql
-- Таблица пользователей
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    first_name TEXT,
    last_name TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    role TEXT DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица сессий
CREATE TABLE user_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Индексы
```sql
-- Пользователи
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Сессии
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token_hash ON user_sessions(token_hash);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
```

## API Endpoints

### Текущие endpoints
```
GET    /health                    # Health check
GET    /api/v1/users             # Список пользователей
POST   /api/v1/users             # Создать пользователя
GET    /api/v1/users/:id         # Получить пользователя
PUT    /api/v1/users/:id         # Обновить пользователя
DELETE /api/v1/users/:id         # Удалить пользователя
POST   /api/v1/auth/login        # Вход в систему
POST   /api/v1/auth/register     # Регистрация
POST   /api/v1/auth/logout       # Выход из системы
GET    /api/v1/auth/me           # Текущий пользователь
```

### Планируемые endpoints
```
# Дополнительные аутентификации
POST   /api/v1/auth/refresh      # Обновление токена
POST   /api/v1/auth/forgot-password # Восстановление пароля
POST   /api/v1/auth/reset-password  # Сброс пароля

# Управление профилем
PUT    /api/v1/auth/profile      # Обновление профиля
POST   /api/v1/auth/change-password # Смена пароля

# Административные
GET    /api/v1/admin/users       # Список пользователей (админ)
PUT    /api/v1/admin/users/:id   # Обновление пользователя (админ)
DELETE /api/v1/admin/users/:id   # Удаление пользователя (админ)
```

## Технические ограничения

### Производительность
- **Время отклика API**: < 100ms для 95% запросов
- **Память**: < 100MB для базовой функциональности
- **CPU**: Эффективное использование ресурсов

### Безопасность
- **Пароли**: Хеширование с bcrypt
- **JWT**: Подпись секретным ключом
- **CORS**: Настройки для Flutter приложения
- **Валидация**: Проверка всех входных данных

### Масштабируемость
- **Архитектура**: Готова к горизонтальному масштабированию
- **База данных**: Поддержка PostgreSQL для продакшена
- **Кэширование**: Возможность добавления Redis
- **Микросервисы**: Архитектура позволяет разделение на сервисы

## Инструменты разработки

### Линтеры и форматтеры
- **golangci-lint**: Основной линтер
- **go fmt**: Форматирование кода
- **go vet**: Статический анализ

### Тестирование
- **go test**: Базовое тестирование
- **testify**: Assertions и моки
- **gomock**: Генерация моков
- **sqlmock**: Мокирование БД
- **gock**: HTTP мокирование

### Документация
- **Swagger**: API документация
- **godoc**: Документация кода
- **README.md**: Основная документация

### CI/CD (планируется)
- **GitHub Actions**: Автоматизация
- **Docker**: Контейнеризация
- **Docker Compose**: Локальная разработка 