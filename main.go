package main

import (
	"context"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/muhammadolammi/jobmatchapi/infra"
)

func main() {

	_ = godotenv.Load()

	cfg := buildConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go infra.ConnectRabbit(ctx, &cfg)
	go infra.LoadAWSConfig(&cfg, cfg.R2)
	go infra.ConnectPubSub(ctx, &cfg)
	infra.ConnectDB(ctx, &cfg)
	defer func() {
		if cfg.DBConn != nil {
			cfg.DBConn.Close()
			// log.Println("✅ Postgres disconnected")
		}

		if cfg.RabbitChan != nil {
			// log.Println("✅ RabbitMQ channel disconnected")
			cfg.RabbitChan.Close()
		}
		if cfg.RabbitConn != nil {
			// log.Println("✅ RabbitMQ connection disconnected")
			cfg.RabbitConn.Close()
		}
		if cfg.PubSubClient != nil {
			// log.Println("✅ Pub/Sub client disconnected")
			cfg.PubSubClient.Close()
		}

	}()
	server(&cfg)
}
