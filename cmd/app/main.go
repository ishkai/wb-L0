package main

import (
	"awesomeProject3/project/cache"
	"awesomeProject3/project/consumer"
	"awesomeProject3/project/database"
	"awesomeProject3/project/http"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	dbConn := os.Getenv("DB_CONN")
	if dbConn == "" {
		dbConn = "postgres://manager:qwerty12@postgres:5432/wb_db?sslmode=disable"
	}
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092"
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "orders"
	}

	kafkaGroup := os.Getenv("KAFKA_GROUP")
	if kafkaGroup == "" {
		kafkaGroup = "group"
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	dlqTopic := os.Getenv("DLQ_TOPIC")
	if dlqTopic == "" {
		dlqTopic = "orders_dlq"
	}

	db, err := database.NewDB(dbConn)
	if err != nil {
		log.Fatalf("Can't connect to database: %v", err)

	}
	defer db.Close()

	dlqWriter := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    dlqTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer dlqWriter.Close()

	c := cache.New(5 * time.Minute)

	if orders, err := db.GetAllOrders(); err == nil {
		for _, o := range orders {
			c.Set(o.OrderUID, o)
		}
		log.Printf("Found %d orders", len(orders))
	} else {
		log.Printf("Can't find orders: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cons := consumer.NewConsumer(db, c, dlqWriter)
	go cons.Start(ctx, kafkaBroker, kafkaTopic, kafkaGroup)

	srv := http.NewServer(db, c)
	go func() {
		if err := srv.Run(httpAddr); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")
	cancel()

}
