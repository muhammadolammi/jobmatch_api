package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func (apiConfig *Config) PostSubscribe(w http.ResponseWriter, r *http.Request, user User) {
	// Check if the user has an active subscription

	body := struct {
		PlanCode string `json:"plan_code"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error decoding req body. err: %v", err))
		return
	}

	dbPlan, err := apiConfig.DB.GetPlanWithPlaneCode(r.Context(), sql.NullString{Valid: true, String: body.PlanCode})
	if err != nil {
		log.Println("error getting plan from db, err: ", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "error retrieving plan")
		return

	}

	// lets check if plan has a valid subscription page
	if !dbPlan.SubscriptionPage.Valid {
		helpers.RespondWithError(w, http.StatusBadRequest, "plan is not configured for subscriptions (missing url)")
		return
	}

	// lets check if user already subscribed
	exist, err := apiConfig.DB.CheckSubscriptionExist(r.Context(), user.ID)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "error checking subscription status")
		return
	}
	if exist {
		userSubscription, err := apiConfig.DB.GetSubscriptionWithUserID(r.Context(), user.ID)
		if err != nil {
			log.Println("user already subscribed, but error getting subscription from db, err: ", err)
			helpers.RespondWithError(w, http.StatusInternalServerError, "error retrieving user subscription")
			return

		}
		if userSubscription.Status == "pending" {
			// Then user hasnt process on paystack subscription page, wwe just return the subscription url
			// lets updste the plan
			if userSubscription.PlanID != dbPlan.ID {
				err = apiConfig.DB.UpdateSubscriptionPlan(r.Context(), database.UpdateSubscriptionPlanParams{
					ID:     userSubscription.ID,
					PlanID: dbPlan.ID,
				})
				if err != nil {
					log.Println("error updating pending subscription plan:", err)
					helpers.RespondWithError(w, http.StatusInternalServerError, "error updating subscription context")
					return
				}
			}
			helpers.RespondWithJson(w, http.StatusOK, map[string]any{
				"Message": "retruning subscription page",
				"Data": map[string]any{
					"subscribe_page": dbPlan.SubscriptionPage.String,
				},
			})
			return
		}

		//
		helpers.RespondWithError(w, http.StatusBadRequest, "user already subscribed")
		return
	}

	// lets add subscription to db
	_, err = apiConfig.DB.CreateSubscription(r.Context(), database.CreateSubscriptionParams{
		UserID: user.ID,
		PlanID: dbPlan.ID,
	})

	if err != nil {
		log.Println("error creating subscription in db, err: ", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "error creating subscription in db")
		return

	}

	helpers.RespondWithJson(w, http.StatusOK, map[string]any{
		"Message": "retruning subscription page",
		"Data": map[string]any{
			"subscribe_page": dbPlan.SubscriptionPage.String,
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

	// case "charge.success":
	// 	handleChargeSuccess(api, event.Data)

	// case "subscription.create":
	// 	handleSubscriptionCreate(api, event.Data)

	// // case "invoice.create", "invoice.update", "invoice.payment_failed":
	// // 	handleInvoiceEvent(api, event.Event, event.Data)

	// case "subscription.disable", "subscription.not_renew":
	// 	handleSubscriptionStatus(api, event.Event, event.Data)
	}
}
