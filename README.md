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

### Создать короткую ссылку

```http
POST /links
Content-Type: application/json

{"url":"https://example.com"}
```

Ответ:

```json
{
  "short_url": "/links/Ab3_xY7kP2"
}
```

### Получить оригинальный URL

```http
GET /links/{short_code}
```

Ответ:

```json
{
  "url": "https://example.com"
}
```

Short code состоит ровно из 10 символов: `a-z`, `A-Z`, `0-9`, `_`.

## Тесты

Запустить unit-тесты:

```bash
go test ./...
```

Запустить HTTP benchmark:

```bash
make bench
```

Запустить нагрузочный тест:

```bash
make load-test
```

## Нагрузочное тестирование

Нагрузочный тест проверяет `GET /links/{short_code}` при 10, 50 и 100 одновременных запросах.

Результаты на локальной машине:

| Concurrency | Requests |    RPS | Avg latency | P95 latency | Errors |
|------------:|---------:| -----: | ----------: | ----------: | -----: |
|          10 |  100 000 | 18 373 |      504 µs |     1.15 ms |      0 |
|          50 |  100 000 | 53 985 |      846 µs |     3.17 ms |      0 |
|         100 |  100 000 | 62 921 |     1.43 ms |     4.25 ms |      0 |

Concurrency - количество workers, которые одновременно выполняют 
HTTP-запросы. Каждый worker отправляет запросы, пока не будет 
достигнуто заданное общее количество запросов.
На всех уровнях нагрузки все 100 000 запросов завершились успешно.

