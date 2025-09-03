package core

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// ReceiptGenerator handles receipt generation for rewards
type ReceiptGenerator struct {
	hmacGen *HMACGenerator
}

// NewReceiptGenerator creates a new receipt generator
func NewReceiptGenerator(secret string) *ReceiptGenerator {
	return &ReceiptGenerator{
		hmacGen: NewHMACGenerator(secret),
	}
}

// GenerateReceipt creates a new signed reward receipt
func (r *ReceiptGenerator) GenerateReceipt(playerID, item string) *RewardReceipt {
	timestamp := time.Now().Unix()
	receiptID := r.generateReceiptID(playerID, item, timestamp)
	signature := r.hmacGen.SignRewardReceipt(receiptID, playerID, item, timestamp)

	return &RewardReceipt{
		ReceiptID: receiptID,
		PlayerID:  playerID,
		Item:      item,
		Signature: signature,
		Timestamp: timestamp,
	}
}

// VerifyReceipt verifies the authenticity of a reward receipt
func (r *ReceiptGenerator) VerifyReceipt(receipt *RewardReceipt) bool {
	return r.hmacGen.VerifyRewardReceipt(
		receipt.ReceiptID,
		receipt.PlayerID,
		receipt.Item,
		receipt.Timestamp,
		receipt.Signature,
	)
}

// generateReceiptID creates a unique receipt ID
func (r *ReceiptGenerator) generateReceiptID(playerID, item string, timestamp int64) string {
	data := fmt.Sprintf("%s:%s:%d", playerID, item, timestamp)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("receipt_%x", hash[:8]) // Use first 8 bytes for shorter ID
}

// IsReceiptExpired checks if a receipt has expired (24 hours validity)
func (r *ReceiptGenerator) IsReceiptExpired(receipt *RewardReceipt) bool {
	expiryTime := time.Unix(receipt.Timestamp, 0).Add(24 * time.Hour)
	return time.Now().After(expiryTime)
}
