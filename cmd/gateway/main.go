package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tkdals69/go-microservices/pkg/config"
	"github.com/tkdals69/go-microservices/pkg/handlers"
	"github.com/tkdals69/go-microservices/pkg/observability"
)

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

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting gateway on :%s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
