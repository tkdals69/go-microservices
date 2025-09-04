package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/tkdals69/go-microservices/pkg/config"
	"github.com/tkdals69/go-microservices/pkg/observability"
)

// 간단한 인메모리 리더보드 저장소
type PlayerScore struct {
	PlayerID string `json:"playerId"`
	Score    int    `json:"score"`
}

type LeaderboardStore struct {
	mu     sync.RWMutex
	scores map[string]int
}

func NewLeaderboardStore() *LeaderboardStore {
	return &LeaderboardStore{
		scores: make(map[string]int),
	}
}

func (ls *LeaderboardStore) UpdateScore(playerID string, deltaScore int) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.scores[playerID] += deltaScore
}

func (ls *LeaderboardStore) GetTopPlayers(limit int) []PlayerScore {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	var players []PlayerScore
	for playerID, score := range ls.scores {
		players = append(players, PlayerScore{
			PlayerID: playerID,
			Score:    score,
		})
	}

	// 점수별로 내림차순 정렬
	sort.Slice(players, func(i, j int) bool {
		return players[i].Score > players[j].Score
	})

	if len(players) > limit {
		players = players[:limit]
	}

	return players
}

// getClientIP는 다양한 헤더를 확인하여 실제 클라이언트 IP를 추출합니다
func getClientIP(r *http.Request) string {
	// X-Forwarded-For 헤더 확인 (프록시/로드밸런서 뒤에 있는 경우)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// 첫 번째 IP만 사용 (클라이언트의 실제 IP)
		ips := strings.Split(forwarded, ",")
		clientIP := strings.TrimSpace(ips[0])
		if clientIP != "" {
			return clientIP
		}
	}

	// X-Real-IP 헤더 확인
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// RemoteAddr에서 IP 추출 (포트 제거)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func main() {
	// Load configuration
	cfg := config.Load()
	_ = cfg // 사용하지 않으면 경고 방지

	// 리더보드 저장소 초기화
	leaderboardStore := NewLeaderboardStore()

	// Initialize router
	r := chi.NewRouter()

	// CORS 미들웨어 추가
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3001", "http://localhost:3000", "http://localhost:3002", "http://localhost:3003", "http://localhost:8081"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Middleware
	r.Use(middleware.Logger)

	// Routes
	r.Get("/healthz", observability.HealthCheck)
	r.Get("/metrics", observability.MetricsHandler().ServeHTTP)

	// 클라이언트 IP 반환 API
	r.Get("/client-ip", func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"ip": clientIP,
		})
	})

	// 리더보드 프록시 (실제 점수 데이터 반환)
	r.Get("/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		players := leaderboardStore.GetTopPlayers(10)

		// 데이터가 없으면 빈 배열 반환
		if len(players) == 0 {
			players = []PlayerScore{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(players)
	})

	// 커스텀 이벤트 핸들러 (점수 업데이트 포함)
	r.Post("/events", func(w http.ResponseWriter, r *http.Request) {
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

		// 클라이언트 IP를 플레이어 ID로 사용
		clientIP := getClientIP(r)

		// progression 이벤트인 경우 리더보드 업데이트
		if event.Type == "progression" {
			if deltaXP, ok := event.Payload["deltaXp"].(float64); ok {
				leaderboardStore.UpdateScore(clientIP, int(deltaXP))
			}
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Event accepted"))
	})

	// 기존 핸들러는 필요시 다른 라우트에 등록 가능
	// handlers.NewGatewayHandler().RegisterRoutes(r)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting gateway on :%s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
