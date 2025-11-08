package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func HelloReady(w http.ResponseWriter, r *http.Request) { helpers.RespondWithJson(w, 200, "hello") }
func ErrorReady(w http.ResponseWriter, r *http.Request) {
	helpers.RespondWithError(w, 200, "this is an error test")
}

func (apiConfig *Config) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	body := struct {
		SessionID string `json:"session_id"`
	}{}
	encoder := json.NewDecoder(r.Body)
	err := encoder.Decode(&body)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error decoding request body. err: "+err.Error())
		return
	}
	if body.SessionID == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "session_id can't be empty ")
		return
	}
	sessionUUid, err := uuid.Parse(body.SessionID)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error parsing uuid. err: "+err.Error())
		return
	}
	session, err := apiConfig.DB.GetSession(r.Context(), sessionUUid)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting session from db. err: "+err.Error())
		return
	}
	//  set session status to pending
	err = apiConfig.DB.UpdateSessionStatus(r.Context(), database.UpdateSessionStatusParams{
		ID:     session.ID,
		Status: "pending",
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error updating session status to pending(db error). err: "+err.Error())
		return
	}

	// publish the session
	err = apiConfig.PublishSession(DbSessionToModelsSession(session), apiConfig.RabbitConn)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error queing session. err: "+err.Error())
		return
	}

	helpers.RespondWithJson(w, http.StatusOK, "workflow queued")
}

func (apiConfig *Config) PresignUploadHandler(w http.ResponseWriter, r *http.Request, user User) {
	sessionID := chi.URLParam(r, "id")
	var body struct {
		Filename string `json:"file_name"`
		MimeType string `json:"mime_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.Filename == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "include file_name in request")
		return
	}
	if body.MimeType == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "include mime_type in request")
		return
	}
	// Generate unique key for this resume
	objectKey := ""
	if user.Role == "employer" {
		objectKey = fmt.Sprintf("sessions/%s/%s", sessionID, body.Filename)
	} else {
		objectKey = fmt.Sprintf("sessions/%s/resume.%s", sessionID, body.MimeType)
	}
	client := s3.NewFromConfig(*apiConfig.AwsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", apiConfig.R2.AccountID))
	})
	presignClient := s3.NewPresignClient(client)
	presignResult, err := presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(apiConfig.R2.Bucket),
		Key:         aws.String(objectKey),
		ContentType: aws.String(body.MimeType),
	})
	if err != nil {
		msg := fmt.Sprintf("Couldn't get presigned URL for PutObject. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	resp := PresignResponse{
		UploadURL:  presignResult.URL,
		ObjectKey:  objectKey,
		Expiration: time.Now().Add(15 * time.Minute).Unix(),
	}
	helpers.RespondWithJson(w, http.StatusOK, resp)

}

func (apiConfig *Config) UploadCompleteHandler(w http.ResponseWriter, r *http.Request, user User) {
	var body struct {
		SessionID  string `json:"session_id"`
		ObjectKey  string `json:"object_key"`
		Filename   string `json:"file_name"`
		Size       int64  `json:"size"`
		MimeType   string `json:"mime_type"`
		StorageUrl string `json:"storage_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Filename == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "include file_name in request")
		return
	}
	if body.ObjectKey == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "include object_key in request")
		return
	}
	if body.SessionID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "include session_id in request")
		return
	}
	if body.MimeType == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "include mime_type in request")
		return
	}
	if body.Size == 0 {
		helpers.RespondWithError(w, http.StatusBadRequest, "include size in request")
		return
	}

	sessionUUid, err := uuid.Parse(body.SessionID)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error parsing uuid. err: %v", err))
		return
	}
	// If user is job seeker just update the resume for that session and create one if session has no resume
	resumeExists, _ := apiConfig.DB.ResumeExists(r.Context(), sessionUUid)
	if user.Role == "job_seeker" && resumeExists {
		err = apiConfig.DB.UpdateResumeStorageUrlForSession(r.Context(), database.UpdateResumeStorageUrlForSessionParams{
			StorageUrl:       body.StorageUrl,
			ObjectKey:        body.ObjectKey,
			OriginalFilename: body.Filename,
			Mime:             body.MimeType,
			SizeBytes:        body.Size,
			StorageProvider:  "r2",
			UploadStatus:     "uploaded",
		})
		if err != nil {
			helpers.RespondWithError(w, http.StatusInternalServerError, "db err: "+err.Error())
			log.Println(err)
			return
		}

		helpers.RespondWithJson(w, http.StatusCreated, "")
		return

	}
	_, err = apiConfig.DB.CreateResume(r.Context(), database.CreateResumeParams{
		SessionID:        sessionUUid,
		ObjectKey:        body.ObjectKey,
		OriginalFilename: body.Filename,
		Mime:             body.MimeType,
		SizeBytes:        body.Size,
		StorageProvider:  "r2",
		StorageUrl:       body.StorageUrl,
		UploadStatus:     "uploaded",
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "db err: "+err.Error())
		log.Println(err)
		return
	}

	helpers.RespondWithJson(w, http.StatusCreated, "")
}
