package core

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// ErrInvalidPlayerID indicates an invalid player ID format
	ErrInvalidPlayerID = errors.New("invalid player ID format")
	// ErrInvalidEventType indicates an unsupported event type
	ErrInvalidEventType = errors.New("invalid event type")
	// ErrMissingPayload indicates missing event payload
	ErrMissingPayload = errors.New("missing event payload")
)

// ValidEventTypes defines allowed event types
var ValidEventTypes = map[string]bool{
	"progression":  true,
	"boss_kill":    true,
	"drop_claimed": true,
}

// ValidateEvent validates an incoming event
func ValidateEvent(event *Event) error {
	if event.PlayerID == "" {
		return ErrInvalidPlayerID
	}

	// Validate player ID format (alphanumeric, 3-50 characters)
	if !isValidPlayerID(event.PlayerID) {
		return ErrInvalidPlayerID
	}

	// Validate event type
	if !ValidEventTypes[event.Type] {
		return ErrInvalidEventType
	}

	// Validate payload exists
	if event.Payload == nil {
		return ErrMissingPayload
	}

	return nil
}

// isValidPlayerID checks if player ID format is valid
func isValidPlayerID(playerID string) bool {
	if len(playerID) < 3 || len(playerID) > 50 {
		return false
	}

	// Allow alphanumeric and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, playerID)
	return matched
}

// ValidateLeaderboardWindow validates leaderboard window
func ValidateLeaderboardWindow(window string) bool {
	validWindows := map[string]bool{
		"daily":    true,
		"weekly":   true,
		"seasonal": true,
	}

	return validWindows[strings.ToLower(window)]
}

// ValidateProgressionPayload validates progression event payload
func ValidateProgressionPayload(payload interface{}) (*ProgressionEvent, error) {
	// This would typically involve more complex validation
	// For now, assume payload is already structured correctly
	progEvent := &ProgressionEvent{}

	if payloadMap, ok := payload.(map[string]interface{}); ok {
		if deltaXp, exists := payloadMap["deltaXp"]; exists {
			if xp, ok := deltaXp.(float64); ok {
				progEvent.DeltaXp = int64(xp)
			}
		}

		if activity, exists := payloadMap["activity"]; exists {
			if act, ok := activity.(string); ok {
				progEvent.Activity = act
			}
		}
	}

	return progEvent, nil
}

// ValidateBossKillPayload validates boss kill event payload
func ValidateBossKillPayload(payload interface{}) (*BossKillEvent, error) {
	bossEvent := &BossKillEvent{}

	if payloadMap, ok := payload.(map[string]interface{}); ok {
		if bossID, exists := payloadMap["bossId"]; exists {
			if id, ok := bossID.(string); ok {
				bossEvent.BossID = id
			}
		}

		if tier, exists := payloadMap["tier"]; exists {
			if t, ok := tier.(float64); ok {
				bossEvent.Tier = int(t)
			}
		}

		if points, exists := payloadMap["points"]; exists {
			if p, ok := points.(float64); ok {
				bossEvent.Points = int64(p)
			}
		}
	}

	return bossEvent, nil
}
