package integration_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHealthzEndpoint(t *testing.T) {
    req, err := http.NewRequest("GET", "/healthz", nil)
    if err != nil {
        t.Fatalf("could not create request: %v", err)
    }

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(HealthzHandler) // Replace with actual handler

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
    handler := http.HandlerFunc(MetricsHandler) // Replace with actual handler

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }
}