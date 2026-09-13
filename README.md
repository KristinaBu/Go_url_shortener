# URL Shortener

HTTP-сервис для сокращения ссылок на Go.

Сервис поддерживает два варианта хранения данных: in-memory и PostgreSQL. Для одного оригинального URL создаётся только одна короткая ссылка.

## Запуск

In-memory:

```bash
go run ./cmd/server -storage memory
```

PostgreSQL:

```bash
docker compose up -d
```

## API

Создать короткую ссылку:

```http
POST /links
Content-Type: application/json

{"url":"https://example.com"}
```

Получить оригинальный URL:

```http
GET /links/{short_code}
```

Short code состоит из 10 символов: `a-z`, `A-Z`, `0-9`, `_`.

## Тесты

```bash
go test ./...
```

