#!/bin/bash

# Быстрый скрипт деплоя с оптимизациями и таймаутами
set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция логирования
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
    exit 1
}

warn() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

# Проверка аргументов
if [ $# -eq 0 ]; then
    error "Использование: $0 <dockerfile_type> [tag]
    
Доступные типы Dockerfile:
    fast    - Ультра-быстрый single-stage (30-40 сек)
    deploy  - Оптимизированный multi-stage (50-60 сек)
    main    - Основной Dockerfile (60-80 сек)
    
Пример: $0 fast v1.0.0"
fi

DOCKERFILE_TYPE=$1
TAG=${2:-$(git rev-parse --short HEAD)}

# Выбор Dockerfile
case $DOCKERFILE_TYPE in
    "fast")
        DOCKERFILE="Dockerfile.fast"
        TIMEOUT=60
        ;;
    "deploy")
        DOCKERFILE="Dockerfile.deploy"
        TIMEOUT=90
        ;;
    "main")
        DOCKERFILE="Dockerfile"
        TIMEOUT=120
        ;;
    *)
        error "Неизвестный тип Dockerfile: $DOCKERFILE_TYPE"
        ;;
esac

log "Начинаем деплой с настройками:"
log "  Dockerfile: $DOCKERFILE"
log "  Таймаут: $TIMEOUT секунд"
log "  Тег: $TAG"

# Проверяем существование Dockerfile
if [ ! -f "$DOCKERFILE" ]; then
    error "Файл $DOCKERFILE не найден!"
fi

# Проверяем версию Go в go.mod
GO_VERSION=$(grep "^go " go.mod | awk '{print $2}')
log "Версия Go в проекте: $GO_VERSION"

if [[ $GO_VERSION > "1.23" ]]; then
    warn "Версия Go $GO_VERSION может быть недоступна в Docker образе!"
fi

# Функция сборки с таймаутом
build_with_timeout() {
    log "Запускаем сборку Docker образа..."
    
    # Используем timeout для ограничения времени сборки
    if timeout $TIMEOUT docker build -f "$DOCKERFILE" -t "bukhindor-api:$TAG" . ; then
        log "✅ Сборка успешно завершена за $(($TIMEOUT - $(ps -o etime= -p $! | tr -d ' '))) секунд"
        return 0
    else
        error "❌ Сборка не завершилась за $TIMEOUT секунд или завершилась с ошибкой"
    fi
}

# Проверка доступности Docker
if ! docker info > /dev/null 2>&1; then
    error "Docker недоступен. Убедитесь, что Docker запущен."
fi

# Очистка старых образов для экономии места
log "Очищаем старые образы..."
docker system prune -f > /dev/null 2>&1 || true

# Засекаем время
START_TIME=$(date +%s)

# Запускаем сборку
build_with_timeout

# Вычисляем время сборки
END_TIME=$(date +%s)
BUILD_TIME=$((END_TIME - START_TIME))

log "✅ Образ bukhindor-api:$TAG успешно собран за $BUILD_TIME секунд"

# Показываем размер образа
IMAGE_SIZE=$(docker images bukhindor-api:$TAG --format "{{.Size}}" | head -1)
log "📦 Размер образа: $IMAGE_SIZE"

# Тестируем образ
log "🧪 Тестируем образ..."
if docker run --rm -d --name test-container -p 8080:8080 "bukhindor-api:$TAG" > /dev/null; then
    sleep 2
    if docker exec test-container ls /usr/local/bin/ | grep -q api; then
        log "✅ Образ работает корректно"
        docker stop test-container > /dev/null
    else
        warn "⚠️  Не удалось найти бинарь API в образе"
        docker stop test-container > /dev/null
    fi
else
    warn "⚠️  Не удалось запустить тестовый контейнер"
fi

# Рекомендации
log ""
log "🚀 Деплой готов!"
log "Для запуска на сервере используйте:"
log "  docker run -d -p 8080:8080 --name bukhindor-api bukhindor-api:$TAG"
log ""
log "Для отправки на registry:"
log "  docker tag bukhindor-api:$TAG your-registry/bukhindor-api:$TAG"
log "  docker push your-registry/bukhindor-api:$TAG"

# Статистика
log ""
log "📊 Статистика сборки:"
log "  Dockerfile: $DOCKERFILE"
log "  Время сборки: $BUILD_TIME секунд"
log "  Размер образа: $IMAGE_SIZE"
log "  Тег: $TAG"
