package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func (apiConfig *Config) GetResultHandler(w http.ResponseWriter, r *http.Request) {
	session_id := chi.URLParam(r, "sessionID")
	sessionId, err := uuid.Parse(session_id)
	if err != nil {
		msg := fmt.Sprintf("error parsing uuid from param. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	result, err := apiConfig.DB.GetAnalysesResultsBySession(r.Context(), sessionId)
	if err != nil {
		msg := fmt.Sprintf("error getting result for session. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	helpers.RespondWithJson(w, 200, DbAnalysesResultToModelsAnalysesResults(result))
}
