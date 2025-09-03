package integration_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/tkdals69/go-microservices/pkg/observability"
)

func TestHealthzEndpoint(t *testing.T) {
    req, err := http.NewRequest("GET", "/healthz", nil)
    if err != nil {
        t.Fatalf("could not create request: %v", err)
    }

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(observability.HealthCheck)

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }
}

func TestMetricsEndpoint(t *testing.T) {
    req, err := http.NewRequest("GET", "/metrics", nil)
    if err != nil {
        t.Fatalf("could not create request: %v", err)
    }

    rr := httptest.NewRecorder()
    handler := observability.MetricsHandler()

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }
}