# Сервис назначения ревьюеров для Pull Request

Микросервис для автоматического назначения ревьюеров на Pull Request'ы и управления командами разработки.

## Описание

Сервис предоставляет HTTP API для:
- Автоматического назначения ревьюеров на PR из команды автора
- Управления командами и пользователями
- Переназначения ревьюеров
- Отслеживания статуса PR (OPEN/MERGED)

## Технологический стек

- **Backend**: Go
- **База данных**: PostgreSQL
- **Контейнеризация**: Docker + Docker Compose

## Предварительные требования

- Docker и Docker Compose
- Go 1.21+ (для локального запуска тестов)

## Запуск проекта

```bash
docker compose up -d --build
```

## Остановка проекта

```bash
docker compose down -v
```

## Запуск тестов

```bash
docker-compose up -d test-db
go test ./tests/... -v -count=1
```

При завершении тестов может потребоваться нажать Ctrl+C (для ОС Windows)

## Основные команды
- **Запуск**: docker compose up -d api postgres --build

- **Остановка**: docker compose down -v