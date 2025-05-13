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

## Полезные команды

- Сборка и запуск приложения локально:
  ```bash
  go run cmd/sso/main.go -config=config/local.yml
  ```
- Запуск мигратора:
  ```bash
  go run cmd/migrator/main.go -config=config/local.yml
  ```
- Выполнение функциональных тестов:
  ```bash
  go test ./tests/...
  ```
- Генерация JWT-секрета вручную:
  ```go
  secret, _ := jwt.GenerateHS256Secret()
  ```

## Схема базы данных
- Таблица `users`:
  - `uuid UUID PK`, `email TEXT UNIQUE`, `pass_hash BYTEA`, `is_admin BOOLEAN`.
- Таблица `apps`:
  - `id SERIAL PK`, `name TEXT UNIQUE`, `secret TEXT UNIQUE`.

## Лицензия
MIT License.