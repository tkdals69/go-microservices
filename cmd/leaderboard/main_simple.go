package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/tkdals69/go-microservices/pkg/config"
	"github.com/tkdals69/go-microservices/pkg/observability"
)

// Prometheus 메트릭
var (
	leaderboardRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leaderboard_requests_total",
			Help: "Total number of leaderboard requests",
		},
		[]string{"endpoint", "method"},
	)
)

// 리더보드 항목
type LeaderboardEntry struct {
	PlayerID   string    `json:"playerId"`
	PlayerName string    `json:"playerName,omitempty"`
	Score      int64     `json:"score"`
	Rank       int       `json:"rank"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// 점수 업데이트 요청
type ScoreUpdateRequest struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName,omitempty"`
	Score      int64  `json:"score"`
}

// 리더보드 서비스
type LeaderboardService struct {
	mu           sync.RWMutex
	globalBoard  map[string]*LeaderboardEntry
	sortedGlobal []*LeaderboardEntry
	needsSort    bool
}

func NewLeaderboardService() *LeaderboardService {
	return &LeaderboardService{
		globalBoard: make(map[string]*LeaderboardEntry),
		needsSort:   false,
	}
}

// 점수 업데이트
func (ls *LeaderboardService) UpdateScore(w http.ResponseWriter, r *http.Request) {
	leaderboardRequestsTotal.WithLabelValues("update_score", "POST").Inc()

	var req ScoreUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.PlayerID == "" || req.Score < 0 {
		http.Error(w, "Invalid player ID or score", http.StatusBadRequest)
		return
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()

	now := time.Now()
	newEntry := &LeaderboardEntry{
		PlayerID:   req.PlayerID,
		PlayerName: req.PlayerName,
		Score:      req.Score,
		UpdatedAt:  now,
	}

	// 글로벌 리더보드 업데이트 (항상 최고 점수 유지)
	if existing, exists := ls.globalBoard[req.PlayerID]; !exists || req.Score > existing.Score {
		ls.globalBoard[req.PlayerID] = newEntry
		ls.needsSort = true
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"playerId":  req.PlayerID,
		"newScore":  req.Score,
		"timestamp": now,
	})
}

// 리더보드 조회
func (ls *LeaderboardService) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	leaderboardRequestsTotal.WithLabelValues("get_leaderboard", "GET").Inc()

	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.ensureSorted()

	var entries []*LeaderboardEntry
	if len(ls.sortedGlobal) <= limit {
		entries = ls.sortedGlobal
	} else {
		entries = ls.sortedGlobal[:limit]
	}

	response := map[string]interface{}{
		"type":       "global",
		"entries":    entries,
		"totalCount": len(ls.globalBoard),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 플레이어 랭크 조회
func (ls *LeaderboardService) GetPlayerRank(w http.ResponseWriter, r *http.Request) {
	playerID := chi.URLParam(r, "playerId")

	leaderboardRequestsTotal.WithLabelValues("get_player_rank", "GET").Inc()

	ls.mu.RLock()
	defer ls.mu.RUnlock()

	ls.ensureSorted()

	for _, entry := range ls.sortedGlobal {
		if entry.PlayerID == playerID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entry)
			return
		}
	}

	http.Error(w, "Player not found in leaderboard", http.StatusNotFound)
}

// 헬퍼 함수들
func (ls *LeaderboardService) ensureSorted() {
	if ls.needsSort || ls.sortedGlobal == nil {
		ls.sortedGlobal = make([]*LeaderboardEntry, 0, len(ls.globalBoard))
		for _, entry := range ls.globalBoard {
			ls.sortedGlobal = append(ls.sortedGlobal, entry)
		}
		ls.sortEntries()
		ls.needsSort = false
	}
}

func (ls *LeaderboardService) sortEntries() {
	sort.Slice(ls.sortedGlobal, func(i, j int) bool {
		if ls.sortedGlobal[i].Score != ls.sortedGlobal[j].Score {
			return ls.sortedGlobal[i].Score > ls.sortedGlobal[j].Score
		}
		return ls.sortedGlobal[i].UpdatedAt.Before(ls.sortedGlobal[j].UpdatedAt)
	})

	// 순위 설정
	for i, entry := range ls.sortedGlobal {
		entry.Rank = i + 1
	}
}

func main() {
	cfg := config.Load()
	leaderboardService := NewLeaderboardService()

	r := chi.NewRouter()

	// CORS 설정
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// 미들웨어
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 라우트
	r.Get("/healthz", observability.HealthCheck)
	r.Get("/metrics", observability.MetricsHandler().ServeHTTP)

	// Leaderboard 관련 엔드포인트
	r.Post("/score", leaderboardService.UpdateScore)
	r.Get("/board", leaderboardService.GetLeaderboard)
	r.Get("/player/{playerId}/rank", leaderboardService.GetPlayerRank)

	port := cfg.Port
	if port == "" {
		port = "8082"
	}

	log.Printf("Starting leaderboard service on :%s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
