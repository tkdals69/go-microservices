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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/tkdals69/go-microservices/pkg/config"
	"github.com/tkdals69/go-microservices/pkg/observability"
)

// Prometheus 메트릭 정의
var (
	droppedEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dropped_events_total",
			Help: "Total number of dropped events due to fairness violations",
		},
		[]string{"player_id", "reason"},
	)

	anomalyFlagsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "anomaly_flags_total",
			Help: "Total number of anomaly flags raised",
		},
		[]string{"player_id", "type"},
	)

	eventsProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_processed_total",
			Help: "Total number of events processed by fairness service",
		},
		[]string{"player_id", "status"},
	)
)

// 플레이어별 이벤트 추적을 위한 구조체
type PlayerActivity struct {
	PlayerID      string
	EventCount    int
	LastEventTime time.Time
	ScoreSum      int
	IsBlocked     bool
	BlockedUntil  time.Time
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
		eventsProcessedTotal.WithLabelValues(event.PlayerID, "invalid").Inc()
		return
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	// 플레이어 활동 추적 가져오기 또는 생성
	activity, exists := fs.playerActivity[event.PlayerID]
	if !exists {
		activity = &PlayerActivity{
			PlayerID:      event.PlayerID,
			LastEventTime: time.Now(),
		}
		fs.playerActivity[event.PlayerID] = activity
	}

	// 차단된 플레이어인지 확인
	if activity.IsBlocked && time.Now().Before(activity.BlockedUntil) {
		droppedEventsTotal.WithLabelValues(event.PlayerID, "blocked").Inc()
		http.Error(w, "Player is temporarily blocked", http.StatusForbidden)
		return
	} else if activity.IsBlocked && time.Now().After(activity.BlockedUntil) {
		// 차단 해제
		activity.IsBlocked = false
		activity.EventCount = 0
		activity.ScoreSum = 0
	}

	currentTime := time.Now()

	// 이벤트 빈도 체크 (초당 이벤트 수)
	if currentTime.Sub(activity.LastEventTime) < time.Second {
		activity.EventCount++
		if activity.EventCount > fs.anomalyThreshold {
			// 이상 행위 감지 - 플레이어 차단
			activity.IsBlocked = true
			activity.BlockedUntil = currentTime.Add(5 * time.Minute) // 5분 차단
			anomalyFlagsTotal.WithLabelValues(event.PlayerID, "rate_limit").Inc()
			droppedEventsTotal.WithLabelValues(event.PlayerID, "rate_limit").Inc()

			log.Printf("Player %s blocked for rate limiting (events: %d)", event.PlayerID, activity.EventCount)
			http.Error(w, "Rate limit exceeded - player blocked", http.StatusTooManyRequests)
			return
		}
	} else {
		// 새로운 초가 시작됨 - 카운트 리셋
		activity.EventCount = 1
		activity.LastEventTime = currentTime
	}

	// 점수 이상 패턴 체크
	if event.Type == "progression" {
		if deltaXp, ok := event.Payload["deltaXp"].(float64); ok {
			activity.ScoreSum += int(deltaXp)

			// 비정상적으로 높은 점수 체크
			if int(deltaXp) > fs.scoreThreshold {
				activity.IsBlocked = true
				activity.BlockedUntil = currentTime.Add(10 * time.Minute) // 10분 차단
				anomalyFlagsTotal.WithLabelValues(event.PlayerID, "score_anomaly").Inc()
				droppedEventsTotal.WithLabelValues(event.PlayerID, "score_anomaly").Inc()

				log.Printf("Player %s blocked for score anomaly (deltaXp: %d)", event.PlayerID, int(deltaXp))
				http.Error(w, "Score anomaly detected - player blocked", http.StatusForbidden)
				return
			}
		}
	}

	// 정상 이벤트 처리
	eventsProcessedTotal.WithLabelValues(event.PlayerID, "approved").Inc()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "approved",
		"playerId":  event.PlayerID,
		"eventType": event.Type,
		"timestamp": currentTime.Unix(),
	})
}

// 플레이어 상태 조회
func (fs *FairnessService) GetPlayerStatus(w http.ResponseWriter, r *http.Request) {
	playerID := chi.URLParam(r, "playerId")

	fs.mu.RLock()
	activity, exists := fs.playerActivity[playerID]
	fs.mu.RUnlock()

	if !exists {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"playerId":     activity.PlayerID,
		"eventCount":   activity.EventCount,
		"scoreSum":     activity.ScoreSum,
		"isBlocked":    activity.IsBlocked,
		"blockedUntil": activity.BlockedUntil.Unix(),
		"lastEvent":    activity.LastEventTime.Unix(),
	})
}

// 차단된 플레이어 목록 조회
func (fs *FairnessService) GetBlockedPlayers(w http.ResponseWriter, r *http.Request) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var blockedPlayers []map[string]interface{}

	for _, activity := range fs.playerActivity {
		if activity.IsBlocked && time.Now().Before(activity.BlockedUntil) {
			blockedPlayers = append(blockedPlayers, map[string]interface{}{
				"playerId":     activity.PlayerID,
				"blockedUntil": activity.BlockedUntil.Unix(),
				"eventCount":   activity.EventCount,
				"scoreSum":     activity.ScoreSum,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"blockedPlayers": blockedPlayers,
		"count":          len(blockedPlayers),
	})
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
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// 미들웨어
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 라우트
	r.Get("/healthz", observability.HealthCheck)
	r.Get("/metrics", observability.MetricsHandler().ServeHTTP)

	// Fairness 관련 엔드포인트
	r.Post("/check", fairnessService.CheckEvent)
	r.Get("/player/{playerId}", fairnessService.GetPlayerStatus)
	r.Get("/blocked", fairnessService.GetBlockedPlayers)

	port := cfg.Port
	if port == "" {
		port = "8081"
	}

	log.Printf("Starting fairness service on :%s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
