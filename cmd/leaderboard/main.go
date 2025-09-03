package main

import (
    "log"
    "net/http"
    "os"

    "github.com/chi/v5"
    "github.com/tkdals69/go-microservices/pkg/config"
    "github.com/tkdals69/go-microservices/pkg/middleware"
    "github.com/tkdals69/go-microservices/pkg/observability"
    "github.com/tkdals69/go-microservices/pkg/handlers"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("could not load config: %v", err)
    }

    // Initialize router
    r := chi.NewRouter()

    // Middleware
    r.Use(middleware.HMACMiddleware)
    r.Use(middleware.IdempotencyMiddleware)
    r.Use(middleware.RateLimitMiddleware)

    // Health check endpoint
    r.Get("/healthz", observability.HealthCheck)

    // Metrics endpoint
    r.Get("/metrics", observability.Metrics)

    // Leaderboard handlers
    r.Get("/leaderboard", handlers.GetLeaderboard)
    r.Post("/leaderboard", handlers.CreateLeaderboardEntry)

    // Start server
    log.Printf("Starting leaderboard service on port %s", cfg.Port)
    if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
        log.Fatalf("could not start server: %v", err)
    }
}