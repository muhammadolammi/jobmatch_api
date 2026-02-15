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

	go infra.ConnectDB(ctx, &cfg)
	go infra.ConnectRabbit(ctx, &cfg)
	go infra.LoadAWSConfig(&cfg, cfg.R2)
	server(&cfg)
}
