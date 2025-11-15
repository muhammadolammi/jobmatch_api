package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func (apiConfig *Config) PostSubscribe(w http.ResponseWriter, r *http.Request, user User) {
	// Check if the user has an active subscription
	exist, err := apiConfig.DB.CheckSubscriptionExist(r.Context(), user.ID)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error checking if subscription exist (db)")
		return
	}

	body := struct {
		IsUpgrade bool   `json:"is_upgrade"`
		PlanCode  string `json:"plan_code"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error decoding req body. err: %v", err))
		return
	}
	if exist {
		// handle existing subscription
		// check if update in request payload.
		if body.IsUpgrade {

			return
		}

		//
		helpers.RespondWithError(w, http.StatusBadRequest, "subscription exist, and no upgrade in payload")
		return
	}
	payload := struct {
		Email  string `json:"email"`
		Plan   string `json:"plan"`
		Amount int    `json:"amount"`
	}{
		Email:  user.Email,
		Plan:   body.PlanCode,
		Amount: 1,
	}
	res := struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	}{}
	paystackRes, err := helpers.CallPaystack(apiConfig.PaystackApi+"/transaction/initialize", "POST", apiConfig.PaystackSecretKey, payload, apiConfig.HttpClient)
	if err != nil {
		log.Printf("error calling paystack. err: %v\n", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error calling paystack. err: %v", err))
		return
	}

	err = json.Unmarshal(paystackRes.Data, &res)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error decoding paystack response data. err: %v", err))
		return
	}
	helpers.RespondWithJson(w, http.StatusOK, map[string]any{
		"Message": "authorization_url generated",
		"Data": map[string]any{
			"authorization_url": res.AuthorizationURL,
		},
	})

}

func (api *Config) PaystackWebhook(w http.ResponseWriter, r *http.Request) {
	event := struct {
		Event string          `json:"event"`
		Data  json.RawMessage `json:"data"`
	}{}

	json.NewDecoder(r.Body).Decode(&event)

	switch event.Event {

	case "charge.success":
		handleChargeSuccess(api, event.Data)

	case "subscription.create":
		handleSubscriptionCreate(api, event.Data)

	// case "invoice.create", "invoice.update", "invoice.payment_failed":
	// 	handleInvoiceEvent(api, event.Event, event.Data)

	case "subscription.disable", "subscription.not_renew":
		handleSubscriptionStatus(api, event.Event, event.Data)
	}
}
