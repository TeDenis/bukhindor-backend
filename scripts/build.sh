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

# Собираем Docker образ
log "Собираем Docker образ..."
docker build -t bukhindor-api:$TAG .

if [ $? -eq 0 ]; then
    log "Docker образ успешно собран: bukhindor-api:$TAG"
else
    error "Ошибка при сборке Docker образа"
fi

# Проверяем размер образа
IMAGE_SIZE=$(docker images bukhindor-api:$TAG --format "table {{.Size}}" | tail -n 1)
log "Размер образа: $IMAGE_SIZE"

log "Создаю/обновляю скрипт деплоя..."
cat > scripts/deploy.sh << 'EOF'
#!/bin/bash
set -e

IMAGE_TAG="$1"
if [ -z "$IMAGE_TAG" ]; then
  echo "Usage: $0 <image-tag>" && exit 1
fi

APP_NAME="bukhindor-api"
CONTAINER_NAME="$APP_NAME"

echo "Stopping old container (if exists)..."
docker rm -f "$CONTAINER_NAME" 2>/dev/null || true

echo "Starting new container..."
docker run -d --restart unless-stopped \
  --name "$CONTAINER_NAME" \
  -p 8080:8080 -p 9091:9091 \
  -e SERVER_PORT=8080 \
  -e SERVER_HOST=0.0.0.0 \
  -e POSTGRES_HOST=${POSTGRES_HOST:-localhost} \
  -e POSTGRES_PORT=${POSTGRES_PORT:-5432} \
  -e POSTGRES_USER=${POSTGRES_USER:-bukhindor} \
  -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-password} \
  -e POSTGRES_DB=${POSTGRES_DB:-bukhindor} \
  -e POSTGRES_SSLMODE=${POSTGRES_SSLMODE:-disable} \
  -e REDIS_URL=${REDIS_URL:-redis://localhost:6379} \
  -e REDIS_PASSWORD=${REDIS_PASSWORD:-} \
  -e REDIS_DB=${REDIS_DB:-0} \
  -e JWT_SECRET=${JWT_SECRET:-your-secret-key} \
  -e LOG_LEVEL=${LOG_LEVEL:-info} \
  -e LOG_FORMAT=${LOG_FORMAT:-json} \
  -e METRICS_PORT=${METRICS_PORT:-9091} \
  bukhindor-api:"$IMAGE_TAG"

echo "Deployed bukhindor-api:$IMAGE_TAG"
EOF

chmod +x scripts/deploy.sh

log "Сборка завершена. Для деплоя: ./scripts/deploy.sh $TAG"