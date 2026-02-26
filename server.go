package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/muhammadolammi/jobmatchapi/internal/handlers"
)

func server(apiConfig *handlers.Config) {

	// Define CORS options
	corsOptions := cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:8081", "https://gojobmatch.com", "https://jobmatch-backend-755404739186.us-east1.run.app"}, // You can customize this based on your needs

		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"client-api-key",
			"X-CSRF-Token",
			"x-paystack-signature",
			"Accept",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}
	router := chi.NewRouter()
	apiRoute := chi.NewRouter()
	router.Use(cors.Handler(corsOptions))
	// ADD MIDDLREWARE
	// A good base middleware stack
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(apiConfig.ClientAuth())

	// ADD ROUTES
	apiRoute.Get("/hello", handlers.HelloReady)
	apiRoute.Get("/error", handlers.ErrorReady)
	// auth
	apiRoute.Get("/me", apiConfig.AuthMiddleware(apiConfig.GetUserHandler))
	apiRoute.Post("/login", apiConfig.LoginHandler)
	apiRoute.Post("/register", apiConfig.RegisterHandler)
	apiRoute.Post("/refresh", apiConfig.RefreshTokens)

	// session
	apiRoute.Post("/sessions", apiConfig.AuthMiddleware(apiConfig.CreateSession))
	apiRoute.Post("/sessions/{id}/presign", apiConfig.AuthMiddleware(apiConfig.PresignUploadHandler))
	apiRoute.Get("/sessions", apiConfig.AuthMiddleware(apiConfig.GetSessions))
	apiRoute.Get("/sessions/{id}", apiConfig.AuthMiddleware(apiConfig.GetSession))

	apiRoute.Get("/sessions/sse/{id}/updates", apiConfig.AuthMiddleware(apiConfig.HandleSessionUpdates))

	// analyze
	apiRoute.Post("/uploads/complete", apiConfig.AuthMiddleware(apiConfig.UploadCompleteHandler))
	apiRoute.Post("/analyze", apiConfig.AnalyzeRateLimiter(apiConfig.AnalyzeHandler))

	// plans & subscription
	apiRoute.Post("/plans", apiConfig.RoleMiddleware([]string{"admin"}, apiConfig.PostPlanHandler))
	apiRoute.Post("/plans/subpage", apiConfig.RoleMiddleware([]string{"admin"}, apiConfig.PostPlanSubPageHandler))

	apiRoute.Get("/plans", apiConfig.GetPlansHandler)

	apiRoute.Post("/subscribe", apiConfig.AuthMiddleware(apiConfig.PostSubscribe))
	apiRoute.Get("/subscription/me", apiConfig.AuthMiddleware(apiConfig.HandleGetMySubscription))

	// webhooks
	apiRoute.Post("/webhook/paystack", apiConfig.PaystackWebhook)

	// professions
	apiRoute.Post("/professions", apiConfig.PostProfessionsHandler)
	apiRoute.Get("/professions", apiConfig.GetProfessionsHandler)

	apiRoute.Post("/user-professions", apiConfig.AuthMiddleware(apiConfig.PostUserProfessionsHandler))
	apiRoute.Get("/user-professions", apiConfig.AuthMiddleware(apiConfig.GetUserProfessionsHandler))
	apiRoute.Delete("/user-professions", apiConfig.AuthMiddleware(apiConfig.DeleteUserProfessionHandler))

	//  contacts
	apiRoute.Get("/contact-departments", apiConfig.GetContactDepartmentsHandler)
	apiRoute.Post("/contact", apiConfig.ContactRateLimiter(apiConfig.PostContactMessagesHandler))

	apiRoute.Get("/results/{sessionID}", apiConfig.GetResultHandler)
	router.Mount("/api", apiRoute)
	srv := &http.Server{
		Addr:              ":" + apiConfig.Port,
		Handler:           router,
		ReadHeaderTimeout: time.Minute,
	}

	log.Printf("Serving on port: %s\n", apiConfig.Port)
	log.Fatal(srv.ListenAndServe())
}
