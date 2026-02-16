package handlers

import (
	"context"
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
			// if r.Method == http.MethodOptions {
			// 	next.ServeHTTP(w, r)
			// 	return
			// }
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
			// Bypass Paystack webhook
			if strings.HasPrefix(r.URL.Path, "/api/webhook/paystack") {
				// TODO handle paystack athorization
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
			log.Println("aith error here")
			helpers.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid token")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		authclaims := &jwt.RegisteredClaims{}

		authJwt, err := jwt.ParseWithClaims(tokenString, authclaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(apiConfig.JwtKey), nil
		})
		if err != nil || !authJwt.Valid {
			helpers.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		if authclaims.ExpiresAt != nil && authclaims.ExpiresAt.Time.Before(time.Now().UTC()) {
			helpers.RespondWithError(w, http.StatusUnauthorized, "Token expired")
			return
		}

		userId, err := authJwt.Claims.GetIssuer()
		if err != nil {
			helpers.RespondWithError(w, http.StatusUnauthorized, "Invalid token issuer")
			return
		}

		id, err := uuid.Parse(userId)
		if err != nil {
			helpers.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
			return
		}

		user, err := apiConfig.DB.GetUser(r.Context(), id)
		if err != nil {
			helpers.RespondWithError(w, http.StatusUnauthorized, "User not found")
			return
		}

		ctx := context.WithValue(r.Context(), "user", DbUserToModelUser(user))
		next(w, r.WithContext(ctx), DbUserToModelUser(user))
	})
}
func (apiConfig *Config) RateLimiter(next func(http.ResponseWriter, *http.Request, User)) http.HandlerFunc {
	return apiConfig.AuthMiddleware(func(w http.ResponseWriter, r *http.Request, user User) {
		if user.Role == "admin" {
			next(w, r, user)
			return
		}
		now := time.Now()
		usage, err := apiConfig.DB.GetUserUsage(r.Context(), user.ID)
		switch {
		// ✅ First-time usage
		case err == sql.ErrNoRows:
			err = apiConfig.DB.InsertUserUsage(r.Context(), database.InsertUserUsageParams{
				UserID:     user.ID,
				MaxDaily:   int32(apiConfig.RateLimit),
				Count:      1,
				LastUsedAt: now,
			})
			if err != nil {
				helpers.RespondWithError(w, http.StatusInternalServerError, "failed to initialize usage")
				return
			}

		// ✅ Some other DB error
		case err != nil:
			helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error checking usage: %v", err))
			return

		// ✅ Normal usage flow
		default:
			if now.Sub(usage.LastUsedAt) < 24*time.Hour {
				if usage.Count >= usage.MaxDaily {
					remaining := 24*time.Hour - now.Sub(usage.LastUsedAt)
					helpers.RespondWithJson(w, http.StatusTooManyRequests, map[string]any{
						"error":             "daily_usage_limit_reached",
						"message":           "Daily usage limit reached",
						"remaining_seconds": int(remaining.Seconds()),
					})
					return
				}

				_ = apiConfig.DB.UpdateUserUsage(r.Context(), database.UpdateUserUsageParams{
					UserID:     user.ID,
					Count:      usage.Count + 1,
					LastUsedAt: now,
				})
			} else {
				// new day
				_ = apiConfig.DB.UpdateUserUsage(r.Context(), database.UpdateUserUsageParams{
					UserID:     user.ID,
					Count:      1,
					LastUsedAt: now,
				})
			}
		}

		// ✅ Only reach here if allowed
		next(w, r, user)
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
