package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func (apiConfig *Config) CreateSession(w http.ResponseWriter, r *http.Request, user User) {
	body := struct {
		Name string `json:"name"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		msg := fmt.Sprintf("error decoding request body. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	session, err := apiConfig.DB.CreateSession(r.Context(), database.CreateSessionParams{
		Name:   body.Name,
		UserID: user.ID,
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
