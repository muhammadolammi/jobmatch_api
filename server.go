package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Middleware to check for the API key in the authorization header for all POST, PUT, DELETE, and OPTIONS requests
func (apiConfig *Config) clientAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authToken := r.Header.Get("Authorization")
			if authToken == "" {
				log.Println("empty auth token in request.")
				respondWithError(w, http.StatusUnauthorized, "empty auth token in request.")
				return
			}
			if authToken != apiConfig.AUTH_TOKEN {
				log.Println("invalid auth token in request.")
				respondWithError(w, http.StatusUnauthorized, "invalid auth token in request.")
				return

			}

			next.ServeHTTP(w, r)
		})
	}
}
func server(apiConfig *Config) {

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
	router.Use(apiConfig.clientAuth())

	// ADD ROUTES
	apiRoute.Get("/hello", helloReady)
	apiRoute.Get("/error", errorReady)

	apiRoute.Post("/upload", apiConfig.uploadHandler)
	apiRoute.Post("/analyze", apiConfig.analyzeHandler)
	apiRoute.Get("/results/{sessionID}", apiConfig.getResultHandler)

	router.Mount("/api", apiRoute)
	srv := &http.Server{
		Addr:              ":" + apiConfig.PORT,
		Handler:           router,
		ReadHeaderTimeout: time.Minute,
	}

	log.Printf("Serving on port: %s\n", apiConfig.PORT)
	log.Fatal(srv.ListenAndServe())
}
