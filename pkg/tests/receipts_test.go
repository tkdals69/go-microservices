package tests

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/tkdals69/go-microservices/pkg/core"
)

func TestReceiptGeneration(t *testing.T) {
	generator := core.NewReceiptGenerator("test-secret-key-32-characters-long")

	// Test receipt generation
	receipt := generator.GenerateReceipt("player_123", "legendary_sword")

	if receipt.PlayerID != "player_123" {
		t.Errorf("Expected player_123, got %s", receipt.PlayerID)
	}

	if receipt.Item != "legendary_sword" {
		t.Errorf("Expected legendary_sword, got %s", receipt.Item)
	}

	if receipt.Signature == "" {
		t.Error("Expected non-empty signature")
	}

	// Test receipt verification
	if !generator.VerifyReceipt(receipt) {
		t.Error("Receipt verification failed")
	}

	// Test tampered receipt
	tamperedReceipt := *receipt
	tamperedReceipt.Item = "common_sword"
	if generator.VerifyReceipt(&tamperedReceipt) {
		t.Error("Tampered receipt should not verify")
	}
}

func TestHMACSignature(t *testing.T) {
	hmacGen := core.NewHMACGenerator("test-secret")

	message := "test-message"
	signature := hmacGen.GenerateSignature(message)

	if signature == "" {
		t.Error("Expected non-empty signature")
	}

	if !hmacGen.VerifySignature(message, signature) {
		t.Error("Signature verification failed")
	}

	// Test with different message
	if hmacGen.VerifySignature("different-message", signature) {
		t.Error("Different message should not verify with same signature")
	}
}

func TestEventValidation(t *testing.T) {
	// Valid event
	event := &core.Event{
		Type:      "progression",
		PlayerID:  "player_123",
		Timestamp: 1638360000,
		Payload: map[string]interface{}{
			"deltaXp":  150,
			"activity": "quest_completion",
		},
	}

	if err := core.ValidateEvent(event); err != nil {
		t.Errorf("Valid event should not return error: %v", err)
	}

	// Invalid player ID
	invalidEvent := *event
	invalidEvent.PlayerID = ""
	if err := core.ValidateEvent(&invalidEvent); err == nil {
		t.Error("Empty player ID should return error")
	}

	// Invalid event type
	invalidEvent = *event
	invalidEvent.Type = "invalid_type"
	if err := core.ValidateEvent(&invalidEvent); err == nil {
		t.Error("Invalid event type should return error")
	}
}

// Mock HTTP tests
func TestHTTPEventHandling(t *testing.T) {
	// This would test the HTTP handlers
	// For now, just a placeholder to show structure

	eventData := map[string]interface{}{
		"type":     "progression",
		"playerId": "player_123",
		"ts":       1638360000,
		"payload": map[string]interface{}{
			"deltaXp":  150,
			"activity": "quest_completion",
		},
	}

	jsonData, _ := json.Marshal(eventData)
	req := httptest.NewRequest("POST", "/events", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", "test-signature")
	req.Header.Set("Idempotency-Key", "test-key")

	// This would normally test against actual handlers
	// For now, just verify request setup
	if req.Method != "POST" {
		t.Error("Expected POST method")
	}
}
