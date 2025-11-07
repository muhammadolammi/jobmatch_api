package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func (apiConfig *Config) CreateSession(w http.ResponseWriter, r *http.Request, user User) {
	body := struct {
		Name           string `json:"name"`
		JobTitle       string `json:"job_title"`
		JobDescription string `json:"job_description"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		msg := fmt.Sprintf("error decoding request body. err: %v", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	if body.Name == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "session name can't be empty")
		return

	}
	if body.JobTitle == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "session job_title can't be empty")
		return
	}
	if body.JobDescription == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "session job_description can't be empty")
		return
	}
	session, err := apiConfig.DB.CreateSession(r.Context(), database.CreateSessionParams{
		Name:           body.Name,
		UserID:         user.ID,
		JobTitle:       body.JobTitle,
		JobDescription: body.JobDescription,
	})
	if err != nil {
		msg := fmt.Sprintf("error creating session. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return

	}
	helpers.RespondWithJson(w, http.StatusOK, DbSessionToModelsSession(session))
}

func (apiConfig *Config) GetSessions(w http.ResponseWriter, r *http.Request, user User) {
	sessions, err := apiConfig.DB.GetUserSessions(r.Context(), user.ID)
	if err != nil {
		msg := fmt.Sprintf("error getting sessions. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return

	}
	helpers.RespondWithJson(w, http.StatusOK, DbSessionsToModelsSessions(sessions))
}

func (apiConfig *Config) HandleSessionUpdates(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		helpers.RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported")
		return
	}
	// Send a ping to open the stream immediately
	fmt.Fprintf(w, "data: %s\n\n", `{"status":"connected"}`)

	updates := make(chan string)
	// Register this client with a global session broadcaster
	apiConfig.SessionBroadcaster.Register(sessionID, updates)
	defer apiConfig.SessionBroadcaster.Unregister(sessionID, updates)
	// Keep connection open and stream messages
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return // client closed connection
		case msg := <-updates:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}
