package handlers

import (
	"encoding/json"

	"github.com/muhammadolammi/jobmatchapi/internal/database"
)

// User model helpers
func DbUserToModelsUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		Role:      dbUser.Role,
		CreatedAt: dbUser.CreatedAt,
	}

}

func DbUsersToModelsUsers(dbUsers []database.User) []User {
	users := []User{}
	for _, dbUser := range dbUsers {
		users = append(users, DbUserToModelsUser(dbUser))
	}
	return users
}

// Session model helpers
func DbSessionToModelsSession(dbSession database.Session) Session {
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

func DbSessionsToModelsSessions(dbSessions []database.Session) []Session {
	sessions := []Session{}
	for _, dbSession := range dbSessions {
		sessions = append(sessions, DbSessionToModelsSession(dbSession))
	}
	return sessions
}

// User model helpers
func DbAnalysesResultToModelsAnalysesResults(dbAnalysesResults database.AnalysesResult) AnalysesResults {
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
