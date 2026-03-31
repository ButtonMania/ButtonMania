package web

import (
	"testing"
	"time"

	"buttonmania.win/protocol"
)

func TestMessageUpdateFrequencies(t *testing.T) {
	if len(MessageUpdateFrequencies) != len(MessageUpdateTimeIntervals) {
		t.Errorf("frequencies (%d) and intervals (%d) must have the same length",
			len(MessageUpdateFrequencies), len(MessageUpdateTimeIntervals))
	}

	// Frequencies should be non-decreasing
	for i := 1; i < len(MessageUpdateFrequencies); i++ {
		if MessageUpdateFrequencies[i] < MessageUpdateFrequencies[i-1] {
			t.Errorf("frequencies should be non-decreasing: index %d (%d) < index %d (%d)",
				i, MessageUpdateFrequencies[i], i-1, MessageUpdateFrequencies[i-1])
		}
	}

	// Time intervals should be strictly increasing
	for i := 1; i < len(MessageUpdateTimeIntervals); i++ {
		if MessageUpdateTimeIntervals[i] <= MessageUpdateTimeIntervals[i-1] {
			t.Errorf("intervals should be increasing: index %d (%d) <= index %d (%d)",
				i, MessageUpdateTimeIntervals[i], i-1, MessageUpdateTimeIntervals[i-1])
		}
	}
}

func TestValidateGameSessionUpdate_ValidTransitions(t *testing.T) {
	session := &GameSession{}

	ts := int64(1000)
	dur := int64(10)

	// Hold -> Release is valid
	current := &protocol.GameplayContext{
		ButtonPhase: protocol.Hold,
		Timestamp:   &ts,
		Duration:    &dur,
	}
	incoming := &protocol.GameplayContext{
		ButtonPhase: protocol.Release,
		Timestamp:   &ts,
		Duration:    &dur,
	}
	if err := session.validateGameSessionUpdate(current, incoming); err != nil {
		t.Errorf("Hold -> Release should be valid: %v", err)
	}
}

func TestValidateGameSessionUpdate_InvalidPushToPush(t *testing.T) {
	session := &GameSession{}

	ts := int64(1000)
	dur := int64(10)

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
		t.Error("Push -> Push should be invalid")
	}
}

func TestValidateGameSessionUpdate_InvalidTimestamp(t *testing.T) {
	session := &GameSession{}

	ts1 := int64(1000)
	ts2 := int64(2000)
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
		t.Error("different timestamps should be invalid")
	}
}

func TestValidateGameSessionUpdate_InvalidDuration(t *testing.T) {
	session := &GameSession{}

	ts := int64(1000)
	durBig := int64(100)
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
		t.Error("decreasing duration should be invalid")
	}
}

func TestShouldSendNewRandomMessage(t *testing.T) {
	dur := int64(0)
	ts := time.Now().Unix()

	ctx := &protocol.GameplayContext{
		Duration:  &dur,
		Timestamp: &ts,
	}

	// Recently sent message - should not send
	session := &GameSession{
		ctx:         ctx,
		lastMsgTime: time.Now().Unix(),
	}
	if session.shouldSendNewRandomMessage() {
		t.Error("should not send message immediately after last one")
	}

	// Old message - should send
	session.lastMsgTime = time.Now().Unix() - 10
	if !session.shouldSendNewRandomMessage() {
		t.Error("should send message after sufficient time gap")
	}
}

func TestNewGameSession(t *testing.T) {
	session := NewGameSession("user1", "payload", protocol.EN, nil, nil)

	if session.userID != "user1" {
		t.Errorf("expected userID 'user1', got %s", session.userID)
	}
	if session.payload != "payload" {
		t.Errorf("expected payload 'payload', got %s", session.payload)
	}
	if session.locale != protocol.EN {
		t.Errorf("expected locale EN, got %s", session.locale)
	}
	if session.ctx != nil {
		t.Error("expected ctx to be nil")
	}
}

func TestSessionErrors(t *testing.T) {
	// Ensure errors are distinct and non-nil
	errors := []error{
		ErrGameSessionAlreadyExists,
		ErrGameSessionFailedToStart,
		ErrFailedToReadGameSessionUpdate,
		ErrGameSessionInvalidUpdate,
		ErrGameSessionInvalidButtonPhase,
		ErrGameSessionInvalidPushTimestamp,
		ErrGameSessionInvalidHoldDuration,
	}
	for i, err := range errors {
		if err == nil {
			t.Errorf("error at index %d should not be nil", i)
		}
	}
}
