package main

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
)

func helloReady(w http.ResponseWriter, r *http.Request) { respondWithJson(w, 200, "hello") }
func errorReady(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 200, "this is an error test")
}

func (apiConfig *Config) uploadHandler(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, 20<<20)
	// Parse up to 10MB of file parts kept in memory before writing to temp files
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}
	sessionID := r.FormValue("session_id")
	if sessionID == "" {
		respondWithError(w, http.StatusBadRequest, "Missing session_id in form data")
		return
	}
	result, err := apiConfig.DB.GetResultBySession(r.Context(), sessionID)
	resultFound := false
	if err != nil {
		if !(err == sql.ErrNoRows) {
			//  lets handle error that's not empty row
			msg := fmt.Sprintf("Error check result, err: %v", err)

			log.Println(msg)
			respondWithError(w, http.StatusInternalServerError, msg)
			return
		}

	}
	if result.ID != uuid.Nil {
		resultFound = true
	}
	if resultFound {
		// 	if time.Since(result.CreatedAt) < 24*time.Hour {

		if time.Since(result.CreatedAt) < 1*time.Hour {

			remaining := 1*time.Hour - time.Since(result.CreatedAt)
			msg := fmt.Sprintf("You have already analyzed a resume recently. Please wait %.0f minutes before trying again.", remaining.Minutes())
			respondWithJson(w, http.StatusTooManyRequests, map[string]any{
				"error":            "rate_limit_exceeded",
				"message":          msg,
				"retry_in_minutes": int(remaining.Minutes()),
			})
			return
		}
	}
	files := r.MultipartForm.File["file"]
	if len(files) == 0 {

		respondWithError(w, http.StatusBadRequest, "No file uploaded")
		return
	}

	if len(files) > 1 {

		respondWithError(w, http.StatusBadRequest, "Only one file allowed")
		return
	}
	// Step 4.1: get the original filename
	fileHeader := files[0]
	filename := fileHeader.Filename
	filename = filepath.Base(filename)

	log.Println("Processing file:", filename)

	// Step 4.2: open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		msg := fmt.Sprintf("Error opening file: %v, err: %v", filename, err)
		log.Println(msg)

		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	allowed := map[string]bool{".pdf": true, ".docx": true, ".txt": true}
	if !allowed[ext] {
		msg := fmt.Sprintf("Invalid file type: %v", filename)
		log.Println(msg)
		respondWithError(w, http.StatusInternalServerError, msg)
		return

	}

	data, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		msg := fmt.Sprintf("Error reading file:%v , err: %v", filename, err)
		log.Println(msg)
		respondWithError(w, http.StatusInternalServerError, msg)
		return

	}

	log.Printf("Read %d bytes from %s\n", len(data), filename)
	//  save data to db
	encoded := base64.StdEncoding.EncodeToString(data)

	_, err = apiConfig.DB.CreateResume(r.Context(), database.CreateResumeParams{
		FileName:  filename,
		Text:      encoded,
		SessionID: sessionID,
	})
	if err != nil {
		msg := fmt.Sprintf("error uploading files. filename: %v, err: %v\n", filename, err)
		log.Println(msg)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	respondWithJson(w, http.StatusOK, map[string]string{
		"message": "Upload successful",
	})

}

func (apiConfig *Config) analyzeHandler(w http.ResponseWriter, r *http.Request) {
	body := struct {
		SessionID      string `json:"session_id"`
		JobTitle       string `json:"job_title"`
		JobDescription string `json:"job_description"`
	}{}
	encoder := json.NewDecoder(r.Body)
	err := encoder.Decode(&body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error decoding request body. err: "+err.Error())
		return
	}
	if body.SessionID == "" {
		respondWithError(w, http.StatusInternalServerError, "session_id can't be empty ")
		return

	}
	if body.JobDescription == "" {
		respondWithError(w, http.StatusInternalServerError, "job_description can't be empty")
		return

	}
	if body.JobTitle == "" {
		respondWithError(w, http.StatusInternalServerError, "job_title can't be empty")
		return

	}
	langflowInput := fmt.Sprintf("Job Title \n %s\n Job Description\n%s\n", body.JobTitle, body.JobDescription)
	payload := map[string]interface{}{
		"input_type":  "chat",
		"output_tye":  "chat",
		"input_value": langflowInput,
		"session_id":  body.SessionID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error encoding payload: "+err.Error())
		return
	}
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 60 * 2 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, apiConfig.LANGFLOW_URL, bytes.NewReader(payloadBytes))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating request: "+err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiConfig.LANGFLOW_API_KEY)
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("error making API request: " + err.Error())
		respondWithError(w, http.StatusBadGateway, "error making API request: "+err.Error())
		return
	}
	defer resp.Body.Close()
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error reading response: "+err.Error())
		return
	}

	// If Langflow returned an error status, forward an appropriate error message
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Avoid logging or echoing headers that may contain secrets.
		msg := fmt.Sprintf("langflow returned status %d: %s", resp.StatusCode, string(respBody))
		log.Println(msg)
		respondWithError(w, http.StatusBadGateway, msg)
		return
	}

	respondWithJson(w, http.StatusOK, "langflow analyzed data.")
}

func (apiConfig *Config) getResultHandler(w http.ResponseWriter, r *http.Request) {
	session_id := chi.URLParam(r, "sessionID")

	result, err := apiConfig.DB.GetResultBySession(r.Context(), session_id)
	if err != nil {
		msg := fmt.Sprintf("error getting result for session. err: %v", err)
		log.Println(msg)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	respondWithJson(w, 200, result.Result)
}
