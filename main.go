package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/muhammadolammi/jobmatchapi/infra"
)

func main() {
	_ = godotenv.Load()

	cfg := buildConfig()

	// Base context for all connections
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect services in goroutines
	go infra.ConnectRabbit(ctx, &cfg)
	go infra.LoadAWSConfig(&cfg, cfg.R2)
	go infra.ConnectPubSub(ctx, &cfg)

	// Blocking DB connection (or just ensure connection pool)
	infra.ConnectDB(ctx, &cfg)

	// Start your server in goroutine
	go func() {
		server(&cfg)
	}()

	// Wait for Cloud Run shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)

	<-stop
	log.Println("SIGTERM received: shutting down gracefully...")

	_, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Close all connections safely
	if cfg.DBConn != nil {
		cfg.DBConn.Close()
		log.Println("Postgres connection closed")
	}

	if cfg.RabbitChan != nil {
		cfg.RabbitChan.Close()
		log.Println("RabbitMQ channel closed")
	}
	if cfg.RabbitConn != nil {
		cfg.RabbitConn.Close()
		log.Println("RabbitMQ connection closed")
	}
	if cfg.PubSubClient != nil {
		cfg.PubSubClient.Close()
		log.Println("Pub/Sub client closed")
	}

	log.Println("All resources cleaned up. Exiting...")
}
