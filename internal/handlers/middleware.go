package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

// Middleware to check for the API key in the authorization header for all POST, PUT, DELETE, and OPTIONS requests
func (apiConfig *Config) ClientAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientApiKey := r.Header.Get("client-api-key")
			if clientApiKey == "" {
				log.Println("empty client api key  in request.")
				helpers.RespondWithError(w, http.StatusUnauthorized, "empty client api key in request.")
				return
			}
			if clientApiKey != apiConfig.ClientApiKey {
				log.Println("invalid client api key in request.")
				helpers.RespondWithError(w, http.StatusUnauthorized, "invalid client api key in request.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (apiConfig *Config) AuthMiddleware(next func(http.ResponseWriter, *http.Request, User)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			helpers.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid token")
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		authclaims := &jwt.RegisteredClaims{}
		authJwt, err := jwt.ParseWithClaims(
			tokenString,
			authclaims,
			func(token *jwt.Token) (interface{}, error) { return []byte(apiConfig.JwtKey), nil },
		)

		if err != nil {
			helpers.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("error parsing jwt claims, err: %v", err))
			return
		}

		if authclaims.ExpiresAt != nil && authclaims.ExpiresAt.Time.Before(time.Now().UTC()) {
			helpers.RespondWithError(w, http.StatusUnauthorized, "auth token expired")
			return
		}

		userId, err := authJwt.Claims.GetIssuer()
		if err != nil {
			helpers.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("error getting issuer from jwt claims, err: %v", err))
			return
		}
		id, err := uuid.Parse(userId)
		if err != nil {
			helpers.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("error parsing id, err: %v", err))
			return
		}
		user, err := apiConfig.DB.GetUser(r.Context(), id)
		if err != nil {
			helpers.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("error getting user, err: %v", err))
			return
		}
		next(w, r, DbUserToModelsUser(user))
	})
}

func (apiConfig *Config) RoleMiddleware(allowedRoles []string, next func(http.ResponseWriter, *http.Request, User)) http.HandlerFunc {
	return apiConfig.AuthMiddleware(func(w http.ResponseWriter, r *http.Request, user User) {
		for _, role := range allowedRoles {
			if user.Role == role {
				next(w, r, user)
				return
			}
		}
		helpers.RespondWithError(w, http.StatusForbidden, "Forbidden: insufficient permissions")
	})
}
