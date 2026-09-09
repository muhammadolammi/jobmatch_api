package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func (cfg *Config) CreateJobPost(w http.ResponseWriter, r *http.Request, user User) {
	body := struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		msg := fmt.Sprintf("error decoding request body. err: %v", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	if body.Title == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "JOB title can't be empty")
		return
	}
	if body.Description == "" {
		helpers.RespondWithError(w, http.StatusInternalServerError, "JOB description can't be empty")
		return
	}
	employer, err := cfg.DB.GetEmployerProfileByUserID(r.Context(), user.ID)
	if err != nil {
		msg := fmt.Sprintf("error getting  employer details body. ", err)
		log.Printf("error getting  employer details body. err: %v\n", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	jobpost, err := cfg.DB.CreateJobPost(r.Context(), database.CreateJobPostParams{
		EmployerID:  employer.ID,
		Title:       body.Title,
		Description: body.Description,
	})
	if err != nil {
		msg := fmt.Sprintf("error creating job post. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return

	}
	helpers.RespondWithJson(w, http.StatusOK, DbJobPostToModelJobPost(jobpost))
}

func (cfg *Config) GetJobPosts(w http.ResponseWriter, r *http.Request, user User) {
	jobPosts, err := cfg.DB.GetJobPosts(r.Context())
	if err != nil {
		msg := fmt.Sprintf("error getting job posts. err: %v", err)
		log.Println(msg)
		helpers.RespondWithError(w, http.StatusInternalServerError, msg)
		return

	}
	helpers.RespondWithJson(w, http.StatusOK, DbJobPostsToModelJobPosts(jobPosts))
}
