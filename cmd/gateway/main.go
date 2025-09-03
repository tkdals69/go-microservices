package main

import (
    "log"
    "net/http"
        "log"
        "net/http"
        "github.com/go-chi/chi/v5"
        "github.com/go-chi/chi/v5/middleware"
        "github.com/tkdals69/go-microservices/pkg/config"
        "github.com/tkdals69/go-microservices/pkg/handlers"
        "github.com/tkdals69/go-microservices/pkg/observability"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/tkdals69/go-microservices/pkg/config"
    "github.com/tkdals69/go-microservices/pkg/handlers"
func main() {
    // Load configuration
    cfg := config.Load()
    _ = cfg // 사용하지 않으면 경고 방지

    // Initialize router
    r := chi.NewRouter()

    // Middleware
    r.Use(middleware.Logger)

    // Routes
    r.Get("/healthz", observability.HealthCheck)
    r.Get("/metrics", observability.MetricsHandler().ServeHTTP)
    handlers.NewGatewayHandler().RegisterRoutes(r)

    log.Println("Starting gateway on :8080...")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatalf("could not start server: %v", err)
    }
}

    // Start server
    log.Printf("Starting gateway service on %s", cfg.ServerAddress)
    if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
        log.Fatalf("could not start server: %v", err)
    }
}