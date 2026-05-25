# Effective Mobile test task

REST-сервис для хранения онлайн-подписок пользователей и подсчета суммарной стоимости подписок за период.

## Стек

- Go
- PostgreSQL
- Docker Compose
- golang-migrate

## Запуск

```bash
docker compose up --build
```

Приложение будет доступно на `http://localhost:8080`.

Переменные окружения лежат в `.env`:

```env
HTTP_PORT=8080
DATABASE_URL=postgres://postgres:postgres@db:5432/subscriptions?sslmode=disable
```

## API

Swagger/OpenAPI спецификация: `docs/openapi.yaml`.

### Создать подписку

```bash
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

### Получить список подписок

```bash
curl "http://localhost:8080/subscriptions"
```

Можно фильтровать:

```bash
curl "http://localhost:8080/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus"
```

### Получить подписку по id

```bash
curl "http://localhost:8080/subscriptions/{id}"
```

### Обновить подписку

```bash
curl -X PUT http://localhost:8080/subscriptions/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 500,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025",
    "end_date": "12-2025"
  }'
```

### Удалить подписку

```bash
curl -X DELETE "http://localhost:8080/subscriptions/{id}"
```

### Посчитать сумму за период

```bash
curl "http://localhost:8080/subscriptions/total?start_date=07-2025&end_date=12-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus"
```

Ответ:

```json
{
  "total": 2400
}
```
