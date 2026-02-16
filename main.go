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

	infra.ConnectRabbit(ctx, &cfg)
	infra.LoadAWSConfig(&cfg, cfg.R2)
	infra.ConnectPubSub(ctx, &cfg)
	infra.ConnectDB(ctx, &cfg)

	server(&cfg)
}
