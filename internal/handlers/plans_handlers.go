package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

func (apiConfig *Config) PostPlanHandler(w http.ResponseWriter, r *http.Request, user User) {
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
	if body.Amount == 0 {
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
	paystackRes, err := helpers.CallPaystack(apiConfig.PaystackApi+"/plan", "POST", apiConfig.PaystackSecretKey, payload, apiConfig.HttpClient)
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

	_, err = apiConfig.DB.CreatePlan(r.Context(), database.CreatePlanParams{
		Name:       body.Name,
		PlanCode:   res.PlanCode,
		Amount:     int32(body.Amount),
		DailyLimit: dailyLimit,
		Currency:   body.Currency,
	})
	if err != nil {
		log.Println("error saving plan to db. err: ", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error saving plan to db response. err: %v", err))
		return
	}
	helpers.RespondWithJson(w, http.StatusOK, map[string]string{
		"messgae": "plan created",
	})
}

func (apiConfig *Config) GetPlansHandler(w http.ResponseWriter, r *http.Request, user User) {
	plans, err := apiConfig.DB.GetPlans(r.Context())
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
