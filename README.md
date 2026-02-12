# Order Service (Kafka + PostgreSQL + Go)

Сервис получает сообщения из Kafka, сохраняет их в PostgreSQL, кэширует в памяти и отдает через HTTP API.  
Также есть простой фронт на HTML для поиска заказа, DLQ для невалидных сообщений.

---

## ⚙️ Стек
- **Go** — основной сервер
- **PostgreSQL** — база данных
- **Kafka** — очередь сообщений
- **Docker Compose** — запуск 
- **HTML** — фронтенд для просмотра заказа

---

## 🗣Основные команды

### Запуск инфраструктуры

```bash
docker compose up -d --build
```

### Проверка контейнеров

```bash
docker compose ps
```

### Просмотр логов приложения

```bash
docker compose logs -f app
```

### Генератор тестовых заказов

```bash
docker compose up generator
```


### Проверка HTTP

```bash
http://localhost:8080
```

---
### Сделаны Unit-Tests для Consumer и Server

Запуск теста
```bash
go test ./
```


## 📡 API

### Отправка заказа вручную

```bash
docker compose exec kafka bash -lc 'kafka-console-producer --bootstrap-server kafka:9092 --topic orders'
```
После вставляем JSON заказ в одну строку


### Получить заказ по `order_uid`

**Пример:**
```bash
curl http://localhost:8080/order/test124
```

---

## 📝 Пример заказа
```json
{
  "order_uid": "test124",
  "track_number": "TRACK124",
  "entry": "WBIL",
  "delivery": { "name": "John Doe", "city": "New York" },
  "payment": { "transaction": "txn124", "amount": 1500, "currency": "USD" },
  "items": [
    { "rid": "RID123", "name": "T-shirt", "brand": "Nike", "price": 500 },
    { "rid": "RID124", "name": "Sneakers", "brand": "Adidas", "price": 800 }
  ]
}
```

---

## 🗂Структура проекта

```bash
cmd/
  app/          
  generator/    

project/
  cache/
  consumer/
  database/
  http/
  model/

migrations/
docker-compose.yml
Dockerfile
```

## ✅В проекте сделано:
- **Подключение к Kafka**
- **Транзакционное сохранение в PostgreSQL**
- **Валидация входящих данных**
- **DLQ**
- **In-memory cache**
- **Восстановление кэша при старте**
- **HTTP API**
- **HTML интерфейс**
- **Генератор тестовых данных**
- **Docker Compose**
- **SQL Миграции**
- **Unit-тесты**
