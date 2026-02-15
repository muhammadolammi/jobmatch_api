package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/muhammadolammi/jobmatchapi/internal/handlers"
)

func main() {

	_ = godotenv.Load()
	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "deployment"
	}
	if environment != "deployment" && environment != "development" {
		log.Fatal("ENV can only be deployment or development. got: ", environment)
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("empty DB_URL in environment")
	}

	rabbitmqUrl := os.Getenv("RABBITMQ_URL")
	if rabbitmqUrl == "" {
		log.Fatal("empty RABBITMQ_URL in env")
	}

	r2AccountId := os.Getenv("R2_ACCOUNT_ID")
	if r2AccountId == "" {
		log.Fatal("empty R2_ACCOUNT_ID in environment")
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

	//  we assume its api mode if no runmode is provider
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	clientApiKey := os.Getenv("CLIENT_API_KEY")
	if clientApiKey == "" {
		log.Fatal("empty CLIENT_API_KEY in environment")
	}
	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		log.Fatal("empty JWT_KEY in environment")
	}
	paystackSecretKey := os.Getenv("PAYSTACK_SECRET_KEY")
	if paystackSecretKey == "" {
		log.Fatal("empty PAYSTACK_SECRET_KEY in environment")
	}

	httpClient := http.Client{
		Timeout: time.Minute,
	}
	apiConfig := handlers.Config{
		// DB : dbqueries,
		DBURL:        dbUrl,
		RABBITMQUrl:  rabbitmqUrl,
		Port:         port,
		ClientApiKey: clientApiKey,
		JwtKey:       jwtKey,
		R2:           &r2Config,
		// AwsConfig:                  &awsConfig,
		RefreshTokenEXpirationTime: 60 * 24 * 7, //7 days
		AcessTokenEXpirationTime:   15,
		// RabbitConn:                 conn,
		RateLimit:         2, // lets just rate limit to 2 for now
		PaystackApi:       "https://api.paystack.co",
		HttpClient:        &httpClient,
		PaystackSecretKey: paystackSecretKey,
		ENV:               environment,
	}
	go connectRabbitMQ(&apiConfig)
	go connectDB(&apiConfig)
	go loadAWSConfig(&apiConfig, &r2Config)
	server(&apiConfig)
}
