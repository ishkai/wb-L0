package consumer

import (
	"awesomeProject3/project/cache"
	"awesomeProject3/project/database"
	"awesomeProject3/project/model"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	DB    *database.Database
	cache *cache.Cache
}

func NewConsumer(db *database.Database, cache *cache.Cache) *Consumer {
	return &Consumer{DB: db, cache: cache}
}

func (c *Consumer) Start(ctx context.Context, broker string, topic string, group string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: group,
	})
	log.Printf("Consumer starting listening on: topic=%s group=%s", topic, group)
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer shutting down...")
			return
		default:
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Can't read message: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
			var order model.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("Can't unmarshal json: %v", err)
				continue
			}
			if err := c.DB.InsertOrder(order); err != nil {
				log.Printf("Can't insert order: %v", err)
				continue
			}
			c.cache.Set(order.OrderUID, order)

			if err := r.CommitMessages(ctx, msg); err != nil {
				log.Printf("Can't commit messages: %v", err)
			}

			log.Printf("Consumer committed order: %s", order.OrderUID)

		}

	}
}
