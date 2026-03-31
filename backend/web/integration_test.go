package web

import (
	"testing"

	"buttonmania.win/protocol"
)

// TestGameSessionPhaseTransitions tests the full lifecycle of button phase transitions.
func TestGameSessionPhaseTransitions(t *testing.T) {
	// Test valid phase transitions
	validTransitions := []struct {
		from protocol.ButtonPhase
		to   protocol.ButtonPhase
	}{
		{protocol.Push, protocol.Hold},
		{protocol.Push, protocol.Release},
		{protocol.Hold, protocol.Hold},
		{protocol.Hold, protocol.Release},
	}

	session := &GameSession{}
	ts := int64(1000)
	dur := int64(10)

	for _, tt := range validTransitions {
		current := &protocol.GameplayContext{
			ButtonPhase: tt.from,
			Timestamp:   &ts,
			Duration:    &dur,
		}
		incoming := &protocol.GameplayContext{
			ButtonPhase: tt.to,
			Timestamp:   &ts,
			Duration:    &dur,
		}
		err := session.validateGameSessionUpdate(current, incoming)
		if err != nil {
			t.Errorf("transition %s -> %s should be valid, got error: %v", tt.from, tt.to, err)
		}
	}
}

// TestGameSessionInvalidPhaseTransitions tests invalid button phase transitions.
func TestGameSessionInvalidPhaseTransitions(t *testing.T) {
	session := &GameSession{}
	ts := int64(1000)
	dur := int64(10)

	// Push -> Push is invalid
	current := &protocol.GameplayContext{
		ButtonPhase: protocol.Push,
		Timestamp:   &ts,
		Duration:    &dur,
	}
	incoming := &protocol.GameplayContext{
		ButtonPhase: protocol.Push,
		Timestamp:   &ts,
		Duration:    &dur,
	}
	err := session.validateGameSessionUpdate(current, incoming)
	if err == nil {
		t.Error("Push -> Push transition should be invalid")
	}

	// Hold -> Push is invalid
	current.ButtonPhase = protocol.Hold
	incoming.ButtonPhase = protocol.Push
	err = session.validateGameSessionUpdate(current, incoming)
	if err == nil {
		t.Error("Hold -> Push transition should be invalid")
	}
}

// TestGameSessionTimestampTampering tests that timestamp tampering is detected.
func TestGameSessionTimestampTampering(t *testing.T) {
	session := &GameSession{}
	ts1 := int64(1000)
	ts2 := int64(9999)
	dur := int64(10)

	current := &protocol.GameplayContext{
		ButtonPhase: protocol.Hold,
		Timestamp:   &ts1,
		Duration:    &dur,
	}
	incoming := &protocol.GameplayContext{
		ButtonPhase: protocol.Hold,
		Timestamp:   &ts2,
		Duration:    &dur,
	}
	err := session.validateGameSessionUpdate(current, incoming)
	if err == nil {
		t.Error("different timestamps should be rejected")
	}
}

// TestGameSessionDurationTampering tests that duration tampering is detected.
func TestGameSessionDurationTampering(t *testing.T) {
	session := &GameSession{}
	ts := int64(1000)
	durBig := int64(999)
	durSmall := int64(10)

	current := &protocol.GameplayContext{
		ButtonPhase: protocol.Hold,
		Timestamp:   &ts,
		Duration:    &durBig,
	}
	incoming := &protocol.GameplayContext{
		ButtonPhase: protocol.Hold,
		Timestamp:   &ts,
		Duration:    &durSmall,
	}
	err := session.validateGameSessionUpdate(current, incoming)
	if err == nil {
		t.Error("decreasing duration should be rejected")
	}
}

// TestMessageFrequencyScaling tests the message frequency algorithm.
func TestMessageFrequencyScaling(t *testing.T) {
	// Verify that frequencies scale up as duration increases
	if len(MessageUpdateFrequencies) == 0 {
		t.Fatal("MessageUpdateFrequencies should not be empty")
	}

	// First frequency should be the shortest (most frequent messages at start)
	if MessageUpdateFrequencies[0] > MessageUpdateFrequencies[len(MessageUpdateFrequencies)-1] {
		t.Error("message frequency should increase over time")
	}

	// Time intervals should increase
	for i := 1; i < len(MessageUpdateTimeIntervals); i++ {
		if MessageUpdateTimeIntervals[i] <= MessageUpdateTimeIntervals[i-1] {
			t.Errorf("time intervals should be strictly increasing at index %d", i)
		}
	}
}

// TestMaxChatMessageLength verifies the chat message length constant.
func TestMaxChatMessageLength(t *testing.T) {
	if MaxChatMessageLength <= 0 {
		t.Error("MaxChatMessageLength should be positive")
	}
	if MaxChatMessageLength > 10000 {
		t.Error("MaxChatMessageLength should be reasonable")
	}
}

// TestMaxMessageSize verifies the WebSocket message size limit.
func TestMaxMessageSize(t *testing.T) {
	if MaxMessageSize <= 0 {
		t.Error("MaxMessageSize should be positive")
	}
}
