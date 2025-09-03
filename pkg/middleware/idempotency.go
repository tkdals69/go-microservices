package middleware

import (
	"net/http"
	"sync"
)

type IdempotencyKeyMiddleware struct {
	keys sync.Map
}

func NewIdempotencyKeyMiddleware() *IdempotencyKeyMiddleware {
	return &IdempotencyKeyMiddleware{}
}

func (m *IdempotencyKeyMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	idempotencyKey := r.Header.Get("idempotency-key")
	if idempotencyKey == "" {
		next(w, r)
		return
	}

	if _, loaded := m.keys.LoadOrStore(idempotencyKey, struct{}{}); loaded {
		http.Error(w, "Duplicate request", http.StatusConflict)
		return
	}

	defer m.keys.Delete(idempotencyKey)
	next(w, r)
}