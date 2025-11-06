package handlers

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
)

type Config struct {
	DB           *database.Queries
	JwtKey       string
	ClientApiKey string
	Port         string
	GeminiApiKey string
	R2           *R2Config
	AwsConfig    *aws.Config
	RABBITMQUrl  string
}

type EmployerProfile struct {
	ID              uuid.UUID
	CompanyName     string
	CompanyWebsite  string
	CompanySize     int32
	CompanyIndustry string
	UserID          uuid.UUID
}

type JobSeekerProfile struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	ResumeUrl sql.NullString
	UserID    uuid.UUID
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Result struct {
	ID        uuid.UUID
	Result    json.RawMessage
	CreatedAt time.Time
	SessionID uuid.UUID
}

type Resume struct {
	ID        uuid.UUID
	FileName  string
	Text      string
	CreatedAt time.Time
	SessionID uuid.UUID
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
type Session struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	Name           string    `json:"name"`
	UserID         uuid.UUID `json:"user_id"`
	Status         string    `json:"status"`
	JobTitle       string    `json:"job_title"`
	JobDescription string    `json:"job_description"`
}

type PresignResponse struct {
	UploadURL  string `json:"upload_url"`
	ObjectKey  string `json:"object_key"`
	Expiration int64  `json:"expiration"`
}

// Config you’ll have in your API struct
type R2Config struct {
	AccountID string
	Bucket    string
	AccessKey string
	SecretKey string
}

type PublishPayload struct {
	SessionID uuid.UUID `json:"session_d"`
}
