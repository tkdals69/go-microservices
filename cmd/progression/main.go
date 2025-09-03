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

	// Middleware
	hmacMw := middleware.NewHMACMiddleware(cfg.HMACSecret)
	r.Use(hmacMw.Verify)

	idempMw := middleware.NewIdempotencyKeyMiddleware()
	// chi 미들웨어 래퍼 필요시 별도 래핑 필요
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			idempMw.ServeHTTP(w, req, func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		})
	})

	rl := middleware.NewRateLimiter(10, 20) // 값은 예시
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

	// Progression routes
	r.Route("/progression", func(r chi.Router) {
		r.Get("/", handlers.GetProgression)
		r.Post("/", handlers.CreateProgression)
	})

	// Reward routes
	r.Route("/rewards", func(r chi.Router) {
		r.Post("/claim", handlers.ClaimReward)
	})

	port := cfg.Port
	if port == "" {
		port = "8083"
	}

	// Start server
	log.Printf("Starting progression service on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
