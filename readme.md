# go-sso

## Описание
`go-sso` — это микросервис единой системы аутентификации (SSO) на базе gRPC, PostgreSQL и HashiCorp Vault. Позволяет регистрировать пользователей, выполнять вход и выдавать JWT-токены для отдельных клиентских приложений.

## Основные компоненты
- **gRPC API** для взаимодействия клиентов: регистрация, вход, получение ключа подписи.
- **PostgreSQL** для хранения данных о пользователях и приложениях.
- **HashiCorp Vault** для безопасного хранения и генерации секретных ключей приложений.
- **Миграции БД** реализованы через библиотеку [golang-migrate](https://github.com/golang-migrate/migrate).
- **Конфигурация** через YAML-файлы + переменные окружения с помощью `cleanenv`.
- **Логирование** на базе `zap` от Uber.
- **Тестирование**: функциональные тесты gRPC-сервиса.

## Репозиторий proto
Схема gRPC и сообщения описаны в репозитории:
https://github.com/passwordhash/protos (модуль `github.com/passwordhash/protos/gen/go/go-sso`)

## Установка и запуск

### Зависимости
- Go 1.24+
- Docker и Docker Compose
- PostgreSQL 15+
- Vault (dev-режим для локальной разработки)

### Локальный запуск через Docker Compose
1. Скопируйте файл `.env.example` в `.env` и заполните значения:
   ```bash
   cp .env.example .env
   ```
2. Запустите контейнеры:
   ```bash
   docker compose -f docker-compose.local.yml up --build
   ```
3. В контейнере `app` будет запущен gRPC-сервер на порту, заданном в `GRPC_PORT`.

### Конфигурация
Файл конфигурации по умолчанию лежит в `config/local.yml`.
Основные параметры:
- `env` — среда исполнения (`local`, `dev`, `prod`).
- `grpc.host`, `grpc.port`, `grpc.timeout` — настройки gRPC-сервера.
- `psql.host`, `psql.port`, `psql.user`, `psql.pass`, `psql.db` — подключение к PostgreSQL.
- `vault.addr`, `vault.token`, `vault.timeout` — Vault-клиент.

Путь до файла в контейнере передаётся через переменную `CONFIG_PATH`.

### Миграции базы данных
Миграции хранятся в `migrations/`.
Применение миграций вручную:
```bash
go run cmd/migrator/main.go -config=config/local.yml
```
Или через Docker Compose сервис `migrate` (авто-запуск после поднятия БД).

## API gRPC

### Сервисы
- `Register(RegisterRequest) -> RegisterResponse`
  Регистрирует нового пользователя.
  Параметры: `email`, `password`.
  Возвращает: `user_uuid`.

- `Login(LoginRequest) -> LoginResponse`
  Аутентификация пользователя и выдача JWT.
  Параметры: `email`, `password`, `app_id`.
  Возвращает: `token`.

- `SigningKey(SigningKeyRequest) -> SigningKeyResponse`
  Получение/генерация секретного ключа для приложения.
  Параметр: `app_name`.
  Возвращает: `signing_key`.

### Пример запроса gRPC (Go-клиент)
```go
conn, _ := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := go-sso.NewAuthClient(conn)

resp, err := client.Register(ctx, &RegisterRequest{Email:"user@example.com", Password:"pass"})
```

## Тестирование
Функциональные тесты в папке `tests/`:
```bash
go test -v ./tests/...
```
- Перед тестами используется свой набор миграций `tests/migrations`.
- Конфиг для тестов: `config/test.yml`.

---

## CI/CD и миграции базы данных

### GitHub Workflows

В проекте настроены два workflow:

- **CI (`.github/workflows/ci.yml`)** — автоматический запуск юнит- и функциональных тестов при каждом пуше и pull request в ветки `master` и `develop`, а также при создании тегов.
- **CD (`.github/workflows/cd.yml`)** — деплой на продакшн при пуше тега вида `v*.*.*`. Включает сборку Docker-образа, пуш в Docker Hub, копирование конфигов на сервер и запуск контейнера.

### Миграции базы данных в CD

Миграции для продакшн-базы выполняются через официальный Docker-образ [migrate/migrate](https://github.com/golang-migrate/migrate):

**Пример команды для ручного запуска:**
```bash
docker run --rm \
  -v $(pwd)/migrations:/migrations \
  migrate/migrate \
  -path=/migrations \
  -database "postgres://<USER>:<PASSWORD>@<HOST>:<PORT>/<DBNAME>?sslmode=disable" \
  up
```

- Все параметры подключения к базе должны храниться в GitHub Secrets.
