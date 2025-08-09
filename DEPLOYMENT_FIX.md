# Исправление проблемы зависания Docker build в GitHub Actions

## Проблема
GitHub Actions зависает на команде `docker build` более 15 минут без логов.

## Найденные причины

1. **Неправильная версия Go**: `golang:1.24-alpine` не существует
2. **CGO_ENABLED=1**: замедляет сборку и может вызывать зависания с Alpine
3. **Отсутствие оптимизаций**: нет кеширования слоев и зависимостей
4. **Избыточные пакеты**: gcc, musl-dev не нужны для вашего проекта

## Что исправлено

### 1. Основной Dockerfile
- ✅ Исправлена версия Go: `1.23-alpine`
- ✅ Убран CGO: `CGO_ENABLED=0`
- ✅ Добавлены оптимизации компиляции: `-ldflags="-w -s"`
- ✅ Убраны ненужные пакеты: gcc, musl-dev
- ✅ Добавлен GOPROXY для ускорения скачивания
- ✅ Исправлено копирование .env файла

### 2. Улучшенный .dockerignore
- ✅ Исключены файлы разработки
- ✅ Исключены скрипты и тесты
- ✅ Добавлены паттерны для профилирования

### 3. Альтернативные Dockerfile'ы
- `Dockerfile.simple` - для быстрого тестирования
- `Dockerfile.ci` - с максимальными оптимизациями

## Быстрое решение

### Вариант 1: Использовать исправленный Dockerfile
Текущий `Dockerfile` уже исправлен и должен работать быстрее.

### Вариант 2: Использовать простой Dockerfile
```bash
# В GitHub Actions замените команду на:
docker build -f Dockerfile.simple -t bukhindor-api:$TAG .
```

### Вариант 3: Оптимизированная команда с кешем
```bash
# В GitHub Actions добавьте BuildKit и кеширование:
export DOCKER_BUILDKIT=1
docker build \
  --cache-from bukhindor-api:latest \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  -t bukhindor-api:$TAG .
```

## Рекомендации для GitHub Actions

### 1. Добавьте тайм-ауты
```yaml
- name: Build Docker image
  timeout-minutes: 10  # Добавьте тайм-аут
  run: |
    docker build -t bukhindor-api:$TAG .
```

### 2. Включите детальное логирование
```yaml
- name: Build Docker image
  env:
    DOCKER_BUILDKIT: 1
    BUILDKIT_PROGRESS: plain  # Подробные логи
  run: |
    docker build -t bukhindor-api:$TAG .
```

### 3. Используйте кеширование
```yaml
- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3

- name: Build and push
  uses: docker/build-push-action@v5
  with:
    context: .
    push: false
    tags: bukhindor-api:${{ env.TAG }}
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

## Проверка локально

```bash
# Тестируйте сборку локально:
time docker build -t test-build .

# Или с простым Dockerfile:
time docker build -f Dockerfile.simple -t test-build .
```

## Ожидаемый результат

- ⏱️ Время сборки: 2-5 минут вместо 15+
- 📦 Размер образа: уменьшен за счет убранного CGO
- 🚀 Стабильность: нет зависаний
- 💾 Кеширование: слои будут переиспользоваться

## Если проблема продолжается

1. Проверьте ресурсы GitHub runner'а
2. Используйте `Dockerfile.simple` для диагностики
3. Добавьте мониторинг ресурсов в Action
4. Рассмотрите использование self-hosted runner'ов
