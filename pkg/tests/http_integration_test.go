package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/tkdals69/go-microservices/pkg/handlers"
)

func TestGatewayEventIngestion(t *testing.T) {
	// Setup router
	r := chi.NewRouter()
	gateway := handlers.NewGatewayHandler()
	gateway.RegisterRoutes(r)

	// Test event data
	eventData := map[string]interface{}{
		"type":     "progression",
		"playerId": "player_123",
		"ts":       1638360000,
		"payload": map[string]interface{}{
			"deltaXp":  150,
			"activity": "quest_completion",
		},
	}

	jsonData, _ := json.Marshal(eventData)

	// Test POST /events
	req := httptest.NewRequest("POST", "/events", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", "test-signature")
	req.Header.Set("Idempotency-Key", "test-key-123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected 202, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Event accepted") {
		t.Errorf("Expected 'Event accepted' in response, got %s", w.Body.String())
	}
}

func TestLeaderboardEndpoints(t *testing.T) {
	// Setup router
	r := chi.NewRouter()
	leaderboard := handlers.NewLeaderboardHandler()
	r.Mount("/leaderboard", leaderboard.Routes())

	// Test GET /leaderboard
	req := httptest.NewRequest("GET", "/leaderboard/?window=weekly&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response["window"] != "weekly" {
		t.Errorf("Expected window=weekly, got %v", response["window"])
	}
}

func TestProgressionEndpoints(t *testing.T) {
	// Test progression event processing
	eventData := map[string]interface{}{
		"type":     "progression",
		"playerId": "player_456",
		"payload": map[string]interface{}{
			"deltaXp":  300,
			"activity": "boss_kill",
		},
	}

	jsonData, _ := json.Marshal(eventData)

	req := httptest.NewRequest("POST", "/progression/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handlers.CreateProgression(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", w.Code)
	}

	// Test getting progression
	req = httptest.NewRequest("GET", "/progression/?playerId=player_456", nil)
	w = httptest.NewRecorder()
	handlers.GetProgression(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestRewardClaim(t *testing.T) {
	claimData := map[string]interface{}{
		"playerId": "player_789",
		"item":     "epic_helmet",
	}

	jsonData, _ := json.Marshal(claimData)

	req := httptest.NewRequest("POST", "/rewards/claim", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handlers.ClaimReward(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var receipt map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &receipt); err != nil {
		t.Errorf("Failed to parse receipt: %v", err)
	}

	if receipt["playerId"] != "player_789" {
		t.Errorf("Expected playerId=player_789, got %v", receipt["playerId"])
	}

	if receipt["item"] != "epic_helmet" {
		t.Errorf("Expected item=epic_helmet, got %v", receipt["item"])
	}

	if receipt["sig"] == nil || receipt["sig"] == "" {
		t.Error("Expected non-empty signature")
	}
}
