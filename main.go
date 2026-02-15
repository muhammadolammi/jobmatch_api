package main

import (
	"context"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {

	_ = godotenv.Load()

	cfg := buildConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go connectDB(ctx, &cfg)
	go connectRabbit(ctx, &cfg)
	go loadAWSConfig(&cfg, cfg.R2)
	server(&cfg)
}
