# Bukhindor Backend

REST API для Flutter приложения Bukhindor с поддержкой аутентификации, PostgreSQL, Redis и мониторинга.

## 🚀 Быстрый старт

### Требования

- Docker и Docker Compose
- Go 1.24+ (для локальной разработки)

### Развертывание с Docker

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/TeDenis/bukhindor-backend.git
cd bukhindor-backend
```

2. **Соберите и запустите:**
```bash
# Сборка с секретами
./scripts/build.sh "your-jwt-secret-key" "postgres-password" "redis-password" v1.0.0

# Запуск инфраструктуры
./scripts/run.sh
```

3. **Проверьте работу:**
```bash
# Health check
curl http://localhost:8080/health

# API документация
open http://localhost:8080/docs
```

### Доступные сервисы

- **API**: http://localhost:8080
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## 🏗️ Архитектура

### Технологический стек

- **Язык**: Go 1.24
- **HTTP Framework**: Fiber v2
- **База данных**: PostgreSQL 15 + PgPool-II
- **Кеш**: Redis 7
- **Аутентификация**: JWT токены
- **Мониторинг**: Prometheus + Grafana
- **Контейнеризация**: Docker + Docker Compose

### Архитектурные принципы

- **Чистая гексагональная архитектура**
- **Разделение слоев**: web → service → adapters → domain
- **Зависимости через интерфейсы**
- **Слабая связанность компонентов**

## 📁 Структура проекта

```
bukhindor-backend/
├── cmd/                    # Точки входа
│   ├── api/               # API сервер
│   └── cli/               # CLI инструменты
├── deployments/           # Конфигурация развертывания
│   ├── postgres/          # Миграции PostgreSQL
│   └── monitoring/        # Конфигурация мониторинга
├── internal/              # Внутренний код
│   ├── adapters/          # Адаптеры (БД, Redis)
│   ├── config/            # Конфигурация
│   ├── domain/            # Доменные модели
│   ├── monitoring/        # Метрики Prometheus
│   ├── service/           # Бизнес-логика
│   └── web/               # HTTP слой
├── scripts/               # Скрипты развертывания
├── docs/                  # Документация
│   └── openapi.yaml       # OpenAPI спецификация
├── docker-compose.yml     # Docker Compose
├── Dockerfile            # Docker образ
└── README.md             # Документация
```

## 🔐 API Endpoints

### Аутентификация

| Метод | Endpoint | Описание | Авторизация |
|-------|----------|----------|-------------|
| POST | `/api/v1/auth/register` | Регистрация пользователя | ❌ |
| POST | `/api/v1/auth/login` | Вход в систему | ❌ |
| POST | `/api/v1/auth/refresh` | Обновление токенов | ❌ |
| POST | `/api/v1/auth/reset-password` | Сброс пароля | ❌ |
| GET | `/api/v1/auth/me` | Информация о пользователе | ✅ |

### Пользователи

| Метод | Endpoint | Описание | Авторизация |
|-------|----------|----------|-------------|
| GET | `/api/v1/users` | Список пользователей | ✅ |
| POST | `/api/v1/users` | Создание пользователя | ✅ |
| GET | `/api/v1/users/{id}` | Получение пользователя | ✅ |
| PUT | `/api/v1/users/{id}` | Обновление пользователя | ✅ |
| DELETE | `/api/v1/users/{id}` | Удаление пользователя | ✅ |

### Система

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus метрики |

## 🔧 Конфигурация

### Переменные окружения

```bash
# Сервер
SERVER_PORT=8080
SERVER_HOST=localhost

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=bukhindor
POSTGRES_PASSWORD=password
POSTGRES_DB=bukhindor

# PgPool
PGPOOL_HOST=localhost
PGPOOL_PORT=5432
PGPOOL_USER=bukhindor
PGPOOL_PASSWORD=password

# Redis
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION=15m
REFRESH_TOKEN_EXPIRATION=7d

# Логирование
LOG_LEVEL=info
LOG_FORMAT=json

# Метрики
METRICS_PORT=9091
```

### Обязательные заголовки

Все API запросы должны содержать:

- `X-App-Version`: Версия приложения
- `X-App-Type`: Тип приложения (ios/android/web)
- `X-Device-ID`: Идентификатор устройства

## 🧪 Тестирование

### Запуск тестов

```bash
# Unit тесты
go test ./...

# Integration тесты
go test -tags=integration ./...

# Покрытие кода
go test -cover ./...
```

### API тестирование

```bash
# Запуск тестового скрипта
./test_api.sh
```

## 📊 Мониторинг

### Метрики Prometheus

- **HTTP метрики**: Запросы, время ответа, ошибки
- **Бизнес метрики**: Регистрации, входы, сбросы паролей
- **Системные метрики**: Сессии, подключения к БД/Redis

### Grafana дашборды

- **API метрики**: HTTP запросы и производительность
- **Бизнес метрики**: Активность пользователей
- **Системные метрики**: Состояние инфраструктуры

## 🚀 Развертывание

### Продакшен

1. **Настройте секреты:**
```bash
export JWT_SECRET="your-super-secret-jwt-key"
export POSTGRES_PASSWORD="strong-postgres-password"
export REDIS_PASSWORD="strong-redis-password"
```

2. **Соберите образ:**
```bash
./scripts/build.sh "$JWT_SECRET" "$POSTGRES_PASSWORD" "$REDIS_PASSWORD" v1.0.0
```

3. **Запустите:**
```bash
./scripts/run.sh
```

### Локальная разработка

```bash
# Установите зависимости
go mod download

# Запустите инфраструктуру
docker-compose up -d postgres pgpool redis

# Примените миграции
go run cmd/cli/cli.go migrate up

# Запустите API
go run cmd/api/api.go
```

## 🔍 Отладка

### Логи

```bash
# Просмотр логов всех сервисов
./scripts/logs.sh

# Просмотр логов конкретного сервиса
./scripts/logs.sh bukhindor-api
```

### База данных

```bash
# Подключение к PostgreSQL
docker-compose exec postgres psql -U bukhindor -d bukhindor

# Просмотр миграций
go run cmd/cli/cli.go migrate status
```

## 📝 Разработка

### Добавление новых endpoints

1. Создайте handler в `internal/web/api/`
2. Добавьте роут в `SetupRoutes()`
3. Обновите OpenAPI спецификацию
4. Напишите тесты

### Добавление новых метрик

1. Добавьте метрику в `internal/monitoring/metrics.go`
2. Записывайте метрику в бизнес-логике
3. Обновите Grafana дашборд

### Миграции базы данных

```bash
# Создание новой миграции
touch deployments/postgres/migrations/00003_new_feature.sql

# Применение миграций
go run cmd/cli/cli.go migrate up

# Откат миграции
go run cmd/cli/cli.go migrate down
```

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature branch
3. Внесите изменения
4. Добавьте тесты
5. Создайте Pull Request

## 📄 Лицензия

MIT License

## 🆘 Поддержка

- **Issues**: GitHub Issues
- **Email**: support@bukhindor.com
- **Документация**: `/docs/openapi.yaml` 