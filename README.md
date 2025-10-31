# URL Shortener (Go)

Небольшой сервис сокращения ссылок на Go с авторизацией по JWT и сбором статистики кликов.

## Возможности

- Регистрация и вход, выдача JWT (`/auth/register`, `/auth/login`).
- Создание сокращённой ссылки (`/link`) и переход по алиасу (`/{alias}`).
- Получение списка ссылок с пагинацией (`/link?limit&offset`).
- Обновление и удаление ссылки (требует авторизации) (`PATCH /link/{id}`, `DELETE /link/{id}`).
- Сбор статистики посещений с агрегированием по дням/месяцам (`GET /stat?from&to&by`).
- Middleware: CORS, логирование запросов, проверка JWT.

## Технологии

- `net/http` — HTTP сервер и роутер (`ServeMux`).
- `gorm` + `postgres` — ORM и база данных.
- `github.com/joho/godotenv` — загрузка `.env`.
- JWT: простая обёртка в `pkg/jwt`.
- Внутренние пакеты: `internal/*` (handlers, services, repositories), `pkg/*` (middleware, req/res, eventbus, db).

## Требования

- Go 1.21+ (или совместимая версия).
- PostgreSQL (локально или в Docker).
- Переменные окружения: `DSN`, `SECRET`.

Пример `.env`:

```
DSN=postgres://user:password@localhost:5432/shortener?sslmode=disable
SECRET=supersecret
```

## Быстрый старт

1. Установите зависимости и проверьте сборку:
   - `go build ./...`

2. Запустите миграции (создание таблиц):
   - `go run migrations/auto.go`

3. Запуск приложения:
   - `go run cmd/main.go`
   - Сервер слушает на `http://localhost:8081`.

Опционально: используйте `docker-compose.yml` для запуска PostgreSQL (если файл настроен). После старта БД — выполните миграции и запустите сервер, как указано выше.

## Маршруты API

Аутентификация:
- `POST /auth/register` — регистрирует пользователя, возвращает `token`.
- `POST /auth/login` — логин, возвращает `token`.

Ссылки:
- `POST /link` — создать ссылку. Тело: `{ "url": "https://example.com" }`. Ответ: объект `Link` с `id`, `url`, `hash`.
- `GET /link?limit=10&offset=0` — получить список ссылок и `count`.
- `PATCH /link/{id}` — обновить `url` и/или `hash`. Требует `Authorization: Bearer <token>`.
- `DELETE /link/{id}` — удалить ссылку. Возвращает `204 No Content`. Требует `Authorization: Bearer <token>`.
- `GET /{alias}` — редирект на исходный `url` (`307 Temporary Redirect`). Параллельно публикуется событие для статистики.

Статистика (требует авторизацию):
- `GET /stat?from=YYYY-MM-DD&to=YYYY-MM-DD&by=day|month` — отдаёт агрегированную статистику.

## Примеры запросов

Регистрация:

```
curl -X POST http://localhost:8081/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@mail.com","password":"123","name":"User"}'
```

Логин:

```
curl -X POST http://localhost:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@mail.com","password":"123"}'
```

Создать ссылку:

```
curl -X POST http://localhost:8081/link \
  -H "Content-Type: application/json" \
  -d '{"url":"https://golang.org"}'
```

Обновить ссылку (пример с токеном):

```
curl -X PATCH http://localhost:8081/link/1 \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev","hash":"go1234"}'
```

Удалить ссылку:

```
curl -X DELETE http://localhost:8081/link/1 \
  -H "Authorization: Bearer <TOKEN>"
```

Перейти по алиасу:

```
open http://localhost:8081/<HASH>
```

Получить статистику:

```
curl "http://localhost:8081/stat?from=2024-01-01&to=2024-12-31&by=month" \
  -H "Authorization: Bearer <TOKEN>"
```

## Тесты

- Запустить все тесты: `go test ./...`
- Есть модульные тесты для `internal/auth` и `pkg/jwt`, а также интеграционные тесты (`cmd/auth_test.go`) с `httptest`.

## Архитектура

- `cmd/main.go` — сборка приложения: конфиг, БД, шина событий, репозитории, сервисы, хендлеры, последовательность middleware.
- `internal/auth/*` — аутентификация и авторизация, `AuthService`, обработчики.
- `internal/link/*` — модели, репозиторий и `LinkService` (генерация уникального хеша, CRUD, редирект с публикацией события), обработчики.
- `internal/stat/*` — репозиторий/сервис и хендлер статистики; сервис слушает события из `EventBus` и записывает клики.
- `pkg/middleware/*` — CORS, логирование, проверка JWT.
- `pkg/req` и `pkg/res` — декодирование/валидация запросов и унифицированная отдача ответов.
- `pkg/event` — простая шина событий (канал), используется для считывания кликов.
- `pkg/db` — инициализация подключения к Postgres через GORM.
- `configs` — загрузка переменных окружения.

## Примечания

- Для пагинации по умолчанию `limit=10`, `offset=0` (если параметры не переданы или некорректны).
- Для защищённых маршрутов используйте заголовок `Authorization: Bearer <token>`.
- Перед первым запуском не забудьте выполнить миграции: `go run migrations/auto.go`.