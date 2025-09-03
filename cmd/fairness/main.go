package main

import (
    "log"
    "net/http"
    "os"

    "github.com/chi/v5"
    "github.com/tkdals69/go-microservicesyourusername/go-microservices/pkg/config"
    "github.com/tkdals69/go-microservicesyourusername/go-microservices/pkg/middleware"
    "github.com/tkdals69/go-microservicesyourusername/go-microservices/pkg/observability"
    "github.com/tkdals69/go-microservicesyourusername/go-microservices/pkg/handlers"
)

func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("could not load config: %v", err)
    }

    r := chi.NewRouter()

    r.Use(middleware.HMACMiddleware)
    r.Use(middleware.IdempotencyMiddleware)
    r.Use(middleware.RateLimitMiddleware)

    r.Get("/healthz", observability.HealthCheck)
    r.Get("/metrics", observability.Metrics)

    r.Route("/fairness", func(r chi.Router) {
        r.Get("/", handlers.FairnessHandler)
        // Add more routes as needed
    })

    log.Printf("Starting fairness service on port %s", cfg.Port)
    if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
        log.Fatalf("could not start server: %v", err)
    }
}