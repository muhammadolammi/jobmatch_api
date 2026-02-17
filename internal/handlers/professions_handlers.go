package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func (cfg *Config) PostUserProfessionsHandler(w http.ResponseWriter, r *http.Request, user User) {
	body := struct {
		ProfessionID string `json:"profession_id"`
		UserID       string `json:"user_id"`
	}{}
	encoder := json.NewDecoder(r.Body)
	err := encoder.Decode(&body)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error decoding request body. err: "+err.Error())
		return
	}
	if body.ProfessionID == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "profession_id can't be empty ")
		return
	}
	if body.UserID == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "user_id can't be empty ")
		return
	}
	userIDUUID, err := uuid.Parse(body.UserID)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error parsing user_id to uuid. err: "+err.Error())
		return
	}
	professionIdUUid, err := uuid.Parse(body.ProfessionID)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error parsing profession_id to uuid. err: "+err.Error())
		return
	}
	userProfession, err := cfg.DB.CreateUserProfession(r.Context(), database.CreateUserProfessionParams{
		ProfessionID: professionIdUUid,
		UserID:       userIDUUID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error creating profession in db. err: "+err.Error())
		return
	}
	helpers.RespondWithJson(w, http.StatusCreated, DbUserProfessionToModelUserProfession(userProfession))
}

func (cfg *Config) GetUserProfessionsHandler(w http.ResponseWriter, r *http.Request, user User) {
	userProfessions, err := cfg.DB.GetUserProfessions(r.Context(), user.ID)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error fetching professions from db. err: "+err.Error())
		return
	}
	helpers.RespondWithJson(w, http.StatusOK, DbUserProfessionsToModelUserProfessions(userProfessions))
}

func (cfg *Config) DeleteUserProfessionHandler(w http.ResponseWriter, r *http.Request, user User) {
	body := struct {
		ProfessionID string `json:"profession_id"`
		UserID       string `json:"user_id"`
	}{}
	encoder := json.NewDecoder(r.Body)
	err := encoder.Decode(&body)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error decoding request body. err: "+err.Error())
		return
	}
	if body.ProfessionID == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "profession_id can't be empty ")
		return
	}
	if body.UserID == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "user_id can't be empty ")
		return
	}
	userIDUUID, err := uuid.Parse(body.UserID)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error parsing user_id to uuid. err: "+err.Error())
		return
	}
	professionIdUUid, err := uuid.Parse(body.ProfessionID)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error parsing profession_id to uuid. err: "+err.Error())
		return
	}
	err = cfg.DB.DeleteUserProfession(r.Context(), database.DeleteUserProfessionParams{
		ProfessionID: professionIdUUid,
		UserID:       userIDUUID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error deleting profession in db. err: "+err.Error())
		return
	}
	helpers.RespondWithJson(w, http.StatusOK, map[string]string{"message": "profession deleted successfully"})
}

func (cfg *Config) PostProfessionsHandler(w http.ResponseWriter, r *http.Request) {
	body := struct {
		Name string `json:"name"`
	}{}
	encoder := json.NewDecoder(r.Body)
	err := encoder.Decode(&body)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error decoding request body. err: "+err.Error())
		return
	}
	if body.Name == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "name can't be empty ")
		return
	}
	profession, err := cfg.DB.CreateProfession(r.Context(), body.Name)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error creating profession in db. err: "+err.Error())
		return
	}
	helpers.RespondWithJson(w, http.StatusCreated, DbProfessionToModelProfession(profession))
}

func (cfg *Config) GetProfessionsHandler(w http.ResponseWriter, r *http.Request) {
	professions, err := cfg.DB.GetProfessions(r.Context())
	if err != nil {
		log.Println("error fetching professions from db. err: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "db error ")
		return
	}
	helpers.RespondWithJson(w, http.StatusOK, DbProfessionsToModelProfessions(professions))
}
