package core

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HMACGenerator handles HMAC signature generation and verification
type HMACGenerator struct {
	secret []byte
}

// NewHMACGenerator creates a new HMAC generator with the provided secret
func NewHMACGenerator(secret string) *HMACGenerator {
	return &HMACGenerator{
		secret: []byte(secret),
	}
}

// GenerateSignature generates HMAC signature for a message
func (h *HMACGenerator) GenerateSignature(message string) string {
	mac := hmac.New(sha256.New, h.secret)
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature verifies HMAC signature for a message
func (h *HMACGenerator) VerifySignature(message, signature string) bool {
	expected := h.GenerateSignature(message)
	return hmac.Equal([]byte(signature), []byte(expected))
}

// SignRewardReceipt generates a signed reward receipt
func (h *HMACGenerator) SignRewardReceipt(receiptID, playerID, item string, timestamp int64) string {
	message := fmt.Sprintf("%s:%s:%s:%d", receiptID, playerID, item, timestamp)
	return h.GenerateSignature(message)
}

// VerifyRewardReceipt verifies a reward receipt signature
func (h *HMACGenerator) VerifyRewardReceipt(receiptID, playerID, item string, timestamp int64, signature string) bool {
	message := fmt.Sprintf("%s:%s:%s:%d", receiptID, playerID, item, timestamp)
	return h.VerifySignature(message, signature)
}

// SignEventData generates signature for event data (typically used for webhook validation)
func (h *HMACGenerator) SignEventData(eventType, playerID string, timestamp int64, payloadHash string) string {
	message := fmt.Sprintf("%s:%s:%d:%s", eventType, playerID, timestamp, payloadHash)
	return h.GenerateSignature(message)
}
