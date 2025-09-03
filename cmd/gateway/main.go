package main

import (
    "log"
    "net/http"
    "os"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/tkdals69/go-microservices/pkg/config"
    "github.com/tkdals69/go-microservices/pkg/handlers"
    "github.com/tkdals69/go-microservices/pkg/observability"
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
    r.Use(middleware.Logger)
    r.Use(observability.HealthCheckMiddleware)
    r.Use(observability.MetricsMiddleware)

    // Routes
    r.Get("/healthz", observability.HealthCheck)
    r.Route("/api", func(r chi.Router) {
        r.Mount("/gateway", handlers.NewGatewayHandler())
    })

    // Start server
    log.Printf("Starting gateway service on %s", cfg.ServerAddress)
    if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
        log.Fatalf("could not start server: %v", err)
    }
}