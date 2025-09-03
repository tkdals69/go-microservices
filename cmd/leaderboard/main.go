package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tkdals69/go-microservices/pkg/config"
	"github.com/tkdals69/go-microservices/pkg/handlers"
	"github.com/tkdals69/go-microservices/pkg/middleware"
	"github.com/tkdals69/go-microservices/pkg/observability"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize router
	r := chi.NewRouter()

	// Initialize middlewares properly
	hmacMw := middleware.NewHMACMiddleware(cfg.HMACSecret)
	r.Use(hmacMw.Verify)

	idempMw := middleware.NewIdempotencyKeyMiddleware()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			idempMw.ServeHTTP(w, req, func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		})
	})

	rl := middleware.NewRateLimiter(cfg.RateLimitMax, cfg.RateLimitMax*2)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if rl.Allow() {
				next.ServeHTTP(w, req)
			} else {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			}
		})
	})

	// Health check endpoint
	r.Get("/healthz", observability.HealthCheck)

	// Metrics endpoint
	r.Get("/metrics", observability.MetricsHandler().ServeHTTP)

	// Initialize leaderboard handler and register routes
	leaderboardHandler := handlers.NewLeaderboardHandler()
	r.Mount("/leaderboard", leaderboardHandler.Routes())

	port := cfg.Port
	if port == "" {
		port = "8082"
	}

	// Start server
	log.Printf("Starting leaderboard service on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
