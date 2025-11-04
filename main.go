package main

import (
	"database/sql"
	"log"
	"os"

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
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("empty PORT in environment")
	}
	clientApiKey := os.Getenv("CLIENT_API_KEY")
	if clientApiKey == "" {
		log.Fatal("empty CLIENT_API_KEY in environment")
	}
	langflowUrl := os.Getenv("LANGFLOW_URL")
	if langflowUrl == "" {
		log.Fatal("empty LANGFLOW_URL in environment")
	}
	langflowApiKey := os.Getenv("LANGFLOW_API_KEY")
	if langflowApiKey == "" {
		log.Fatal("empty LANGFLOW_API_KEY in environment")
	}
	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		log.Fatal("empty JWT_KEY in environment")
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("error opening db. err: ", err)
	}

	queries := database.New(db)
	apiConfig := handlers.Config{
		DB:             queries,
		Port:           port,
		ClientApiKey:   clientApiKey,
		LangflowApiKey: langflowApiKey,
		LangflowUrl:    langflowUrl,
		JwtKey:         jwtKey,
	}
	server(&apiConfig)

}
