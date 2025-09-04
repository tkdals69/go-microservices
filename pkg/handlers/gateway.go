package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type GatewayHandler struct{}

func NewGatewayHandler() *GatewayHandler {
	return &GatewayHandler{}
}

// POST /events: {type, playerId, ts, payload} 수집
// HMAC 서명 헤더 X-Signature 검증, idempotency-key로 중복 이벤트 무시, OK면 버스(or 인메모리 채널)로 이벤트 전달
func (h *GatewayHandler) PostEvent(w http.ResponseWriter, r *http.Request) {
	// 1. HMAC 서명 검증 (데모용으로 임시 비활성화)
	signature := r.Header.Get("X-Signature")
	_ = signature // 사용하지 않는다는 표시

	// 2. Idempotency-key 체크 (선택사항)
	idempotencyKey := r.Header.Get("idempotency-key")
	_ = idempotencyKey // 사용하지 않는다는 표시

	// 실제로는 Redis 등 외부 저장소에서 중복 체크 필요 (여기선 메모리 맵 예시)
	// 3. 이벤트 파싱
	type Event struct {
		Type     string      `json:"type"`
		PlayerID string      `json:"playerId"`
		Ts       int64       `json:"ts"`
		Payload  interface{} `json:"payload"`
	}
	var evt Event
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		http.Error(w, "Invalid event payload", http.StatusBadRequest)
		return
	}
	// 4. (생략) 이벤트를 메시지 버스 또는 인메모리 채널로 전달
	// 예시: fmt.Println("Event accepted:", evt)
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Event accepted"))
}

func (h *GatewayHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *GatewayHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	// Implement metrics logic here
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Metrics"))
}

func (h *GatewayHandler) RegisterRoutes(r chi.Router) {
	r.Get("/healthz", h.HealthCheck)
	r.Get("/metrics", h.Metrics)
	r.Post("/events", h.PostEvent)
}
