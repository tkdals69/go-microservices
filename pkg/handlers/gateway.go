package handlers

import (
    "net/http"
    "github.com/go-chi/chi/v5"
)

type GatewayHandler struct{}

func NewGatewayHandler() *GatewayHandler {
    return &GatewayHandler{}
}

func (h *GatewayHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func (h *GatewayHandler) Metrics(w http.ResponseWriter, r *http.Request) {
    // Implement metrics logic here
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Metrics"))
}

func (h *GatewayHandler) RegisterRoutes(r chi.Router) {
    r.Get("/healthz", h.HealthCheck)
    r.Get("/metrics", h.Metrics)
}