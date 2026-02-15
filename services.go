package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/handlers"
	"github.com/streadway/amqp"
)

func connectRabbitMQ(config *handlers.Config) {
	conn, err := amqp.Dial(config.RABBITMQUrl)
	if err != nil {
		log.Println("RabbitMQ not ready:", err)
		return

	}

	config.RabbitConn = conn
}

func connectDB(config *handlers.Config) {
	db, err := sql.Open("postgres", config.DBURL)
	if err != nil {
		log.Println("DB error:", err)
		return
	}

	dbqueries := database.New(db)

	config.DB = dbqueries
}

func loadAWSConfig(apiconfig *handlers.Config, r2Config *handlers.R2Config) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r2Config.AccessKey, r2Config.SecretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Println("error creating aws config", err)
		return
	}
	apiconfig.AwsConfig = &awsConfig
}
