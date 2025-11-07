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
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:8081", "https://jobmatch.qtechconsults.com"}, // You can customize this based on your needs

		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"}, // You can customize this based on your needs
		AllowCredentials: true,
		MaxAge:           300, // Maximum age for cache, in seconds
	}
	router := chi.NewRouter()
	apiRoute := chi.NewRouter()
	// ADD MIDDLREWARE
	// A good base middleware stack
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(cors.Handler(corsOptions))
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
	apiRoute.Get("/sessions/{id}/updates", apiConfig.HandleSessionUpdates)

	// analyze
	apiRoute.Post("/uploads/complete", apiConfig.AuthMiddleware(apiConfig.UploadCompleteHandler))
	apiRoute.Post("/analyze", apiConfig.AnalyzeHandler)
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
