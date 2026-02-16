package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func (cfg *Config) CreateSession(w http.ResponseWriter, r *http.Request, user User) {
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
	session, err := cfg.DB.CreateSession(r.Context(), database.CreateSessionParams{
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
	helpers.RespondWithJson(w, http.StatusOK, DbSessionToModelSession(session))
}

func (cfg *Config) GetSessions(w http.ResponseWriter, r *http.Request, user User) {
	sessions, err := cfg.DB.GetUserSessions(r.Context(), user.ID)
	if err != nil {
		msg := fmt.Sprintf("error getting sessions. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return

	}
	helpers.RespondWithJson(w, http.StatusOK, DbSessionsToModelSessions(sessions))
}

func (cfg *Config) HandleSessionUpdates(w http.ResponseWriter, r *http.Request, user User) {
	sessionID := chi.URLParam(r, "id")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Consume RabbitMQ queue
	ch, err := cfg.RabbitConn.Channel()
	if err != nil {
		http.Error(w, "Failed to connect to RabbitMQ", http.StatusInternalServerError)
		return
	}
	defer ch.Close()

	// Declare topic exchange
	err = ch.ExchangeDeclare(
		"session_updates",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		http.Error(w, "Failed to declare exchange", http.StatusInternalServerError)
		return
	}
	// Create a temporary queue for this SSE connection
	q, err := ch.QueueDeclare(
		"",    // empty name = let RabbitMQ generate unique name
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false,
		nil,
	)
	if err != nil {
		http.Error(w, "Failed to declare queue", http.StatusInternalServerError)
		return
	}

	// Bind queue to this session's routing key
	routingKey := fmt.Sprintf("session.%s", sessionID)
	err = ch.QueueBind(q.Name, routingKey, "session_updates", false, nil)
	if err != nil {
		http.Error(w, "Failed to bind queue", http.StatusInternalServerError)
		return
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		http.Error(w, "Failed to consume queue", http.StatusInternalServerError)
		return
	}

	// log.Println("logging sse")
	ctx := r.Context()
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C: // NEW: Send a heartbeat comment
			fmt.Fprintf(w, ":keep-alive\n\n")
			flusher.Flush()
		case d := <-msgs:
			// log.Println(string(d.Body))
			fmt.Fprintf(w, "data: %s\n\n", d.Body)
			flusher.Flush()
			if strings.Contains(string(d.Body), "completed") {
				return
			}
		}
	}
}

func (cfg *Config) GetSession(w http.ResponseWriter, r *http.Request, user User) {
	sessionIDString := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(sessionIDString)
	if err != nil {
		msg := fmt.Sprintf("error parsing session id. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return

	}

	session, err := cfg.DB.GetSession(r.Context(), sessionID)
	if err != nil {
		msg := fmt.Sprintf("error getting session. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return

	}
	helpers.RespondWithJson(w, http.StatusOK, DbSessionToModelSession(session))
}
