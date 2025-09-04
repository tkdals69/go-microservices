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

// Prometheus 메트릭 정의
var (
	leaderboardRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leaderboard_requests_total",
			Help: "Total number of leaderboard requests",
		},
		[]string{"endpoint", "method"},
	)

	leaderboardEntriesGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "leaderboard_entries_total",
			Help: "Total number of entries in leaderboard",
		},
		[]string{"leaderboard_type"},
	)

	topPlayerScoreGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "top_player_score",
			Help: "Score of the top player",
		},
		[]string{"leaderboard_type"},
	)
)

// 리더보드 항목
type LeaderboardEntry struct {
	PlayerID   string                 `json:"playerId"`
	PlayerName string                 `json:"playerName,omitempty"`
	Score      int64                  `json:"score"`
	Rank       int                    `json:"rank"`
	Level      int                    `json:"level,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	UpdatedAt  time.Time              `json:"updatedAt"`
}

// 리더보드 응답
type LeaderboardResponse struct {
	Type        string              `json:"type"`
	Entries     []*LeaderboardEntry `json:"entries"`
	TotalCount  int                 `json:"totalCount"`
	PlayerRank  *LeaderboardEntry   `json:"playerRank,omitempty"`
	LastUpdated time.Time           `json:"lastUpdated"`
}

// 점수 업데이트 요청
type ScoreUpdateRequest struct {
	PlayerID   string                 `json:"playerId"`
	PlayerName string                 `json:"playerName,omitempty"`
	Score      int64                  `json:"score"`
	Level      int                    `json:"level,omitempty"`
	GameMode   string                 `json:"gameMode,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// 리더보드 서비스
type LeaderboardService struct {
	mu               sync.RWMutex
	globalBoard      map[string]*LeaderboardEntry // playerID -> entry
	weeklyBoard      map[string]*LeaderboardEntry
	monthlyBoard     map[string]*LeaderboardEntry
	sortedGlobal     []*LeaderboardEntry
	sortedWeekly     []*LeaderboardEntry
	sortedMonthly    []*LeaderboardEntry
	lastWeekReset    time.Time
	lastMonthReset   time.Time
	needsGlobalSort  bool
	needsWeeklySort  bool
	needsMonthlySort bool
}

func NewLeaderboardService() *LeaderboardService {
	now := time.Now()
	return &LeaderboardService{
		globalBoard:    make(map[string]*LeaderboardEntry),
		weeklyBoard:    make(map[string]*LeaderboardEntry),
		monthlyBoard:   make(map[string]*LeaderboardEntry),
		lastWeekReset:  now,
		lastMonthReset: now,
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

	// 주간/월간 리더보드 초기화 체크
	ls.checkPeriodResets()

	now := time.Now()
	newEntry := &LeaderboardEntry{
		PlayerID:   req.PlayerID,
		PlayerName: req.PlayerName,
		Score:      req.Score,
		Level:      req.Level,
		Metadata:   req.Metadata,
		UpdatedAt:  now,
	}

	// 글로벌 리더보드 업데이트 (항상 최고 점수 유지)
	if existing, exists := ls.globalBoard[req.PlayerID]; !exists || req.Score > existing.Score {
		ls.globalBoard[req.PlayerID] = newEntry
		ls.needsGlobalSort = true
	}

	// 주간 리더보드 업데이트
	if existing, exists := ls.weeklyBoard[req.PlayerID]; !exists || req.Score > existing.Score {
		weeklyEntry := *newEntry
		ls.weeklyBoard[req.PlayerID] = &weeklyEntry
		ls.needsWeeklySort = true
	}

	// 월간 리더보드 업데이트
	if existing, exists := ls.monthlyBoard[req.PlayerID]; !exists || req.Score > existing.Score {
		monthlyEntry := *newEntry
		ls.monthlyBoard[req.PlayerID] = &monthlyEntry
		ls.needsMonthlySort = true
	}

	// 메트릭 업데이트
	leaderboardEntriesGauge.WithLabelValues("global").Set(float64(len(ls.globalBoard)))
	leaderboardEntriesGauge.WithLabelValues("weekly").Set(float64(len(ls.weeklyBoard)))
	leaderboardEntriesGauge.WithLabelValues("monthly").Set(float64(len(ls.monthlyBoard)))

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
	leaderboardType := chi.URLParam(r, "type")
	if leaderboardType == "" {
		leaderboardType = "global"
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	playerID := r.URL.Query().Get("playerId")

	leaderboardRequestsTotal.WithLabelValues("get_leaderboard", "GET").Inc()

	ls.mu.RLock()
	defer ls.mu.RUnlock()

	// 주간/월간 리더보드 초기화 체크
	ls.checkPeriodResets()

	var entries []*LeaderboardEntry
	var totalCount int
	var lastUpdated time.Time

	switch leaderboardType {
	case "weekly":
		ls.ensureWeeklySorted()
		entries = ls.getTopEntries(ls.sortedWeekly, limit)
		totalCount = len(ls.weeklyBoard)
		lastUpdated = ls.lastWeekReset
	case "monthly":
		ls.ensureMonthlySorted()
		entries = ls.getTopEntries(ls.sortedMonthly, limit)
		totalCount = len(ls.monthlyBoard)
		lastUpdated = ls.lastMonthReset
	default: // global
		ls.ensureGlobalSorted()
		entries = ls.getTopEntries(ls.sortedGlobal, limit)
		totalCount = len(ls.globalBoard)
		if len(ls.sortedGlobal) > 0 {
			lastUpdated = ls.sortedGlobal[0].UpdatedAt
		}
	}

	response := &LeaderboardResponse{
		Type:        leaderboardType,
		Entries:     entries,
		TotalCount:  totalCount,
		LastUpdated: lastUpdated,
	}

	// 특정 플레이어의 랭크 정보 추가
	if playerID != "" {
		if playerRank := ls.getPlayerRank(leaderboardType, playerID); playerRank != nil {
			response.PlayerRank = playerRank
		}
	}

	// 메트릭 업데이트
	if len(entries) > 0 {
		topPlayerScoreGauge.WithLabelValues(leaderboardType).Set(float64(entries[0].Score))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 플레이어 랭크 조회
func (ls *LeaderboardService) GetPlayerRank(w http.ResponseWriter, r *http.Request) {
	playerID := chi.URLParam(r, "playerId")
	leaderboardType := r.URL.Query().Get("type")
	if leaderboardType == "" {
		leaderboardType = "global"
	}

	leaderboardRequestsTotal.WithLabelValues("get_player_rank", "GET").Inc()

	ls.mu.RLock()
	defer ls.mu.RUnlock()

	playerRank := ls.getPlayerRank(leaderboardType, playerID)
	if playerRank == nil {
		http.Error(w, "Player not found in leaderboard", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playerRank)
}

// 리더보드 통계 조회
func (ls *LeaderboardService) GetLeaderboardStats(w http.ResponseWriter, r *http.Request) {
	leaderboardRequestsTotal.WithLabelValues("get_stats", "GET").Inc()

	ls.mu.RLock()
	defer ls.mu.RUnlock()

	// 각 리더보드별 통계 계산
	stats := map[string]interface{}{
		"global": map[string]interface{}{
			"totalPlayers": len(ls.globalBoard),
			"topScore":     ls.getTopScore(ls.globalBoard),
			"avgScore":     ls.getAverageScore(ls.globalBoard),
		},
		"weekly": map[string]interface{}{
			"totalPlayers": len(ls.weeklyBoard),
			"topScore":     ls.getTopScore(ls.weeklyBoard),
			"avgScore":     ls.getAverageScore(ls.weeklyBoard),
			"resetTime":    ls.lastWeekReset,
		},
		"monthly": map[string]interface{}{
			"totalPlayers": len(ls.monthlyBoard),
			"topScore":     ls.getTopScore(ls.monthlyBoard),
			"avgScore":     ls.getAverageScore(ls.monthlyBoard),
			"resetTime":    ls.lastMonthReset,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// 헬퍼 함수들
func (ls *LeaderboardService) checkPeriodResets() {
	now := time.Now()

	// 주간 리셋 (매주 월요일)
	if now.Sub(ls.lastWeekReset) >= 7*24*time.Hour {
		ls.weeklyBoard = make(map[string]*LeaderboardEntry)
		ls.sortedWeekly = nil
		ls.lastWeekReset = now
		ls.needsWeeklySort = true
	}

	// 월간 리셋 (매월 1일)
	if now.Month() != ls.lastMonthReset.Month() || now.Year() != ls.lastMonthReset.Year() {
		ls.monthlyBoard = make(map[string]*LeaderboardEntry)
		ls.sortedMonthly = nil
		ls.lastMonthReset = now
		ls.needsMonthlySort = true
	}
}

func (ls *LeaderboardService) ensureGlobalSorted() {
	if ls.needsGlobalSort || ls.sortedGlobal == nil {
		ls.sortedGlobal = make([]*LeaderboardEntry, 0, len(ls.globalBoard))
		for _, entry := range ls.globalBoard {
			ls.sortedGlobal = append(ls.sortedGlobal, entry)
		}
		ls.sortEntries(ls.sortedGlobal)
		ls.needsGlobalSort = false
	}
}

func (ls *LeaderboardService) ensureWeeklySorted() {
	if ls.needsWeeklySort || ls.sortedWeekly == nil {
		ls.sortedWeekly = make([]*LeaderboardEntry, 0, len(ls.weeklyBoard))
		for _, entry := range ls.weeklyBoard {
			ls.sortedWeekly = append(ls.sortedWeekly, entry)
		}
		ls.sortEntries(ls.sortedWeekly)
		ls.needsWeeklySort = false
	}
}

func (ls *LeaderboardService) ensureMonthlySorted() {
	if ls.needsMonthlySort || ls.sortedMonthly == nil {
		ls.sortedMonthly = make([]*LeaderboardEntry, 0, len(ls.monthlyBoard))
		for _, entry := range ls.monthlyBoard {
			ls.sortedMonthly = append(ls.sortedMonthly, entry)
		}
		ls.sortEntries(ls.sortedMonthly)
		ls.needsMonthlySort = false
	}
}

func (ls *LeaderboardService) sortEntries(entries []*LeaderboardEntry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Score != entries[j].Score {
			return entries[i].Score > entries[j].Score // 점수 높은 순
		}
		return entries[i].UpdatedAt.Before(entries[j].UpdatedAt) // 같은 점수면 먼저 달성한 순
	})

	// 순위 설정
	for i, entry := range entries {
		entry.Rank = i + 1
	}
}

func (ls *LeaderboardService) getTopEntries(entries []*LeaderboardEntry, limit int) []*LeaderboardEntry {
	if len(entries) <= limit {
		return entries
	}
	return entries[:limit]
}

func (ls *LeaderboardService) getPlayerRank(leaderboardType, playerID string) *LeaderboardEntry {
	var sorted []*LeaderboardEntry

	switch leaderboardType {
	case "weekly":
		ls.ensureWeeklySorted()
		sorted = ls.sortedWeekly
	case "monthly":
		ls.ensureMonthlySorted()
		sorted = ls.sortedMonthly
	default:
		ls.ensureGlobalSorted()
		sorted = ls.sortedGlobal
	}

	for _, entry := range sorted {
		if entry.PlayerID == playerID {
			return entry
		}
	}
	return nil
}

func (ls *LeaderboardService) getTopScore(board map[string]*LeaderboardEntry) int64 {
	var topScore int64
	for _, entry := range board {
		if entry.Score > topScore {
			topScore = entry.Score
		}
	}
	return topScore
}

func (ls *LeaderboardService) getAverageScore(board map[string]*LeaderboardEntry) float64 {
	if len(board) == 0 {
		return 0
	}

	var totalScore int64
	for _, entry := range board {
		totalScore += entry.Score
	}
	return float64(totalScore) / float64(len(board))
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
	r.Get("/board/{type}", leaderboardService.GetLeaderboard)
	r.Get("/board", leaderboardService.GetLeaderboard) // 기본값: global
	r.Get("/player/{playerId}/rank", leaderboardService.GetPlayerRank)
	r.Get("/stats", leaderboardService.GetLeaderboardStats)

	port := cfg.Port
	if port == "" {
		port = "8082"
	}

	log.Printf("Starting leaderboard service on :%s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
