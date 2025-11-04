package handlers

import "github.com/muhammadolammi/jobmatchapi/internal/database"

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
