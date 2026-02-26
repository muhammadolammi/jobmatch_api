package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
)

// Middleware to check for the API key in the authorization header for all POST, PUT, DELETE, and OPTIONS requests
func (cfg *Config) ClientAuth() func(http.Handler) http.Handler {
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
			if clientApiKey != cfg.ClientApiKey {
				log.Println("invalid client api key in request.")
				helpers.RespondWithError(w, http.StatusUnauthorized, "invalid client api key in request.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (cfg *Config) AuthMiddleware(next func(http.ResponseWriter, *http.Request, User)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			helpers.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid token")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		authclaims := &jwt.RegisteredClaims{}

		authJwt, err := jwt.ParseWithClaims(tokenString, authclaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtKey), nil
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

		user, err := cfg.DB.GetUser(r.Context(), id)
		if err != nil {
			helpers.RespondWithError(w, http.StatusUnauthorized, "User not found")
			return
		}

		ctx := context.WithValue(r.Context(), "user", DbUserToModelUser(user))
		next(w, r.WithContext(ctx), DbUserToModelUser(user))
	})
}
func (cfg *Config) AnalyzeRateLimiter(next func(http.ResponseWriter, *http.Request, User)) http.HandlerFunc {
	return cfg.AuthMiddleware(func(w http.ResponseWriter, r *http.Request, user User) {
		if user.Role == "admin" {
			next(w, r, user)
			return
		}
		now := time.Now()
		usage, err := cfg.DB.GetUserUsage(r.Context(), user.ID)
		switch {
		// ✅ First-time usage
		case err == sql.ErrNoRows:
			err = cfg.DB.InsertUserUsage(r.Context(), database.InsertUserUsageParams{
				UserID:     user.ID,
				MaxDaily:   int32(cfg.RateLimit),
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

				_ = cfg.DB.UpdateUserUsage(r.Context(), database.UpdateUserUsageParams{
					UserID:     user.ID,
					Count:      usage.Count + 1,
					LastUsedAt: now,
				})
			} else {
				// new day
				_ = cfg.DB.UpdateUserUsage(r.Context(), database.UpdateUserUsageParams{
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

func (cfg *Config) RoleMiddleware(allowedRoles []string, next func(http.ResponseWriter, *http.Request, User)) http.HandlerFunc {
	return cfg.AuthMiddleware(func(w http.ResponseWriter, r *http.Request, user User) {
		for _, role := range allowedRoles {
			if user.Role == role {
				next(w, r, user)
				return
			}
		}
		helpers.RespondWithError(w, http.StatusForbidden, "Forbidden: insufficient permissions")
	})
}

func (cfg *Config) ContactRateLimiter(next func(http.ResponseWriter, *http.Request, PostContactMessageBody)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := PostContactMessageBody{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&body)
		if err != nil {
			log.Println("here")
			log.Println(err)
			log.Println(body)

			helpers.RespondWithError(w, http.StatusBadRequest, "error decoding request body")
			return
		}
		if body.Message == "" {
			log.Println("here")

			helpers.RespondWithError(w, http.StatusBadRequest, "contact message can't be empty")
			return

		}
		if body.FirstName == "" {
			log.Println("here")

			helpers.RespondWithError(w, http.StatusBadRequest, "contact first name can't be empty")
			return

		}

		if body.LastName == "" {
			log.Println("here")

			helpers.RespondWithError(w, http.StatusBadRequest, "contact last name can't be empty")
			return

		}
		if body.Email == "" {
			log.Println("here")

			helpers.RespondWithError(w, http.StatusBadRequest, "contact email can't be empty")
			return

		}
		if utf8.RuneCountInString(body.Message) > 500 {
			log.Println("here")

			helpers.RespondWithError(w, http.StatusBadRequest, "message too long")
			return
		}
		last24HourMessages, err := cfg.DB.GetEmailLastHourContactMessages(context.Background(), body.Email)
		if err != nil {
			helpers.RespondWithError(w, http.StatusBadRequest, "error validating request")
			return
		}
		if last24HourMessages >= 10 {
			helpers.RespondWithError(w, http.StatusTooManyRequests, "daily limit reached")
			return

		}
		lastHourMessages, err := cfg.DB.GetEmailLastHourContactMessages(context.Background(), body.Email)
		if err != nil {
			helpers.RespondWithError(w, http.StatusBadRequest, "error validating request")
			return
		}
		if lastHourMessages >= 4 {
			helpers.RespondWithError(w, http.StatusTooManyRequests, "too many messages this hour")
			return

		}
		next(w, r, body)

	})
}
