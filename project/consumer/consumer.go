package consumer

import (
	"awesomeProject3/project/cache"
	"awesomeProject3/project/database"
	"awesomeProject3/project/model"
	"awesomeProject3/project/validation"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type DLQMessage struct {
	Error     string          `json:"error"`
	Payload   json.RawMessage `json:"payload"`
	Topic     string          `json:"topic"`
	Partition int             `json:"partition"`
	Offset    int64           `json:"offset"`
	Timestamp time.Time       `json:"timestamp"`
}

type CS interface {
	Start(ctx context.Context, broker string, topic string, group string)
}

type Consumer struct {
	DB         database.DB
	cache      cache.CC
	dlqWriter  *kafka.Writer
	validateFn func(*model.Order) error
}

func (c *Consumer) sendToDLQ(ctx context.Context, msg kafka.Message, procErr error) {
	if c.dlqWriter == nil {
		log.Printf("DLQ writer is nil, can't send message to DLQ: %v", procErr)
		return
	}

	dlq := DLQMessage{
		Error:     procErr.Error(),
		Payload:   msg.Value,
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(dlq)
	if err != nil {
		log.Printf("Can't marshal DLQ message: %v", err)
		return
	}

	if err := c.dlqWriter.WriteMessages(ctx, kafka.Message{
		Key:   msg.Key,
		Value: data,
	}); err != nil {
		log.Printf("Can't write message to DLQ: %v", err)
	}
}

func NewConsumer(db database.DB, cache cache.CC, dlqwritrer *kafka.Writer) *Consumer {
	return &Consumer{DB: db, cache: cache, dlqWriter: dlqwritrer, validateFn: validation.ValidateOrder}
}

func (c *Consumer) Start(ctx context.Context, broker, topic, group string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: group,
	})
	defer r.Close()

	log.Printf("Consumer starting listening on: topic=%s group=%s", topic, group)

	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("Can't read message: %v", err)
			continue
		}

		commit := c.HandleMessage(ctx, msg)

		if commit {
			if err := r.CommitMessages(ctx, msg); err != nil {
				log.Printf("Can't commit message: %v", err)
			}
		} else {
			log.Printf("Retrying message...")
			time.Sleep(2 * time.Second)
		}
	}
}

func (c *Consumer) HandleMessage(ctx context.Context, msg kafka.Message) bool {
	var order model.Order

	if err := json.Unmarshal(msg.Value, &order); err != nil {
		log.Printf("Can't unmarshal json: %v", err)
		c.sendToDLQ(ctx, msg, fmt.Errorf("unmarshal: %w", err))
		return true
	}

	if c.validateFn != nil {
		if err := c.validateFn(&order); err != nil {
			log.Printf("Invalid order data: %v", err)
			c.sendToDLQ(ctx, msg, fmt.Errorf("validation: %w", err))
			return true
		}
	}

	if err := c.DB.InsertOrder(order); err != nil {
		log.Printf("Can't insert order: %v", err)
		return false
	}

	c.cache.Set(order.OrderUID, order)

	log.Printf("Order processed: %s", order.OrderUID)
	return true
}
