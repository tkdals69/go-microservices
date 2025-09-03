package middleware

import (
    "crypto/hmac"
    "crypto/sha256"
    "net/http"
)

type HMACMiddleware struct {
    secret []byte
}

func NewHMACMiddleware(secret string) *HMACMiddleware {
    return &HMACMiddleware{
        secret: []byte(secret),
    }
}

func (h *HMACMiddleware) Verify(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        signature := r.Header.Get("X-Signature")
        if signature == "" {
            http.Error(w, "Missing signature", http.StatusUnauthorized)
            return
        }

        // Create HMAC hash of the request body
        hmacHash := hmac.New(sha256.New, h.secret)
        if _, err := hmacHash.Write([]byte(r.URL.String())); err != nil {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        expectedSignature := hmacHash.Sum(nil)

        // Compare the expected signature with the provided signature
        if !hmac.Equal([]byte(signature), expectedSignature) {
            http.Error(w, "Invalid signature", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}