# Subscriptions Service

Микросервис для учёта онлайн-подписок пользователей.  
Поддерживает CRUDL-операции и подсчёт суммарной стоимости подписок за период.

---

## Возможности

- CRUDL для подписок (`/subscriptions`)
- Подсчёт суммы подписок за период (`/subscriptions/summary`)
- Фильтрация по `user_id` и `service_name`
- PostgreSQL с миграциями
- Логирование и конфигурация через `.env` / `.yaml`
- Swagger-документация по OpenAPI 3.0

---

## Запуск

```bash
git clone https://github.com/pavel97go/subscriptions.git
cd subscriptions

docker compose up --build
```

После запуска сервис доступен по адресу:  
**http://localhost:8080**

Swagger UI — по адресу:  
**http://localhost:8081**

---

## Примеры запросов

### Создать подписку
```bash
curl -X POST http://localhost:8080/subscriptions   -H 'Content-Type: application/json'   -d '{
        "service_name": "Yandex Plus",
        "price": 400,
        "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
        "start_date": "07-2025",
        "end_date": "09-2025"
      }'
```

### Получить список подписок
```bash
curl "http://localhost:8080/subscriptions?limit=50&offset=0"
```

### Получить одну подписку
```bash
curl "http://localhost:8080/subscriptions/<id>"
```

### Обновить подписку
```bash
curl -X PUT "http://localhost:8080/subscriptions/<id>"   -H 'Content-Type: application/json'   -d '{
        "service_name": "Netflix",
        "price": 900,
        "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
        "start_date": "08-2025",
        "end_date": "10-2025"
      }'
```

### Удалить подписку
```bash
curl -X DELETE "http://localhost:8080/subscriptions/<id>"
```

### Подсчитать сумму подписок за период
```bash
curl "http://localhost:8080/subscriptions/summary?from=07-2025&to=09-2025"
```
Фильтрация:
```bash
curl "http://localhost:8080/subscriptions/summary?from=07-2025&to=09-2025&user_id=<uuid>&service_name=Netflix"
```

---

## Конфигурация

### `.env`
```env
APP_PORT=8080
DB_HOST=db
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=subscriptions_db
LOG_LEVEL=info
```

### `config.yaml`
```yaml
app_port: "8080"

db:
  host: "db"
  port: 5432
  user: "user"
  password: "password"
  name: "subscriptions_db"

log_level: "info"
```

---

## Стек технологий

- Go 1.25  
- Fiber (HTTP фреймворк)  
- PostgreSQL  
- Docker / Docker Compose  
- Swagger (OpenAPI 3.0)  
- YAML / .env для конфигурации  
- Логирование через logrus

---

## Структура проекта

```
cmd/app/            # main.go — точка входа
internal/
  ├── config/       # конфигурация (.env / YAML)
  ├── domain/       # модели данных
  ├── http/         # маршруты и хендлеры Fiber
  ├── repo/         # PostgreSQL-репозиторий
  ├── util/         # утилиты (работа с датами)
  └── logger/       # логирование
migrations/          # SQL миграции
api/openapi.yaml     # Swagger документация
docker-compose.yml
Dockerfile
config.yaml
```

---

## Тестирование API

```bash
# Создание подписки
curl -X POST http://localhost:8080/subscriptions -H 'Content-Type: application/json' -d '{"service_name":"Spotify","price":300,"user_id":"b13b8dbb-b3dc-4a1c-86b7-6bcd0f16a9ff","start_date":"09-2025"}'

# Получение списка
curl http://localhost:8080/subscriptions

# Подсчёт суммы
curl "http://localhost:8080/subscriptions/summary?from=09-2025&to=10-2025"
```

---

## Swagger Doc Reference

- Файл: `api/openapi.yaml`  
- UI: [http://localhost:8081](http://localhost:8081)

---

## Чек-лист соответствия ТЗ

| № | Требование                                                                 | Статус |
|---|-----------------------------------------------------------------------------|--------|
| 1 | HTTP CRUDL ручки для управления подписками                                 | ✅ |
| 2 | Подсчёт суммарной стоимости подписок за период                             | ✅ |
| 3 | Используется PostgreSQL, есть миграции                                    | ✅ |
| 4 | Код покрыт логированием                                                   | ✅ |
| 5 | Конфигурация вынесена в `.env` и `.yaml`                                 | ✅ |
| 6 | Swagger-документация предоставлена (OpenAPI 3.0)                           | ✅ |
| 7 | Проект запускается через Docker Compose                                   | ✅ |
|   | Проверка пользователя не требуется (соответствует примечанию в ТЗ)         | ✅ |
|   | Целые числа для стоимости (копейки не учитываются)                         | ✅ |

---
 
**Задание:** Реализация REST API сервиса учёта онлайн-подписок  
**Стек:** Golang, PostgreSQL, Docker, Fiber, Swagger  
