package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

// Middleware to check for the API key in the authorization header for all POST, PUT, DELETE, and OPTIONS requests
func (apiConfig *Config) ClientAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Bypass SSE endpoint and inject Authorization header
			if strings.HasPrefix(r.URL.Path, "/api/sessions/sse") {
				// Get token from query parameter
				accessToken := r.URL.Query().Get("access_token")
				if accessToken == "" {
					helpers.RespondWithError(w, http.StatusUnauthorized, "missing access token")
					return
				}

				// Inject Authorization header for downstream AuthMiddleware
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

				// Continue to next handler
				next.ServeHTTP(w, r)
				return
			}
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

		authJwt, err := jwt.ParseWithClaims(tokenString, authclaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(apiConfig.JwtKey), nil
		})
		if err != nil {
			helpers.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("error parsing jwt claims: %v", err))
			return
		}

		if authclaims.ExpiresAt != nil && authclaims.ExpiresAt.Time.Before(time.Now().UTC()) {
			helpers.RespondWithError(w, http.StatusUnauthorized, "auth token expired")
			return
		}

		userId, err := authJwt.Claims.GetIssuer()
		if err != nil {
			helpers.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("error getting issuer: %v", err))
			return
		}

		id, err := uuid.Parse(userId)
		if err != nil {
			helpers.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("error parsing id: %v", err))
			return
		}

		user, err := apiConfig.DB.GetUser(r.Context(), id)
		if err != nil {
			helpers.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("error getting user: %v", err))
			return
		}
		//serve for admin without rate limiting
		if user.Role == "admin" {
			next(w, r, DbUserToModelsUser(user))
		}
		// ✅ Rate limit check
		usage, err := apiConfig.DB.GetUserUsage(r.Context(), user.ID)
		now := time.Now()

		if err == sql.ErrNoRows {
			_ = apiConfig.DB.InsertUserUsage(r.Context(), database.InsertUserUsageParams{
				UserID:   user.ID,
				MaxDaily: int32(apiConfig.RateLimit),
			})
		} else if err != nil {
			helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error checking usage: %v", err))
			return
		} else {
			if now.Sub(usage.LastUsedAt) < 24*time.Hour {
				if usage.Count >= usage.MaxDaily {
					helpers.RespondWithError(w, http.StatusTooManyRequests, "daily usage limit reached")
					return
				}
				_ = apiConfig.DB.UpdateUserUsageCount(r.Context(), database.UpdateUserUsageCountParams{
					UserID: user.ID,
					Count:  usage.Count + 1,
				})
			} else {
				_ = apiConfig.DB.UpdateUserUsageCount(r.Context(), database.UpdateUserUsageCountParams{
					UserID: user.ID,
					Count:  1,
				})
			}
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
