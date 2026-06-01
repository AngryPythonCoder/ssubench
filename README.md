# SsuBench API

SsuBench — это REST API платформа для размещения заданий. Заказчики публикуют задачи, исполнители откликаются на них, а оплата производится виртуальными баллами.

## Стек технологий
- **Язык:** Go
- **СУБД:** PostgreSQL
- **Драйвер БД:** pgx v5
- **Миграции:** migrate & migrate CLI
- **Контейнеризация:** Docker & Docker Compose
- **Маршрутизация (routing):** chi
- **Валидация данных:** validator
- **Аутентификация:** JWT
- **Хеширование паролей:** bcrypt
- **Тестирование:** testcontainers-go & testify

## Функциональные возможности
1. **Ролевая модель:** Заказчик (Customer), Исполнитель (Performer), Администратор (Admin).
2. **Флоу задачи:** Публикация -> Отклик -> Выбор исполнителя -> Выполнение -> Подтверждение и Оплата.
3. **Атомарность:** Перевод баллов между пользователями выполняется в рамках одной БД-транзакции.
4. **Администрирование:** Блокировка/разблокировка пользователей, установка баланса.

## Инструкция по запуску

### 1. Подготовка окружения
Создайте файл `.env` в корневом каталоге и заполните его (пример):
```env
DB_URL=postgres://user:password@localhost:5432/ssubench?sslmode=disable
JWT_SECRET=super_secret_key
SERVER_PORT=8080
```

Создавать файл не обязательно, при его отсутствии приложение будет использовать конфигурацию по умолчанию.

### 2. Запуск базы данных
```bash
docker-compose up -d
```

### 3. Применение миграций
```bash
migrate -path migrations/ -database "postgres://user:password@localhost:5432/ssubench?sslmode=disable" up
```

### 4. Запуск приложения
```bash
go run main.go
```

## Тестирование
В проекте реализованы тесты с использованием `testcontainers-go`. Для запуска тестов убедитесь, что Docker запущен.

```bash
go test ./...
```

Или

```bash
go test ./tests/integration
```

*Будет запущено порядка 30-40, проверяющих бизнес-логику, транзакции и доступ по ролям.*

## Примеры API запросов (cURL)

### Регистрация
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "user123", "password": "password123", "role": "customer"}'
```

### Логин (получение токена)
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user123", "password": "password123"}'
```

### Логин в аккаунт администратора (данные прописаны в миграциях)
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "123123123"}'
```

### Создание задачи (Требуется JWT)
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer <ТОКЕН>" \
  -H "Content-Type: application/json" \
  -d '{"title": "Написать диплом", "description": "Срочно нужно сдавать", "reward": 500}'
```

### Список задач (Пагинация)
```bash
curl "http://localhost:8080/tasks?limit=10&offset=0" \
  -H "Authorization: Bearer <ТОКЕН>"
```

## Структура проекта
- `internal/app/` — создание компонентов, настройка маршрутизации.
- `internal/domain/` — конфигурация.
- `internal/domain/` — сущности, константы и бизнес-ошибки.
- `internal/handler/` — HTTP-слой, парсинг данных и валидация.
- `internal/service/` — бизнес-логика и координация репозиториев.
- `internal/repository/` — работа с базой данных (Postgres).
- `internal/middleware/` — проверки авторизации, ролей, блокировки; таймауты.
- `migrations/` — SQL-файлы миграций.
- `tests/integration` - тесты.

## Документация
Полная спецификация API в формате OpenAPI 3.0 доступна в файле `openapi.yaml`.