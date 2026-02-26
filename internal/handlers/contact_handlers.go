package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func (cfg *Config) GetContactDepartmentsHandler(w http.ResponseWriter, r *http.Request) {
	contactsDepartments, err := cfg.DB.GetContactDepartments(context.Background())
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error getting contacts departments from db")
		return
	}

	helpers.RespondWithJson(w, http.StatusOK, DbContactDepartmentsToModelContactDepartments(contactsDepartments))

}

func (cfg *Config) PostContactMessagesHandler(w http.ResponseWriter, r *http.Request, body PostContactMessageBody) {
	departmentIdUUid, err := uuid.Parse(body.DepartmentId)
	if err != nil {

		helpers.RespondWithError(w, http.StatusInternalServerError, "error parsing uuid. err: "+err.Error())
		return

	}
	_, err = cfg.DB.CreateContactMessage(context.Background(), database.CreateContactMessageParams{
		FirstName:           body.FirstName,
		LastName:            body.LastName,
		Email:               body.Email,
		ContactDepartmentID: departmentIdUUid,
		Message:             body.Message,
	})
	if err != nil {

		log.Println("db error creating contact message, err: ", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "error creating contact message")
		return

	}
	helpers.RespondWithJson(w, http.StatusOK, "contact message created")
}
