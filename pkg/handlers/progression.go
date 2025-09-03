package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Event types
type ProgressionEvent struct {
	DeltaXp  int64  `json:"deltaXp"`
	Activity string `json:"activity"`
}

type BossKillEvent struct {
	BossId string `json:"bossId"`
	Tier   int    `json:"tier"`
	Points int64  `json:"points"`
}

type DropClaimedEvent struct {
	DropId string  `json:"dropId"`
	Weight float64 `json:"weight"`
	Source string  `json:"source"`
}

// Player progression data
type PlayerProgression struct {
	PlayerID  string `json:"playerId"`
	Season    int    `json:"season"`
	Xp        int64  `json:"xp"`
	Level     int    `json:"level"`
	UpdatedAt int64  `json:"updatedAt"`
}

// Reward receipt
type RewardReceipt struct {
	ReceiptId string `json:"receiptId"`
	PlayerID  string `json:"playerId"`
	Item      string `json:"item"`
	Signature string `json:"sig"`
	Timestamp int64  `json:"timestamp"`
}

var (
	playerProgressions = make(map[string]*PlayerProgression)
	hmacSecret         = "your-secret-key" // In production, load from config
)

func CreateProgression(w http.ResponseWriter, r *http.Request) {
	var event struct {
		Type     string      `json:"type"`
		PlayerID string      `json:"playerId"`
		Payload  interface{} `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get or create player progression
	progression, exists := playerProgressions[event.PlayerID]
	if !exists {
		progression = &PlayerProgression{
			PlayerID:  event.PlayerID,
			Season:    1,
			Xp:        0,
			Level:     1,
			UpdatedAt: time.Now().Unix(),
		}
		playerProgressions[event.PlayerID] = progression
	}

	// Process different event types
	switch event.Type {
	case "progression":
		if payloadBytes, err := json.Marshal(event.Payload); err == nil {
			var progEvent ProgressionEvent
			if err := json.Unmarshal(payloadBytes, &progEvent); err == nil {
				progression.Xp += progEvent.DeltaXp
				// Calculate level (every 1000 XP = 1 level)
				progression.Level = int(progression.Xp/1000) + 1
				progression.UpdatedAt = time.Now().Unix()
			}
		}
	case "boss_kill":
		if payloadBytes, err := json.Marshal(event.Payload); err == nil {
			var bossEvent BossKillEvent
			if err := json.Unmarshal(payloadBytes, &bossEvent); err == nil {
				progression.Xp += bossEvent.Points
				progression.Level = int(progression.Xp/1000) + 1
				progression.UpdatedAt = time.Now().Unix()
			}
		}
	case "drop_claimed":
		// Handle drop claimed events
		// Could add bonus XP based on drop weight
		if payloadBytes, err := json.Marshal(event.Payload); err == nil {
			var dropEvent DropClaimedEvent
			if err := json.Unmarshal(payloadBytes, &dropEvent); err == nil {
				bonusXp := int64(dropEvent.Weight * 100) // Example: weight * 100 = bonus XP
				progression.Xp += bonusXp
				progression.Level = int(progression.Xp/1000) + 1
				progression.UpdatedAt = time.Now().Unix()
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(progression)
}

func GetProgression(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("playerId")
	if playerID == "" {
		// Return all progressions
		var progressionList []*PlayerProgression
		for _, progression := range playerProgressions {
			progressionList = append(progressionList, progression)
		}
		json.NewEncoder(w).Encode(progressionList)
		return
	}

	progression, exists := playerProgressions[playerID]
	if !exists {
		http.Error(w, "Player progression not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(progression)
}

// ClaimReward generates a HMAC-signed receipt for reward claims
func ClaimReward(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PlayerID string `json:"playerId"`
		Item     string `json:"item"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate receipt
	receiptId := fmt.Sprintf("receipt_%d", time.Now().UnixNano())
	timestamp := time.Now().Unix()

	// Create HMAC signature
	message := fmt.Sprintf("%s:%s:%s:%d", receiptId, request.PlayerID, request.Item, timestamp)
	h := hmac.New(sha256.New, []byte(hmacSecret))
	h.Write([]byte(message))
	signature := hex.EncodeToString(h.Sum(nil))

	receipt := RewardReceipt{
		ReceiptId: receiptId,
		PlayerID:  request.PlayerID,
		Item:      request.Item,
		Signature: signature,
		Timestamp: timestamp,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(receipt)
}
