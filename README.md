# PR Reviewer Assignment Service

[![forthebadge](https://forthebadge.com/images/badges/made-with-go.svg)](https://forthebadge.com) [![forthebadge](http://forthebadge.com/images/badges/built-with-love.svg)](http://forthebadge.com)

Микросервис для автоматического назначения ревьюеров на Pull Request'ы с возможностью управления командами и участниками.

Сервис автоматически назначает до двух активных ревьюеров из команды автора PR при его создании, позволяет выполнять переназначение и получать список PR'ов конкретного пользователя. После merge PR изменение состава ревьюверов запрещено.

## Используемые технологии

- **Go** (язык разработки) с веб-фреймворком **Gin**
- **PostgreSQL** (хранилище данных)
- **Docker** и **Docker Compose** (для контейнеризации и запуска сервиса)
- **golang-migrate/migrate** (для миграций базы данных)
- **sqlx** (для работы с PostgreSQL)
- **go-playground/validator** (для валидации запросов)
- **testify** и **testcontainers-go** (для тестирования)

## Реализованные дополнительные задания

- ✅ Эндпоинт статистики `/users/assignments/stat` для получения количества назначений по пользователям
- ✅ Массовая деактивация пользователей команды через `/team/deactivateMembers` с автоматическим переназначением открытых PR
- ✅ Интеграционное тестирование с использованием testcontainers для тестирования слоя репозитория
- ✅ Конфигурация линтера golangci-lint

## Getting Started

Для запуска сервиса необходимо:

1. Создать `.env` файл на основе `.env.example` и заполнить параметры подключения к базе данных:

```bash
cp .env.example .env
```

2. Убедиться, что в `.env` файле заполнены все необходимые переменные:
    - `APP_PORT` - порт для запуска сервиса (по умолчанию 8080)
    - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - параметры подключения к PostgreSQL
    - `DB_SSLMODE` - режим SSL для подключения к БД
    - `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_MAX_LIFETIME` - настройки пула соединений

3. Миграции базы данных применяются автоматически при запуске через docker-compose

## Usage

### Запуск сервиса

Запустить сервис можно с помощью команды:

```bash
make compose-up
```

После запуска сервис будет доступен по адресу `http://localhost:8080`. Все запросы начинаются с `http://localhost:8080/api/v1`.

API документация доступна в файле `docs/openapi.yml`.

### Остановка сервиса

Для остановки без удаления volumes:

```bash
make compose-down
```

Для полной очистки (включая volumes):

```bash
make docker-rm-volume
```

### Тестирование

Запуск unit-тестов и интеграционных тестов:

```bash
make test
```

Получение coverage в консоли:

```bash
make cover
```

Генерация HTML отчета по покрытию:

```bash
make cover-html
```

### Линтинг

Для запуска линтера golangci-lint:

```bash
make linter-golangci
```


### Миграции

Создание новой миграции:

```bash
make migrate-create name=migration_name
```

Применение всех pending миграций:

```bash
make migrate-up
```

Откат всех миграций:

```bash
make migrate-down
```

## API Examples

### Оглавление

- [Health](#health)
    - [Проверка состояния сервиса](#проверка-состояния-сервиса)
- [Teams](#teams)
    - [Создать команду](#создать-команду)
    - [Получить команду](#получить-команду)
    - [Деактивировать участников команды](#деактивировать-участников-команды)
- [Users](#users)
    - [Установить активность пользователя](#установить-активность-пользователя)
    - [Получить PR’ы назначенные пользователю](#получить-prы-назначенные-пользователю)
    - [Получить статистику назначений](#получить-статистику-назначений)
- [Pull Requests](#pull-requests)
    - [Создать PR](#создать-pr)
    - [Пометить PR как MERGED](#пометить-pr-как-merged)
    - [Переназначить ревьювера](#переназначить-ревьювера)

---

### Health

#### Проверка состояния сервиса

### Запрос

```bash
curl --location --request GET 'http://localhost:8080/api/v1/health'
```

### Пример ответа:

```json
{
  "status": "ok"
}
```

---

### Teams

#### Создать команду с участниками

### Запрос

```bash
curl --location --request POST 'http://localhost:8080/api/v1/team/add' \
--header 'Content-Type: application/json' \
--data-raw '{
  "team_name": "payments",
  "members": [
    { "user_id": "u1", "username": "Alice", "is_active": true },
    { "user_id": "u2", "username": "Bob", "is_active": true }
  ]
}'
```

### Пример ответа:

```json
{
  "team": {
    "team_name": "backend",
    "members": [
      { "user_id": "u1", "username": "Alice", "is_active": true },
      { "user_id": "u2", "username": "Bob", "is_active": true }
    ]
  }
}
```

---

### Получить команду

### Запрос

```bash
curl --location --request GET 'http://localhost:8080/api/v1/team/get?team_name=backend'
```

### Пример ответа:

```json
{
  "team_name": "backend",
  "members": [
    { "user_id": "u1", "username": "Alice", "is_active": true },
    { "user_id": "u2", "username": "Bob", "is_active": true }
  ]
}
```

---

### Деактивировать участников команды

### Запрос

```bash
curl --location --request POST 'http://localhost:8080/api/v1/team/deactivateMembers' \
--header 'Content-Type: application/json' \
--data-raw '{
  "team_name": "backend",
  "user_ids": ["u2", "u3"]
}'
```

### Пример ответа:

*(204 No Content — без тела)*

---

### Users

#### Установить активность пользователя

### Запрос

```bash
curl --location --request POST 'http://localhost:8080/api/v1/users/setIsActive' \
--header 'Content-Type: application/json' \
--data-raw '{
  "user_id": "u2",
  "is_active": false
}'
```

### Пример ответа:

```json
{
  "user": {
    "user_id": "u2",
    "username": "Bob",
    "team_name": "backend",
    "is_active": false
  }
}
```

---

### Получить PR’ы назначенные пользователю

### Запрос

```bash
curl --location --request GET 'http://localhost:8080/api/v1/users/getReview?user_id=u2'
```

### Пример ответа:

```json
{
  "user_id": "u2",
  "pull_requests": [
    {
      "pull_request_id": "pr-1001",
      "pull_request_name": "Add search",
      "author_id": "u1",
      "status": "OPEN"
    }
  ]
}
```

---

### Получить статистику назначений

### Запрос

```bash
curl --location --request GET 'http://localhost:8080/api/v1/users/assignments/stat?user_id=u2'
```

### Пример ответа:

```json
{
  "user_id": "u2",
  "assignments_count": 5
}
```

---

### Pull Requests

#### Создать PR

### Запрос

```bash
curl --location --request POST 'http://localhost:8080/api/v1/pullRequest/create' \
--header 'Content-Type: application/json' \
--data-raw '{
  "pull_request_id": "pr-1001",
  "pull_request_name": "Add search",
  "author_id": "u1"
}'
```

### Пример ответа:

```json
{
  "pr": {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search",
    "author_id": "u1",
    "status": "OPEN",
    "assigned_reviewers": ["u2", "u3"]
  }
}
```

---

#### Пометить PR как MERGED

### Запрос

```bash
curl --location --request POST 'http://localhost:8080/api/v1/pullRequest/merge' \
--header 'Content-Type: application/json' \
--data-raw '{
  "pull_request_id": "pr-1001"
}'
```

### Пример ответа:

```json
{
  "pr": {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search",
    "author_id": "u1",
    "status": "MERGED",
    "assigned_reviewers": ["u2", "u3"],
    "mergedAt": "2025-10-24T12:34:56Z"
  }
}
```

---

#### Переназначить ревьювера

### Запрос

```bash
curl --location --request POST 'http://localhost:8080/api/v1/pullRequest/reassign' \
--header 'Content-Type: application/json' \
--data-raw '{
  "pull_request_id": "pr-1001",
  "old_user_id": "u2"
}'
```

### Пример ответа:

```json
{
  "pr": {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search",
    "author_id": "u1",
    "status": "OPEN",
    "assigned_reviewers": ["u3", "u5"]
  },
  "replaced_by": "u5"
}
```

