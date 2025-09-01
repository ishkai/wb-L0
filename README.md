# Order Service (Kafka + PostgreSQL + Go)

Сервис получает сообщения из Kafka, сохраняет их в PostgreSQL, кэширует в памяти и отдает через HTTP API.  
Также есть простой фронт на HTML для поиска заказа.

---

## ⚙️ Стек
- **Go** — основной сервер
- **PostgreSQL** — база данных
- **Kafka** — очередь сообщений
- **Docker Compose** — запуск 
- **HTML** — фронтенд для просмотра заказа

---


## 📡 API

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


