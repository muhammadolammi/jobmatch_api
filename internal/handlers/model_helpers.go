package handlers

import (
	"encoding/json"

	"github.com/muhammadolammi/jobmatchapi/internal/database"
)

// User model helpers
func DbUserToModelUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		Role:      dbUser.Role,
		CreatedAt: dbUser.CreatedAt,
	}

}

func DbUsersToModelUsers(dbUsers []database.User) []User {
	users := []User{}
	for _, dbUser := range dbUsers {
		users = append(users, DbUserToModelUser(dbUser))
	}
	return users
}

// Session model helpers
func DbSessionToModelSession(dbSession database.Session) Session {
	return Session{
		ID:             dbSession.ID,
		Name:           dbSession.Name,
		CreatedAt:      dbSession.CreatedAt,
		UserID:         dbSession.UserID,
		Status:         dbSession.Status,
		JobTitle:       dbSession.JobTitle,
		JobDescription: dbSession.JobDescription,
	}

}

func DbSessionsToModelSessions(dbSessions []database.Session) []Session {
	sessions := []Session{}
	for _, dbSession := range dbSessions {
		sessions = append(sessions, DbSessionToModelSession(dbSession))
	}
	return sessions
}

// AnalysesResult model helpers
func DbAnalysesResultToModelAnalysesResults(dbAnalysesResults database.AnalysesResult) AnalysesResults {
	results := []AnalysesResult{}
	json.Unmarshal(dbAnalysesResults.Results, &results)
	return AnalysesResults{
		ID:        dbAnalysesResults.ID,
		Results:   results,
		SessionID: dbAnalysesResults.SessionID,
		CreatedAt: dbAnalysesResults.CreatedAt,
		UpdatedAt: dbAnalysesResults.UpdatedAt,
	}

}

// Session model helpers
func DbPlanToModelPlan(dbPlan database.Plan) Plan {
	return Plan{
		ID:          dbPlan.ID,
		Name:        dbPlan.Name,
		CreatedAt:   dbPlan.CreatedAt,
		Amount:      dbPlan.Amount,
		Currency:    dbPlan.Currency,
		UpdatedAt:   dbPlan.UpdatedAt,
		Description: dbPlan.Description,
		DailyLimit:  dbPlan.DailyLimit,
		PlanCode:    dbPlan.PlanCode.String,
		Interval:    dbPlan.Interval,
	}

}

func DbPlansToModelPlans(dbPlans []database.Plan) []Plan {
	plans := []Plan{}
	for _, dbPlan := range dbPlans {
		plans = append(plans, DbPlanToModelPlan(dbPlan))
	}
	return plans
}
