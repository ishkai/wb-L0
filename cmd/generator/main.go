package main

import (
	"awesomeProject3/project/model"
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
)

func main() {
	broker := getEnv("KAFKA_BROKER", "localhost:9092")
	topic := getEnv("KAFKA_TOPIC", "orders")

	log.Printf("Start generator: broker=%s topic=%s", broker, topic)

	w := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	defer w.Close()

	gofakeit.Seed(time.Now().UnixNano())

	ctx := context.Background()

	for {
		order := randomOrder()

		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("marshal error: %v", err)
			continue
		}

		err = w.WriteMessages(ctx, kafka.Message{
			Key:   []byte(order.OrderUID),
			Value: data,
		})
		if err != nil {
			log.Printf("write to kafka error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("sent order %s", order.OrderUID)
		time.Sleep(500 * time.Millisecond)
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func randomOrder() model.Order {
	orderUID := gofakeit.UUID()
	track := gofakeit.LetterN(10)

	item := model.Items{
		ChrtID:      gofakeit.Number(1000000, 9999999),
		TrackNumber: track,
		Price:       500,
		Name:        gofakeit.ProductName(),
		Sale:        50,
		Size:        "0",
		TotalPrice:  450,
		NmID:        gofakeit.Number(100000, 999999),
		Brand:       gofakeit.Company(),
		Status:      202,
	}

	delivery := model.Delivery{
		Name:    gofakeit.Name(),
		Phone:   gofakeit.Phone(),
		Zip:     gofakeit.Zip(),
		City:    gofakeit.City(),
		Address: gofakeit.Street(),
		Region:  gofakeit.State(),
		Email:   gofakeit.Email(),
	}

	payment := model.Payment{
		Transaction:  orderUID,
		RequestID:    "",
		Currency:     "USD",
		Provider:     "wbpay",
		Amount:       450 + 200,
		PaymentDT:    time.Now().Unix(),
		Bank:         "alpha",
		DeliveryCost: 200,
		GoodsTotal:   450,
		CustomFee:    0,
	}

	return model.Order{
		OrderUID:        orderUID,
		TrackNumber:     track,
		Entry:           "WBIL",
		Delivery:        delivery,
		Payment:         payment,
		Items:           []model.Items{item},
		Locale:          "en",
		CustomerID:      gofakeit.Username(),
		DeliveryService: "meest",
		Shardkey:        "9",
		SmID:            99,
		DateCreated:     time.Now().UTC().Format(time.RFC3339),
		OofShard:        "1",
	}
}
