# Changelog

Все значимые изменения этого проекта будут документироваться в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.0.0/).

## [Unreleased]

### Added
- Развертывание в CI/CD Dev окружения

### Planned
- Прогон интеграционных тестов в CI/CD

---

## [v1.0.0] — 2025-05-17

### Added
- Базовая реализация микросервиса SSO на Go.
- gRPC API: регистрация, вход, получение ключа подписи.
- Интеграция с PostgreSQL для хранения пользователей и приложений.
- Интеграция с HashiCorp Vault для хранения секретов приложений.
- Миграции базы данных через golang-migrate.
- Конфигурация через YAML и переменные окружения (`cleanenv`).
- Логирование на базе zap.
- Функциональные тесты gRPC-сервиса.
- CI/CD: GitHub Actions для тестирования и деплоя.

---

[Unreleased]: https://github.com/passwordhash/go-sso/compare/v1.0.0...HEAD
[v1.0.0]: https://github.com/passwordhash/go-sso/releases/tag/v1.0.0
