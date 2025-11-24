package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func (apiConfig *Config) HandleGetMySubscription(w http.ResponseWriter, r *http.Request, user User) {
	// 1. Try to get the subscription from DB
	sub, err := apiConfig.DB.GetSubscriptionWithUserID(r.Context(), user.ID)

	// 2. Handle Case: User has never subscribed (Free Tier)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return a specific structure saying "No Plan" / "Free"
			helpers.RespondWithJson(w, http.StatusOK, map[string]any{
				"status":  "free",
				"plan_id": nil,
			})
			return
		}
		// Handle actual DB errors
		log.Println("Error fetching user subscription:", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Error checking subscription status")
		return
	}

	// 3. Handle Case: Subscription exists
	// We respond with the full subscription object so frontend has PlanID and Status
	helpers.RespondWithJson(w, http.StatusOK, DbSubscriptionToModelSubscription(sub))
}
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

func (apiConfig *Config) PaystackWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("request header", r.Header)

	bodyBytes, err := io.ReadAll(r.Body)
	log.Println("request body", string(bodyBytes))

	if err != nil {
		log.Println("Error reading webhook body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	verificationStatus := helpers.VerifyPaystackSignature(bodyBytes, r.Header.Get("x-paystack-signature"), apiConfig.PaystackSecretKey)
	log.Println("verification status", verificationStatus)

	if !verificationStatus {
		log.Println("Unauthorized webhook attempt: Invalid signature")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	event := struct {
		Event string `json:"event"`

		// Fields for subscription/invoice events
		SubscriptionCode string `json:"subscription_code"`

		Data struct {
			NextPaymentDate  string `json:"next_payment_date"`
			SubscriptionCode string `json:"subscription_code"`
			Reference        string `json:"reference"`
			Amount           int    `json:"amount"`
			Status           string `json:"status"`

			Customer struct {
				ID           int    `json:"id"`
				FirstName    string `json:"first_name"`
				LastName     string `json:"last_name"`
				Email        string `json:"email"`
				CustomerCode string `json:"customer_code"`
				Phone        any    `json:"phone"`
				Metadata     any    `json:"metadata"`
				RiskAction   string `json:"risk_action"`
			} `json:"customer"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(bodyBytes, &event); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	ctx := r.Context()

	// get user by email
	user, err := apiConfig.DB.GetUserWithEmail(ctx, event.Data.Customer.Email)
	if err != nil {
		log.Println("error getting user on subscription webhook event")
		return
	}
	log.Println(user)
	// Find subscription by user id
	sub, err := apiConfig.DB.GetSubscriptionWithUserID(ctx, user.ID)
	if err != nil {
		log.Println("webhhok error, error getting user subscription. err: ", err)
		return
	}

	switch event.Event {

	case "subscription.create":
		// Payload has 'next_payment_date' and 'subscription_code'

		// Parse the time string from Paystack
		nextPaymentTime, _ := time.Parse(time.RFC3339, event.Data.NextPaymentDate)

		oldHistoryValue := fmt.Sprintf("Status: %s, NextDate: %v", sub.Status, sub.NextPaymentDate.Time)

		// Update DB: Set Status Active, Save Sub Code, Save Date
		apiConfig.DB.UpdateSubscriptionForActivation(ctx, database.UpdateSubscriptionForActivationParams{
			ID:              sub.ID,
			Status:          "active", // Use "active", not "created" for functional logic
			NextPaymentDate: sql.NullTime{Valid: true, Time: nextPaymentTime},
			PaystackSubCode: sql.NullString{Valid: true, String: event.Data.SubscriptionCode},
		})

		// History
		apiConfig.DB.CreateSubscriptionHistory(ctx, database.CreateSubscriptionHistoryParams{
			SubscriptionID: sub.ID,
			UserID:         sub.UserID,
			EventType:      "subscription_create",
			OldValue:       sql.NullString{String: oldHistoryValue, Valid: true},
			NewValue:       sql.NullString{String: fmt.Sprintf("Status: active, NextDate: %v", nextPaymentTime), Valid: true},
			EventSource:    "webhook",
		})
	case "charge.success":
		// NOTE: The payload provided has NO next_payment_date.
		// It confirms the money was taken.
		// Logic: Ensure subscription is marked active if it wasn't already.

		// If this is a renewal, 'subscription.create' won't fire, so we might need 'invoice.create' or 'invoice.update'
		// to get the new date. But for 'charge.success', we just log the money.

		// Log the successful payment to history
		apiConfig.DB.CreateSubscriptionHistory(ctx, database.CreateSubscriptionHistoryParams{
			SubscriptionID: sub.ID,
			UserID:         sub.UserID,
			EventType:      "charge_success", // Record that money was paid
			OldValue:       sql.NullString{String: "payment_attempt", Valid: true},
			NewValue:       sql.NullString{String: fmt.Sprintf("success | Ref: %s | Amt: %d", event.Data.Reference, event.Data.Amount), Valid: true},
			EventSource:    "webhook",
		})

		// Safety check: Ensure user is active (in case they were 'attention' before)
		if sub.Status != "active" {
			apiConfig.DB.UpdateSubscriptionStatus(ctx, database.UpdateSubscriptionStatusParams{
				ID:     sub.ID,
				Status: "active",
			})
		}
		//    now lets update user max daily usage
		// 1. GET THE PLAN DETAILS to find the DailyLimit
		// We use the PlanID from the subscription we found earlier
		planDetails, err := apiConfig.DB.GetPlan(ctx, sub.PlanID)
		if err != nil {
			log.Println("Webhook Error: could not find plan details to update rate limit", err)
			// Don't return, just log. The subscription is active, but limit update failed.
		} else {
			// 2. UPDATE THE USER USAGE TABLE
			err = apiConfig.DB.UpdateUserDailyUsageLimit(ctx, database.UpdateUserDailyUsageLimitParams{
				UserID:   sub.UserID,
				MaxDaily: planDetails.DailyLimit,
			})
			if err != nil {
				log.Println("Webhook Error: failed to update user rate limit", err)
			}
		}
	case "invoice.payment_failed":
		// Update Status to attention/past_due
		apiConfig.DB.UpdateSubscriptionStatus(ctx, database.UpdateSubscriptionStatusParams{
			ID:     sub.ID,
			Status: "attention",
		})

		apiConfig.DB.CreateSubscriptionHistory(ctx, database.CreateSubscriptionHistoryParams{
			SubscriptionID: sub.ID,
			UserID:         sub.UserID,
			EventType:      "payment_failed",
			OldValue:       sql.NullString{String: "active", Valid: true},
			NewValue:       sql.NullString{String: "attention", Valid: true},
			EventSource:    "webhook",
		})
		// REVERT TO DEFAULT LIMIT (e.g., 2 requests per day)
		defaultLimit := int32(2)

		err = apiConfig.DB.UpdateUserDailyUsageLimit(ctx, database.UpdateUserDailyUsageLimitParams{
			UserID:   sub.UserID,
			MaxDaily: defaultLimit,
		})
		if err != nil {
			log.Printf("Webhook Error: failed to revert user rate limit to default for user %v", sub.UserID)
		}

	case "subscription.disable":
		apiConfig.DB.UpdateSubscriptionStatus(ctx, database.UpdateSubscriptionStatusParams{
			ID:     sub.ID,
			Status: "cancelled",
		})

		apiConfig.DB.CreateSubscriptionHistory(ctx, database.CreateSubscriptionHistoryParams{
			SubscriptionID: sub.ID,
			UserID:         sub.UserID,
			EventType:      "cancelled",
			OldValue:       sql.NullString{String: sub.Status, Valid: true},
			NewValue:       sql.NullString{String: "cancelled", Valid: true},
			EventSource:    "webhook",
		})
		// REVERT TO DEFAULT LIMIT (e.g., 2 requests per day)
		defaultLimit := int32(2)

		err = apiConfig.DB.UpdateUserDailyUsageLimit(ctx, database.UpdateUserDailyUsageLimitParams{
			UserID:   sub.UserID,
			MaxDaily: defaultLimit,
		})
		if err != nil {
			log.Printf("Webhook Error: failed to revert user rate limit to default for user %v", sub.UserID)
		}

	}
	helpers.RespondWithJson(w, http.StatusOK, "")

}
