package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type LeaderboardEntry struct {
	PlayerID string `json:"playerId"`
	Score    int64  `json:"score"`
	Rank     int    `json:"rank"`
}

type LeaderboardData struct {
	Window    string              `json:"window"`
	Season    int                 `json:"season"`
	Entries   []*LeaderboardEntry `json:"entries"`
	UpdatedAt int64               `json:"updatedAt"`
}

type LeaderboardHandler struct {
	leaderboards map[string]*LeaderboardData // key: "season:window"
	mutex        sync.RWMutex
}

func NewLeaderboardHandler() *LeaderboardHandler {
	h := &LeaderboardHandler{
		leaderboards: make(map[string]*LeaderboardData),
	}

	// Initialize default leaderboards
	for _, window := range []string{"daily", "weekly", "seasonal"} {
		key := "1:" + window // season 1
		h.leaderboards[key] = &LeaderboardData{
			Window:    window,
			Season:    1,
			Entries:   []*LeaderboardEntry{},
			UpdatedAt: time.Now().Unix(),
		}
	}

	return h
}

func (h *LeaderboardHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	window := r.URL.Query().Get("window")
	if window == "" {
		window = "weekly"
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	season := 1 // default season
	seasonStr := r.URL.Query().Get("season")
	if seasonStr != "" {
		if s, err := strconv.Atoi(seasonStr); err == nil && s > 0 {
			season = s
		}
	}

	key := strconv.Itoa(season) + ":" + window

	h.mutex.RLock()
	leaderboard, exists := h.leaderboards[key]
	h.mutex.RUnlock()

	if !exists {
		render.JSON(w, r, map[string]interface{}{
			"window":  window,
			"season":  season,
			"entries": []*LeaderboardEntry{},
		})
		return
	}

	// Return top N entries
	entries := leaderboard.Entries
	if len(entries) > limit {
		entries = entries[:limit]
	}

	render.JSON(w, r, map[string]interface{}{
		"window":    window,
		"season":    season,
		"entries":   entries,
		"updatedAt": leaderboard.UpdatedAt,
	})
}

func (h *LeaderboardHandler) UpdateLeaderboard(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PlayerID string `json:"playerId"`
		Score    int64  `json:"score"`
		Window   string `json:"window"`
		Season   int    `json:"season"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.Window == "" {
		request.Window = "weekly"
	}
	if request.Season <= 0 {
		request.Season = 1
	}

	key := strconv.Itoa(request.Season) + ":" + request.Window

	h.mutex.Lock()
	defer h.mutex.Unlock()

	leaderboard, exists := h.leaderboards[key]
	if !exists {
		leaderboard = &LeaderboardData{
			Window:    request.Window,
			Season:    request.Season,
			Entries:   []*LeaderboardEntry{},
			UpdatedAt: time.Now().Unix(),
		}
		h.leaderboards[key] = leaderboard
	}

	// Find existing entry for player
	found := false
	for _, entry := range leaderboard.Entries {
		if entry.PlayerID == request.PlayerID {
			entry.Score = request.Score
			found = true
			break
		}
	}

	// Add new entry if not found
	if !found {
		leaderboard.Entries = append(leaderboard.Entries, &LeaderboardEntry{
			PlayerID: request.PlayerID,
			Score:    request.Score,
		})
	}

	// Sort by score (descending)
	sort.Slice(leaderboard.Entries, func(i, j int) bool {
		return leaderboard.Entries[i].Score > leaderboard.Entries[j].Score
	})

	// Update ranks
	for i, entry := range leaderboard.Entries {
		entry.Rank = i + 1
	}

	leaderboard.UpdatedAt = time.Now().Unix()

	render.JSON(w, r, map[string]interface{}{
		"message":    "Leaderboard updated successfully",
		"playerRank": h.getPlayerRank(leaderboard, request.PlayerID),
	})
}

func (h *LeaderboardHandler) getPlayerRank(leaderboard *LeaderboardData, playerID string) int {
	for _, entry := range leaderboard.Entries {
		if entry.PlayerID == playerID {
			return entry.Rank
		}
	}
	return -1
}

func (h *LeaderboardHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetLeaderboard)
	r.Post("/", h.UpdateLeaderboard)
	return r
}
