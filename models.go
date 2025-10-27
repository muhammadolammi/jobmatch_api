package main

import "github.com/muhammadolammi/jobmatchapi/internal/database"

type Config struct {
	DB               database.Queries
	AUTH_TOKEN       string
	PORT             string
	LANGFLOW_API_KEY string
	LANGFLOW_URL     string
}
