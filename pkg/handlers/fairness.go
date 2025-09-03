package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type FairnessHandler struct {
	db            *sqlx.DB
	cache         *redis.Client
	playerEvents  map[string][]time.Time
	playerScores  map[string]int64
	mutex         sync.RWMutex
	droppedEvents int64
	anomalyFlags  int64
}

func NewFairnessHandler(db *sqlx.DB, cache *redis.Client) *FairnessHandler {
	return &FairnessHandler{
		db:           db,
		cache:        cache,
		playerEvents: make(map[string][]time.Time),
		playerScores: make(map[string]int64),
	}
}

// Event represents an incoming event for fairness checking
type Event struct {
	Type     string      `json:"type"`
	PlayerID string      `json:"playerId"`
	Ts       int64       `json:"ts"`
	Payload  interface{} `json:"payload"`
}

// CheckEventFlood checks if player is sending events too frequently
func (h *FairnessHandler) CheckEventFlood(playerID string) bool {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	now := time.Now()
	events, exists := h.playerEvents[playerID]
	if !exists {
		h.playerEvents[playerID] = []time.Time{now}
		return false
	}

	// Keep only events from last second
	recentEvents := []time.Time{}
	for _, eventTime := range events {
		if now.Sub(eventTime) <= time.Second {
			recentEvents = append(recentEvents, eventTime)
		}
	}
	recentEvents = append(recentEvents, now)
	h.playerEvents[playerID] = recentEvents

	// If more than 10 events per second, it's flooding
	if len(recentEvents) > 10 {
		h.droppedEvents++
		log.Printf("Event flood detected for player %s: %d events in last second", playerID, len(recentEvents))
		return true
	}

	return false
}

// CheckScoreAnomaly detects impossible score increases
func (h *FairnessHandler) CheckScoreAnomaly(playerID string, newScore int64) bool {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	oldScore, exists := h.playerScores[playerID]
	if !exists {
		h.playerScores[playerID] = newScore
		return false
	}

	scoreDiff := newScore - oldScore
	// If score increased by more than 10000 points, it's suspicious
	if scoreDiff > 10000 {
		h.anomalyFlags++
		log.Printf("Score anomaly detected for player %s: score jumped from %d to %d", playerID, oldScore, newScore)
		h.playerScores[playerID] = newScore
		return true
	}

	h.playerScores[playerID] = newScore
	return false
}

func (h *FairnessHandler) ProcessEvent(w http.ResponseWriter, r *http.Request) {
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid event payload", http.StatusBadRequest)
		return
	}

	// Check for event flooding
	if h.CheckEventFlood(event.PlayerID) {
		http.Error(w, "Event flood detected", http.StatusTooManyRequests)
		return
	}

	// Check for score anomalies if it's a score-related event
	if event.Type == "progression" || event.Type == "boss_kill" {
		if payload, ok := event.Payload.(map[string]interface{}); ok {
			if scoreVal, exists := payload["score"]; exists {
				if score, ok := scoreVal.(float64); ok {
					if h.CheckScoreAnomaly(event.PlayerID, int64(score)) {
						// Log but don't block - just flag for review
						log.Printf("Anomalous event flagged for player %s", event.PlayerID)
					}
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Event processed"))
}

func (h *FairnessHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	metrics := map[string]interface{}{
		"dropped_events_total": h.droppedEvents,
		"anomaly_flags_total":  h.anomalyFlags,
		"active_players":       len(h.playerEvents),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *FairnessHandler) GetFairness(w http.ResponseWriter, r *http.Request) {
	h.GetMetrics(w, r)
}

func (h *FairnessHandler) CreateFairness(w http.ResponseWriter, r *http.Request) {
	h.ProcessEvent(w, r)
}

func (h *FairnessHandler) RegisterRoutes(router chi.Router) {
	router.MethodFunc(http.MethodGet, "/fairness", h.GetFairness)
	router.MethodFunc(http.MethodPost, "/fairness", h.CreateFairness)
	router.MethodFunc(http.MethodPost, "/events", h.ProcessEvent)
	router.MethodFunc(http.MethodGet, "/metrics", h.GetMetrics)
}
