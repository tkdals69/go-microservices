package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu        sync.Mutex
	tokens    int
	lastCheck time.Time
	rate      int
	burst     int
}

func NewRateLimiter(rate int, burst int) *RateLimiter {
	return &RateLimiter{
		tokens:    burst,
		lastCheck: time.Now(),
		rate:      rate,
		burst:     burst,
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastCheck).Seconds()
	rl.lastCheck = now

	rl.tokens += int(elapsed * float64(rl.rate))
	if rl.tokens > rl.burst {
		rl.tokens = rl.burst
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}