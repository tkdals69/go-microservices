package observability

import (
    "net/http"
    "encoding/json"
)

type HealthResponse struct {
    Status string `json:"status"`
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
    response := HealthResponse{Status: "healthy"}
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}