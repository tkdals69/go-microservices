package main

import (
	"encoding/json"
	"log"
	"net/http"
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
	progressionEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "progression_events_total",
			Help: "Total number of progression events processed",
		},
		[]string{"player_id", "event_type"},
	)

	playerLevelGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "player_level",
			Help: "Current player level",
		},
		[]string{"player_id"},
	)

	playerXpGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "player_xp",
			Help: "Current player XP",
		},
		[]string{"player_id"},
	)
)

// 플레이어 진행도 데이터
type PlayerProgress struct {
	PlayerID      string                 `json:"playerId"`
	Level         int                    `json:"level"`
	XP            int                    `json:"xp"`
	XPToNextLevel int                    `json:"xpToNextLevel"`
	Achievements  []string               `json:"achievements"`
	Stats         map[string]interface{} `json:"stats"`
	LastUpdated   time.Time              `json:"lastUpdated"`
}

// 보상 데이터
type Reward struct {
	Type        string `json:"type"` // "xp", "item", "currency"
	Amount      int    `json:"amount,omitempty"`
	ItemID      string `json:"itemId,omitempty"`
	Description string `json:"description"`
}

// Progression 서비스
type ProgressionService struct {
	mu              sync.RWMutex
	playerProgress  map[string]*PlayerProgress
	levelThresholds []int // 각 레벨에 필요한 XP
}

func NewProgressionService() *ProgressionService {
	// 레벨별 필요 XP 설정 (예: 레벨 1=100XP, 레벨 2=250XP, ...)
	thresholds := make([]int, 100)
	for i := 0; i < 100; i++ {
		thresholds[i] = (i+1)*100 + i*i*10 // 점진적으로 증가하는 XP 요구량
	}

	return &ProgressionService{
		playerProgress:  make(map[string]*PlayerProgress),
		levelThresholds: thresholds,
	}
}

// XP 추가 및 레벨업 처리
func (ps *ProgressionService) ProcessXpGain(w http.ResponseWriter, r *http.Request) {
	var event struct {
		Type     string                 `json:"type"`
		PlayerID string                 `json:"playerId"`
		Ts       int64                  `json:"ts"`
		Payload  map[string]interface{} `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid event payload", http.StatusBadRequest)
		return
	}

	deltaXP, ok := event.Payload["deltaXp"].(float64)
	if !ok {
		http.Error(w, "Missing or invalid deltaXp", http.StatusBadRequest)
		return
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	// 플레이어 진행도 가져오기 또는 생성
	progress, exists := ps.playerProgress[event.PlayerID]
	if !exists {
		progress = &PlayerProgress{
			PlayerID:     event.PlayerID,
			Level:        1,
			XP:           0,
			Achievements: []string{},
			Stats:        make(map[string]interface{}),
			LastUpdated:  time.Now(),
		}
		ps.playerProgress[event.PlayerID] = progress
	}

	// XP 추가
	progress.XP += int(deltaXP)
	progress.LastUpdated = time.Now()

	// 레벨업 확인
	oldLevel := progress.Level
	newLevel := ps.calculateLevel(progress.XP)

	var rewards []Reward
	if newLevel > oldLevel {
		progress.Level = newLevel

		// 레벨업 보상 생성
		for level := oldLevel + 1; level <= newLevel; level++ {
			rewards = append(rewards, Reward{
				Type:        "currency",
				Amount:      level * 10, // 레벨당 10 골드
				Description: "Level up bonus",
			})

			// 특별 레벨 보상
			if level%5 == 0 {
				rewards = append(rewards, Reward{
					Type:        "item",
					ItemID:      "special_chest_" + strconv.Itoa(level/5),
					Description: "Special level milestone reward",
				})
			}
		}

		// 업적 체크
		if newLevel >= 10 && !ps.hasAchievement(progress, "veteran_player") {
			progress.Achievements = append(progress.Achievements, "veteran_player")
			rewards = append(rewards, Reward{
				Type:        "item",
				ItemID:      "veteran_badge",
				Description: "Veteran Player Achievement",
			})
		}
	}

	// 다음 레벨까지 필요한 XP 계산
	progress.XPToNextLevel = ps.getXPToNextLevel(progress.Level, progress.XP)

	// 통계 업데이트
	if stats, ok := progress.Stats["totalXpGained"]; ok {
		progress.Stats["totalXpGained"] = stats.(int) + int(deltaXP)
	} else {
		progress.Stats["totalXpGained"] = int(deltaXP)
	}

	// Prometheus 메트릭 업데이트
	progressionEventsTotal.WithLabelValues(event.PlayerID, "xp_gain").Inc()
	playerLevelGauge.WithLabelValues(event.PlayerID).Set(float64(progress.Level))
	playerXpGauge.WithLabelValues(event.PlayerID).Set(float64(progress.XP))

	// 응답 생성
	response := map[string]interface{}{
		"success":        true,
		"playerProgress": progress,
		"leveledUp":      newLevel > oldLevel,
		"oldLevel":       oldLevel,
		"newLevel":       newLevel,
		"rewards":        rewards,
		"deltaXP":        int(deltaXP),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 플레이어 진행도 조회
func (ps *ProgressionService) GetPlayerProgress(w http.ResponseWriter, r *http.Request) {
	playerID := chi.URLParam(r, "playerId")

	ps.mu.RLock()
	progress, exists := ps.playerProgress[playerID]
	ps.mu.RUnlock()

	if !exists {
		// 새 플레이어 생성
		progress = &PlayerProgress{
			PlayerID:      playerID,
			Level:         1,
			XP:            0,
			XPToNextLevel: ps.levelThresholds[0],
			Achievements:  []string{},
			Stats:         make(map[string]interface{}),
			LastUpdated:   time.Now(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

// 보상 지급
func (ps *ProgressionService) ClaimReward(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PlayerID   string `json:"playerId"`
		RewardType string `json:"rewardType"`
		RewardID   string `json:"rewardId"`
		Amount     int    `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	_, exists := ps.playerProgress[request.PlayerID]
	if !exists {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	// 보상 지급 처리 (실제로는 게임 서버나 인벤토리 서비스와 연동)
	reward := Reward{
		Type:        request.RewardType,
		Amount:      request.Amount,
		ItemID:      request.RewardID,
		Description: "Manual reward claim",
	}

	// HMAC 서명된 영수증 생성 (실제 구현시 사용)
	receipt := map[string]interface{}{
		"playerId":  request.PlayerID,
		"reward":    reward,
		"timestamp": time.Now().Unix(),
		"signature": "hmac_signature_here", // 실제로는 HMAC 생성
	}

	progressionEventsTotal.WithLabelValues(request.PlayerID, "reward_claim").Inc()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"receipt": receipt,
		"message": "Reward claimed successfully",
	})
}

// 레벨 계산 헬퍼 함수
func (ps *ProgressionService) calculateLevel(xp int) int {
	for level, threshold := range ps.levelThresholds {
		if xp < threshold {
			return level + 1 // 레벨은 1부터 시작
		}
		xp -= threshold
	}
	return len(ps.levelThresholds) // 최대 레벨
}

// 다음 레벨까지 필요한 XP 계산
func (ps *ProgressionService) getXPToNextLevel(currentLevel, currentXP int) int {
	if currentLevel >= len(ps.levelThresholds) {
		return 0 // 최대 레벨
	}

	totalXPForCurrentLevel := 0
	for i := 0; i < currentLevel-1; i++ {
		totalXPForCurrentLevel += ps.levelThresholds[i]
	}

	xpNeededForNextLevel := totalXPForCurrentLevel + ps.levelThresholds[currentLevel-1]
	return xpNeededForNextLevel - currentXP
}

// 업적 확인 헬퍼 함수
func (ps *ProgressionService) hasAchievement(progress *PlayerProgress, achievement string) bool {
	for _, a := range progress.Achievements {
		if a == achievement {
			return true
		}
	}
	return false
}

func main() {
	cfg := config.Load()
	progressionService := NewProgressionService()

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

	// Progression 관련 엔드포인트
	r.Post("/xp", progressionService.ProcessXpGain)
	r.Get("/player/{playerId}", progressionService.GetPlayerProgress)
	r.Post("/claim", progressionService.ClaimReward)

	port := cfg.Port
	if port == "" {
		port = "8083"
	}

	log.Printf("Starting progression service on :%s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
