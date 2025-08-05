#!/bin/bash

# Скрипт для сборки Bukhindor API
set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# Проверяем наличие обязательных аргументов
if [ $# -lt 3 ]; then
    echo "Использование: $0 <JWT_SECRET> <POSTGRES_PASSWORD> <REDIS_PASSWORD> [TAG]"
    echo ""
    echo "Аргументы:"
    echo "  JWT_SECRET        - Секретный ключ для JWT токенов"
    echo "  POSTGRES_PASSWORD - Пароль для PostgreSQL"
    echo "  REDIS_PASSWORD    - Пароль для Redis (может быть пустым)"
    echo "  TAG              - Тег для Docker образа (опционально, по умолчанию 'latest')"
    echo ""
    echo "Пример:"
    echo "  $0 'my-super-secret-jwt-key' 'postgres-password' '' v1.0.0"
    exit 1
fi

# Получаем аргументы
JWT_SECRET="$1"
POSTGRES_PASSWORD="$2"
REDIS_PASSWORD="$3"
TAG="${4:-latest}"

# Проверяем, что секреты не пустые
if [ -z "$JWT_SECRET" ]; then
    error "JWT_SECRET не может быть пустым"
fi

if [ -z "$POSTGRES_PASSWORD" ]; then
    error "POSTGRES_PASSWORD не может быть пустым"
fi

log "Начинаем сборку Bukhindor API..."
log "JWT_SECRET: ${JWT_SECRET:0:10}..."
log "POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:0:10}..."
log "REDIS_PASSWORD: ${REDIS_PASSWORD:0:10}..."
log "TAG: $TAG"

# Проверяем наличие Docker
if ! command -v docker &> /dev/null; then
    error "Docker не установлен"
fi

# Проверяем, что Docker daemon запущен
if ! docker info &> /dev/null; then
    error "Docker daemon не запущен"
fi

# Создаем .env файл для docker-compose
log "Создаем .env файл..."
cat > .env << EOF
JWT_SECRET=$JWT_SECRET
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
REDIS_PASSWORD=$REDIS_PASSWORD
EOF

# Собираем Docker образ
log "Собираем Docker образ..."
docker build \
    --build-arg JWT_SECRET="$JWT_SECRET" \
    --build-arg POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
    --build-arg REDIS_PASSWORD="$REDIS_PASSWORD" \
    -t bukhindor-api:$TAG .

if [ $? -eq 0 ]; then
    log "Docker образ успешно собран: bukhindor-api:$TAG"
else
    error "Ошибка при сборке Docker образа"
fi

# Проверяем размер образа
IMAGE_SIZE=$(docker images bukhindor-api:$TAG --format "table {{.Size}}" | tail -n 1)
log "Размер образа: $IMAGE_SIZE"

# Создаем скрипт для запуска
log "Создаем скрипт запуска..."
cat > scripts/run.sh << 'EOF'
#!/bin/bash

# Скрипт для запуска Bukhindor API с Docker Compose

set -e

# Проверяем наличие .env файла
if [ ! -f .env ]; then
    echo "ERROR: Файл .env не найден. Запустите сначала build.sh"
    exit 1
fi

# Запускаем инфраструктуру
echo "Запускаем инфраструктуру..."
docker-compose up -d postgres pgpool redis prometheus grafana

# Ждем готовности сервисов
echo "Ждем готовности сервисов..."
sleep 30

# Применяем миграции
echo "Применяем миграции..."
docker-compose exec bukhindor-api ./main migrate up

# Запускаем API
echo "Запускаем API..."
docker-compose up -d bukhindor-api

echo "Bukhindor API запущен!"
echo "API: http://localhost:8080"
echo "Grafana: http://localhost:3000 (admin/admin)"
echo "Prometheus: http://localhost:9090"
EOF

chmod +x scripts/run.sh

# Создаем скрипт для остановки
cat > scripts/stop.sh << 'EOF'
#!/bin/bash

# Скрипт для остановки Bukhindor API

echo "Останавливаем Bukhindor API..."
docker-compose down

echo "Bukhindor API остановлен"
EOF

chmod +x scripts/stop.sh

# Создаем скрипт для просмотра логов
cat > scripts/logs.sh << 'EOF'
#!/bin/bash

# Скрипт для просмотра логов Bukhindor API

if [ -z "$1" ]; then
    echo "Просмотр логов всех сервисов..."
    docker-compose logs -f
else
    echo "Просмотр логов сервиса: $1"
    docker-compose logs -f "$1"
fi
EOF

chmod +x scripts/logs.sh

log "Сборка завершена успешно!"
log ""
log "Для запуска используйте:"
log "  ./scripts/run.sh"
log ""
log "Для остановки используйте:"
log "  ./scripts/stop.sh"
log ""
log "Для просмотра логов используйте:"
log "  ./scripts/logs.sh [service_name]"
log ""
log "Доступные сервисы:"
log "  - API: http://localhost:8080"
log "  - Grafana: http://localhost:3000 (admin/admin)"
log "  - Prometheus: http://localhost:9090" 