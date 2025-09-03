package core

import "time"

// Event represents a general event structure
type Event struct {
	Type      string      `json:"type"`
	PlayerID  string      `json:"playerId"`
	Timestamp int64       `json:"ts"`
	Payload   interface{} `json:"payload"`
}

// ProgressionEvent represents progression-related events
type ProgressionEvent struct {
	DeltaXp  int64  `json:"deltaXp"`
	Activity string `json:"activity"`
}

// BossKillEvent represents boss kill events
type BossKillEvent struct {
	BossID string `json:"bossId"`
	Tier   int    `json:"tier"`
	Points int64  `json:"points"`
}

// DropClaimedEvent represents drop claimed events
type DropClaimedEvent struct {
	DropID string  `json:"dropId"`
	Weight float64 `json:"weight"`
	Source string  `json:"source"`
}

// Player represents player data
type Player struct {
	ID        string    `json:"id"`
	Season    int       `json:"season"`
	XP        int64     `json:"xp"`
	Level     int       `json:"level"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	PlayerID string `json:"playerId"`
	Score    int64  `json:"score"`
	Rank     int    `json:"rank"`
	Window   string `json:"window"`
	Season   int    `json:"season"`
}

// RewardReceipt represents a signed reward receipt
type RewardReceipt struct {
	ReceiptID string `json:"receiptId"`
	PlayerID  string `json:"playerId"`
	Item      string `json:"item"`
	Signature string `json:"sig"`
	Timestamp int64  `json:"timestamp"`
}
