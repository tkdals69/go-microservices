package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/tkdals69/go-microservices/pkg/config"
	"github.com/tkdals69/go-microservices/pkg/observability"
)

// 플레이어별 이벤트 추적을 위한 구조체
type PlayerActivity struct {
	PlayerID      string
	EventCount    int
	LastEventTime time.Time
	ScoreSum      int
	IsBlocked     bool
}

// Fairness 서비스 상태 관리
type FairnessService struct {
	mu               sync.RWMutex
	playerActivity   map[string]*PlayerActivity
	anomalyThreshold int // 초당 이벤트 임계값
	scoreThreshold   int // 비정상 점수 임계값
}

func NewFairnessService() *FairnessService {
	return &FairnessService{
		playerActivity:   make(map[string]*PlayerActivity),
		anomalyThreshold: 10,   // 초당 10개 이벤트 이상이면 의심
		scoreThreshold:   1000, // 한 번에 1000점 이상이면 의심
	}
}

// 이벤트 검증 및 공정성 체크
func (fs *FairnessService) CheckEvent(w http.ResponseWriter, r *http.Request) {
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

	fs.mu.Lock()
	defer fs.mu.Unlock()

	// 플레이어 활동 추적
	activity, exists := fs.playerActivity[event.PlayerID]
	if !exists {
		activity = &PlayerActivity{
			PlayerID:      event.PlayerID,
			EventCount:    0,
			LastEventTime: time.Now(),
			ScoreSum:      0,
			IsBlocked:     false,
		}
		fs.playerActivity[event.PlayerID] = activity
	}

	now := time.Now()

	// 차단된 플레이어 체크
	if activity.IsBlocked {
		http.Error(w, "Player is blocked for suspicious activity", http.StatusForbidden)
		return
	}

	// 이벤트 빈도 체크 (1초 내 너무 많은 이벤트)
	if now.Sub(activity.LastEventTime) < time.Second {
		activity.EventCount++
		if activity.EventCount > fs.anomalyThreshold {
			activity.IsBlocked = true
			log.Printf("Player %s blocked for event frequency anomaly: %d events/sec",
				event.PlayerID, activity.EventCount)
			http.Error(w, "Too many events detected", http.StatusTooManyRequests)
			return
		}
	} else {
		// 1초가 지났으면 카운트 리셋
		activity.EventCount = 1
		activity.LastEventTime = now
	}

	// 점수 이상 체크
	if event.Type == "progression" {
		if deltaXP, ok := event.Payload["deltaXp"].(float64); ok {
			if int(deltaXP) > fs.scoreThreshold {
				activity.IsBlocked = true
				log.Printf("Player %s blocked for score anomaly: %d points",
					event.PlayerID, int(deltaXP))
				http.Error(w, "Abnormal score detected", http.StatusForbidden)
				return
			}
			activity.ScoreSum += int(deltaXP)
		}
	}

	// 이벤트가 정상이면 승인
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "approved",
		"playerId":  event.PlayerID,
		"timestamp": now.Unix(),
	})
}

// 블록된 플레이어 조회
func (fs *FairnessService) GetBlockedPlayers(w http.ResponseWriter, r *http.Request) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var blockedPlayers []map[string]interface{}
	for playerID, activity := range fs.playerActivity {
		if activity.IsBlocked {
			blockedPlayers = append(blockedPlayers, map[string]interface{}{
				"playerId":     playerID,
				"eventCount":   activity.EventCount,
				"scoreSum":     activity.ScoreSum,
				"lastActivity": activity.LastEventTime.Unix(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"blockedPlayers": blockedPlayers,
		"totalBlocked":   len(blockedPlayers),
	})
}

// 플레이어 차단 해제
func (fs *FairnessService) UnblockPlayer(w http.ResponseWriter, r *http.Request) {
	playerID := chi.URLParam(r, "playerId")

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if activity, exists := fs.playerActivity[playerID]; exists {
		activity.IsBlocked = false
		activity.EventCount = 0
		log.Printf("Player %s unblocked", playerID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "unblocked",
			"playerId": playerID,
		})
	} else {
		http.Error(w, "Player not found", http.StatusNotFound)
	}
}

func main() {
	cfg := config.Load()
	fairnessService := NewFairnessService()

	r := chi.NewRouter()

	// CORS 설정
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Get("/healthz", observability.HealthCheck)
	r.Get("/metrics", observability.MetricsHandler().ServeHTTP)

	// Fairness API
	r.Post("/check-event", fairnessService.CheckEvent)
	r.Get("/blocked-players", fairnessService.GetBlockedPlayers)
	r.Put("/unblock/{playerId}", fairnessService.UnblockPlayer)

	port := cfg.Port
	if port == "" {
		port = "8081"
	}

	log.Printf("Starting Fairness service on :%s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
