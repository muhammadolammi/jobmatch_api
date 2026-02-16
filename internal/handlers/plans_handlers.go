package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func (cfg *Config) PostPlanHandler(w http.ResponseWriter, r *http.Request, user User) {
	body := struct {
		Name       string `json:"name"`
		Amount     int    `json:"amount"`
		Currency   string `json:"currency"`
		DailyLimit int32  `json:"daily_limit"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Println("error decoding req body. err: ", err)
		helpers.RespondWithError(w, http.StatusBadRequest, "error decoding req body")
		return
	}
	if body.Name == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "empty plan_name")
		return
	}
	if body.Amount <= 0 {
		helpers.RespondWithError(w, http.StatusBadRequest, "empty amount")
		return
	}
	if body.Currency == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "empty currency")
		return
	}
	dailyLimit := int32(0)
	if body.DailyLimit != 0 {
		dailyLimit = body.DailyLimit
	} else {
		ok, val := helpers.GetPlanDailyUsage(body.Name)
		if !ok {
			helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("daily_limit empty in request body and no record for plan_name(%v) locally", body.Name))
			return
		}
		dailyLimit = val
	}
	// change body currency to upper
	body.Currency = strings.ToUpper(body.Currency)
	// add test to plan in development
	if cfg.ENV == "development" {
		body.Name = body.Name + "-test"
	}
	// SAVE PLAN TO DB FIRST
	dbPlan, err := cfg.DB.CreatePlan(r.Context(), database.CreatePlanParams{
		Name:       body.Name,
		Amount:     int32(helpers.ConvertAmount(body.Amount, body.Currency)),
		DailyLimit: dailyLimit,
		Currency:   body.Currency,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			dbPlan, err = cfg.DB.GetPlanWithName(r.Context(), body.Name)
			if err != nil {
				log.Println("error getting old plan from db. err: ", err)
				helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error saving plan to db response. err: %v", err))
				return
			}
		} else {
			log.Println("error creating plan to db. err: ", err)
			helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error saving plan to db response. err: %v", err))
			return
		}

	}

	// only call paystack if theres no plancode in db
	if !dbPlan.PlanCode.Valid {
		payload := struct {
			Name     string `json:"name"`
			Amount   int64  `json:"amount"`
			Interval string `json:"interval"`
			Currency string `json:"currency"`
		}{
			Interval: "monthly",
			Name:     body.Name,
			Amount:   helpers.ConvertAmount(body.Amount, body.Currency),
			Currency: body.Currency,
		}
		res := struct {
			Name         string    `json:"name"`
			Amount       int       `json:"amount"`
			Interval     string    `json:"interval"`
			Integration  int       `json:"integration"`
			Domain       string    `json:"domain"`
			PlanCode     string    `json:"plan_code"`
			SendInvoices bool      `json:"send_invoices"`
			SendSms      bool      `json:"send_sms"`
			HostedPage   bool      `json:"hosted_page"`
			Currency     string    `json:"currency"`
			ID           int       `json:"id"`
			CreatedAt    time.Time `json:"createdAt"`
			UpdatedAt    time.Time `json:"updatedAt"`
		}{}

		// log.Println("domain:: ", res.Data.Domain)
		paystackRes, err := helpers.CallPaystack(cfg.PaystackApi+"/plan", "POST", cfg.PaystackSecretKey, payload, cfg.HttpClient)
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

		err = cfg.DB.UpdatePlanCode(r.Context(), database.UpdatePlanCodeParams{
			PlanCode: sql.NullString{
				Valid:  true,
				String: res.PlanCode,
			},
			ID: dbPlan.ID,
		})
		if err != nil {
			log.Println("error saving plan to db. err: ", err)
			helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error saving plan to db response. err: %v", err))
			return
		}
	}
	helpers.RespondWithJson(w, http.StatusOK, map[string]any{
		"Message": "plan created",
	})

}

func (cfg *Config) GetPlansHandler(w http.ResponseWriter, r *http.Request, user User) {
	plans, err := cfg.DB.GetPlans(r.Context())
	if err != nil {
		log.Println("db error on get plans. err: ", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "db error on get plans")
		return
	}
	helpers.RespondWithJson(w, http.StatusOK, map[string]any{
		"Message": "plans returned succesfully",
		"Data":    DbPlansToModelPlans(plans),
	})

}

func (cfg *Config) PostPlanSubPageHandler(w http.ResponseWriter, r *http.Request, user User) {
	body := struct {
		PlanID           uuid.UUID `json:"plan_id"`
		SubscriptionPage string    `json:"subscription_page"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Println("error decoding req body. err: ", err)
		helpers.RespondWithError(w, http.StatusBadRequest, "error decoding req body")
		return
	}
	if body.PlanID == uuid.Nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "empty plan_id")
		return
	}
	if body.SubscriptionPage == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "empty subscription_page")
		return
	}
	err = cfg.DB.UpdatePLanSubscriptionPage(r.Context(), database.UpdatePLanSubscriptionPageParams{
		SubscriptionPage: sql.NullString{
			Valid:  true,
			String: body.SubscriptionPage,
		},
		ID: body.PlanID,
	})
	if err != nil {
		log.Println("error updating  plan subscription page. err: ", err)
		helpers.RespondWithError(w, http.StatusBadRequest, "error updating plan context")
		return
	}
	helpers.RespondWithJson(w, http.StatusOK, "plan subscription page updated")
}
