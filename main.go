package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/handlers"
)

func main() {

	_ = godotenv.Load()

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("empty DB_URL in environment")
	}
	runMode := os.Getenv("RUN_MODE")
	if runMode == "" {
		log.Fatal("empty RUN_MODE in environment")

	}
	if runMode != "server" && runMode != "worker" {
		log.Fatal("RUN_MODE can either be server or worker")

	}
	rabbitmqUrl := os.Getenv("RABBITMQ_URL")
	if rabbitmqUrl == "" {
		log.Fatal("empty RABBITMQ_URL in env")
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("error opening db. err: ", err)
	}

	queries := database.New(db)
	apiConfig := handlers.Config{
		DB:          queries,
		RABBITMQUrl: rabbitmqUrl,
	}

	if runMode == "server" {
		//  we assume its api mode if no runmode is provider
		port := os.Getenv("PORT")
		if port == "" {
			log.Fatal("empty PORT in environment")
		}
		clientApiKey := os.Getenv("CLIENT_API_KEY")
		if clientApiKey == "" {
			log.Fatal("empty CLIENT_API_KEY in environment")
		}
		jwtKey := os.Getenv("JWT_KEY")
		if jwtKey == "" {
			log.Fatal("empty JWT_KEY in environment")
		}
		r2AccountId := os.Getenv("R2_ACCCOUNT_ID")
		if r2AccountId == "" {
			log.Fatal("empty R2_ACCCOUNT_ID in environment")
		}
		r2Bucket := os.Getenv("R2_BUCKET")
		if r2Bucket == "" {
			log.Fatal("empty R2_BUCKET in environment")
		}
		r2SecretKey := os.Getenv("R2_SECRET_KEY")
		if r2SecretKey == "" {
			log.Fatal("empty R2_SECRET_KEY in environment")
		}
		r2AccessKey := os.Getenv("R2_ACCESS_KEY")
		if r2AccessKey == "" {
			log.Fatal("empty R2_ACCESS_KEY in environment")
		}
		r2Config := handlers.R2Config{
			AccountID: r2AccountId,
			AccessKey: r2AccessKey,
			SecretKey: r2SecretKey,
			Bucket:    r2Bucket,
		}
		awsConfig, err := config.LoadDefaultConfig(context.TODO(),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r2Config.AccessKey, r2Config.SecretKey, "")),
			config.WithRegion("auto"),
		)
		if err != nil {
			log.Fatal("error creating aws config", err)
		}

		apiConfig.Port = port
		apiConfig.ClientApiKey = clientApiKey
		apiConfig.JwtKey = jwtKey
		apiConfig.R2 = &r2Config
		apiConfig.AwsConfig = &awsConfig
		server(&apiConfig)
	}
	if runMode == "worker" {
		geminiApiKey := os.Getenv("GEMINI_API_KEY")
		if geminiApiKey == "" {
			log.Fatal("empty GEMINI_API_KEY in env")
		}
		apiConfig.GeminiApiKey = geminiApiKey
		apiConfig.StartConsumerWorkerPool(3)
	}

}
