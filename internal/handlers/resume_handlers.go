package handlers

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func HelloReady(w http.ResponseWriter, r *http.Request) { helpers.RespondWithJson(w, 200, "hello") }
func ErrorReady(w http.ResponseWriter, r *http.Request) {
	helpers.RespondWithError(w, 200, "this is an error test")
}
func (apiConfig *Config) UploadHandler(w http.ResponseWriter, r *http.Request, user User) {
	const (
		MaxUploadSize  = 20 << 20 // 20MB
		MaxMemoryUsage = 10 << 20 // 10MB
	)

	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
	if err := r.ParseMultipartForm(MaxMemoryUsage); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid multipart form: "+err.Error())
		return
	}

	sessionIDStr := r.FormValue("session_id")
	if sessionIDStr == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Missing session_id in form data")
		return
	}
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid session_id format")
		return
	}

	result, err := apiConfig.DB.GetResultBySession(r.Context(), sessionID)
	if err != nil && err != sql.ErrNoRows {
		log.Println("DB error:", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if result.ID != uuid.Nil && time.Since(result.CreatedAt) < time.Hour {
		remaining := time.Hour - time.Since(result.CreatedAt)
		helpers.RespondWithJson(w, http.StatusTooManyRequests, map[string]any{
			"error":            "rate_limit_exceeded",
			"message":          fmt.Sprintf("Please wait %.0f minutes before trying again.", remaining.Minutes()),
			"retry_in_minutes": int(remaining.Minutes()),
		})
		return
	}

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		helpers.RespondWithError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	if len(files) > 1 {
		helpers.RespondWithError(w, http.StatusBadRequest, "Only one file allowed")
		return
	}

	errorMsgs := []string{}
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			errorMsgs = append(errorMsgs, "Could not open file: "+fh.Filename)
			continue
		}
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if !map[string]bool{".pdf": true, ".docx": true, ".txt": true}[ext] {
			errorMsgs = append(errorMsgs, "Invalid file type: "+fh.Filename)
			continue
		}

		data, err := io.ReadAll(file)
		if err != nil {
			errorMsgs = append(errorMsgs, "Failed reading file: "+fh.Filename)
			continue
		}

		encoded := base64.StdEncoding.EncodeToString(data)
		if _, err := apiConfig.DB.CreateResume(r.Context(), database.CreateResumeParams{
			FileName:  filepath.Base(fh.Filename),
			Text:      encoded,
			SessionID: sessionID,
		}); err != nil {
			log.Println("DB error creating resume:", err)
			errorMsgs = append(errorMsgs, "Internal error processing "+fh.Filename)
			continue
		}
	}

	if len(errorMsgs) > 0 {
		status := http.StatusOK
		if len(errorMsgs) == len(files) {
			status = http.StatusBadRequest
		}
		helpers.RespondWithJson(w, status, map[string]any{
			"message": "Upload completed with some errors",
			"errors":  errorMsgs,
		})
		return
	}

	helpers.RespondWithJson(w, http.StatusOK, map[string]string{"message": "Upload successful"})
}

func (apiConfig *Config) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	body := struct {
		SessionID      string `json:"session_id"`
		JobTitle       string `json:"job_title"`
		JobDescription string `json:"job_description"`
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
	if body.JobDescription == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "job_description can't be empty")
		return

	}
	if body.JobTitle == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "job_title can't be empty")
		return

	}
	langflowInput := fmt.Sprintf("Job Title:\n %s\nJob Description:\n%s\n", body.JobTitle, body.JobDescription)
	payload := map[string]interface{}{
		"input_type":  "chat",
		"output_tye":  "chat",
		"input_value": langflowInput,
		"session_id":  body.SessionID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error encoding payload: "+err.Error())
		return
	}
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 60 * 2 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, apiConfig.LangflowUrl, bytes.NewReader(payloadBytes))
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error creating request: "+err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiConfig.LangflowApiKey)
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("error making API request: " + err.Error())
		helpers.RespondWithError(w, http.StatusBadGateway, "error making API request: "+err.Error())
		return
	}
	defer resp.Body.Close()
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error reading response: "+err.Error())
		return
	}

	// If Langflow returned an error status, forward an appropriate error message
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Avoid logging or echoing headers that may contain secrets.
		msg := fmt.Sprintf("langflow returned status %d: %s", resp.StatusCode, string(respBody))
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusBadGateway, msg)
		return
	}

	helpers.RespondWithJson(w, http.StatusOK, "langflow analyzed data.")
}

func (apiConfig *Config) GetResultHandler(w http.ResponseWriter, r *http.Request) {
	session_id := chi.URLParam(r, "sessionID")
	sessionId, err := uuid.Parse(session_id)
	if err != nil {
		msg := fmt.Sprintf("error parsing uuid from param. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	result, err := apiConfig.DB.GetResultBySession(r.Context(), sessionId)
	if err != nil {
		msg := fmt.Sprintf("error getting result for session. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	helpers.RespondWithJson(w, 200, result.Result)
}
