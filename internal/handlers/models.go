package handlers

import (
	"database/sql"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/streadway/amqp"
)

type Config struct {
	DB                         *database.Queries
	JwtKey                     string
	ClientApiKey               string
	Port                       string
	R2                         *R2Config
	AwsConfig                  *aws.Config
	RABBITMQUrl                string
	RabbitConn                 *amqp.Connection
	RefreshTokenEXpirationTime int //in minute
	AcessTokenEXpirationTime   int //in minute
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

type R2Config struct {
	AccountID string
	Bucket    string
	AccessKey string
	SecretKey string
}

type PublishPayload struct {
	SessionID uuid.UUID `json:"session_d"`
}

type AnalysesResult struct {
	CandidateEmail      string   `json:"candidate_email"`
	MatchScore          int      `json:"match_score"`
	RelevantExperiences []string `json:"relevant_experiences"`
	RelevantSkills      []string `json:"relevant_skills"`
	MissingSkills       []string `json:"missing_skills"`
	Summary             string   `json:"summary"`
	Recomendation       string   `json:"recommendation"`
	// Error result entry
	IsErrorResult bool   `json:"is_error_result"`
	Error         string `json:"error,omitempty"`
}
type AnalysesResults struct {
	ID        uuid.UUID        `json:"id"`
	Results   []AnalysesResult `json:"results" db:"results"`
	CreatedAt time.Time        `json:"created_at"`
	SessionID uuid.UUID        `json:"session_id"`
	UpdatedAt time.Time        `json:"updated_at"`
}
