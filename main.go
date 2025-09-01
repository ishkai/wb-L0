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
)

func main() {
	time.Sleep(10 * time.Second)
	dbConn := os.Getenv("DB_CONN")
	if dbConn == "" {
		dbConn = "postgres://manager:qwerty12@postgres:5432/wb_db?sslmode=disable"
	}
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092"
	}

	db, err := database.NewDB(dbConn)
	if err != nil {
		log.Fatalf("Can't connect to database: %v", err)

	}
	defer db.Close()

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

	cons := consumer.NewConsumer(db, c)
	go cons.Start(ctx, kafkaBroker, "orders", "group")

	srv := http.NewServer(db, c)
	go func() {
		if err := srv.Run(":8080"); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")
	cancel()

}
