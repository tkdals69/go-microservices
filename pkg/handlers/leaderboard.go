package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type LeaderboardHandler struct{}

func NewLeaderboardHandler() *LeaderboardHandler {
	return &LeaderboardHandler{}
}

func (h *LeaderboardHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Logic to retrieve leaderboard data
	render.JSON(w, r, http.StatusOK, map[string]interface{}{
		"leaderboard": []string{"Player1", "Player2", "Player3"},
	})
}

func (h *LeaderboardHandler) UpdateLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Logic to update leaderboard data
	render.JSON(w, r, http.StatusOK, map[string]interface{}{
		"message": "Leaderboard updated successfully",
	})
}

func (h *LeaderboardHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetLeaderboard)
	r.Post("/", h.UpdateLeaderboard)
	return r
}