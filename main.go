package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
)

func main() {

	_ = godotenv.Load()

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("empty DB_URL in environment")
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("empty PORT in environment")
	}
	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		log.Fatal("empty AUTH_TOKEN in environment")
	}
	langflowUrl := os.Getenv("LANGFLOW_URL")
	if langflowUrl == "" {
		log.Fatal("empty LANGFLOW_URL in environment")
	}
	langflowApiKey := os.Getenv("LANGFLOW_API_KEY")
	if langflowApiKey == "" {
		log.Fatal("empty LANGFLOW_API_KEY in environment")
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("error opening db. err: ", err)
	}

	queries := database.New(db)
	apiConfig := Config{
		DB:               *queries,
		PORT:             port,
		AUTH_TOKEN:       authToken,
		LANGFLOW_API_KEY: langflowApiKey,
		LANGFLOW_URL:     langflowUrl,
	}
	server(&apiConfig)

}
