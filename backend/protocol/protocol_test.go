package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewUserLocale(t *testing.T) {
	tests := []struct {
		input    string
		expected UserLocale
	}{
		{"ru", RU},
		{"en", EN},
		{"fr", EN},
		{"", EN},
		{"RU", EN}, // case-sensitive
	}
	for _, tt := range tests {
		result := NewUserLocale(tt.input)
		if result != tt.expected {
			t.Errorf("NewUserLocale(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNewGameplayContext(t *testing.T) {
	before := time.Now().Unix()
	ctx := NewGameplayContext()
	after := time.Now().Unix()

	if ctx.ButtonPhase != Push {
		t.Errorf("expected ButtonPhase Push, got %s", ctx.ButtonPhase)
	}
	if ctx.Timestamp == nil {
		t.Fatal("expected Timestamp to be set")
	}
	if *ctx.Timestamp < before || *ctx.Timestamp > after {
		t.Errorf("Timestamp %d out of range [%d, %d]", *ctx.Timestamp, before, after)
	}
	if ctx.Duration == nil || *ctx.Duration != 0 {
		t.Errorf("expected Duration 0, got %v", ctx.Duration)
	}
}

func TestNewGameplayRecord(t *testing.T) {
	ts := int64(1000)
	dur := int64(60)
	ctx := GameplayContext{
		ButtonPhase: Release,
		Timestamp:   &ts,
		Duration:    &dur,
	}

	record := NewGameplayRecord(ctx)
	expectedTimestamp := time.Unix(ts, 0).Add(time.Duration(dur) * time.Second).Unix()

	if record.Timestamp != expectedTimestamp {
		t.Errorf("expected Timestamp %d, got %d", expectedTimestamp, record.Timestamp)
	}
	if record.Duration != dur {
		t.Errorf("expected Duration %d, got %d", dur, record.Duration)
	}
}

func TestGameplayRecordMarshalBinary(t *testing.T) {
	record := GameplayRecord{Timestamp: 1000, Duration: 60}
	data, err := record.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}

	var unmarshalled GameplayRecord
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if unmarshalled != record {
		t.Errorf("expected %+v, got %+v", record, unmarshalled)
	}
}

func TestNewGameplayError(t *testing.T) {
	msg := GameMessage("test error")
	err := NewGameplayError(msg)
	if err.Message != msg {
		t.Errorf("expected %q, got %q", msg, err.Message)
	}
}

func TestNewClientStats(t *testing.T) {
	online := int64(10)
	rooms := int64(5)
	stats := NewClientStats(&online, &rooms)

	if *stats.UsersOnline != 10 {
		t.Errorf("expected UsersOnline 10, got %d", *stats.UsersOnline)
	}
	if *stats.RoomsCount != 5 {
		t.Errorf("expected RoomsCount 5, got %d", *stats.RoomsCount)
	}
}

func TestNewGameRoomStats(t *testing.T) {
	active := int64(3)
	leaderboard := int64(100)
	bestOverall := int64(3600)
	bestToday := int64(1800)
	payloads := []UserPayload{"p1", "p2"}

	stats := NewGameRoomStats(&active, &leaderboard, &bestOverall, &bestToday, &payloads)

	if *stats.CountActive != 3 {
		t.Errorf("expected CountActive 3, got %d", *stats.CountActive)
	}
	if *stats.CountLeaderboard != 100 {
		t.Errorf("expected CountLeaderboard 100, got %d", *stats.CountLeaderboard)
	}
	if *stats.BestOverallDuration != 3600 {
		t.Errorf("expected BestOverallDuration 3600, got %d", *stats.BestOverallDuration)
	}
}

func TestNewGameplayMessage(t *testing.T) {
	ctx := NewGameplayContext()
	place := int64(1)
	count := int64(10)

	msg := NewGameplayMessage(&ctx, nil, nil, nil, nil, &place, &count, nil, nil, nil, Update)

	if msg.GameState != Update {
		t.Errorf("expected GameState Update, got %d", msg.GameState)
	}
	if msg.Context == nil {
		t.Error("expected Context to be set")
	}
	if *msg.PlaceActive != 1 {
		t.Errorf("expected PlaceActive 1, got %d", *msg.PlaceActive)
	}
	if *msg.CountActive != 10 {
		t.Errorf("expected CountActive 10, got %d", *msg.CountActive)
	}
}

func TestChatMessageMarshalBinary(t *testing.T) {
	msg := ChatMessage{UserID: "user1", Message: "hello"}
	data, err := msg.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}

	var unmarshalled ChatMessage
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if unmarshalled.UserID != msg.UserID || unmarshalled.Message != msg.Message {
		t.Errorf("expected %+v, got %+v", msg, unmarshalled)
	}
}

func TestGameplayMessageJSON(t *testing.T) {
	ctx := NewGameplayContext()
	errMsg := NewGameplayError("error")
	msg := NewGameplayMessage(nil, nil, &errMsg, nil, nil, nil, nil, nil, nil, nil, Error)

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshalled GameplayMessage
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshalled.GameState != Error {
		t.Errorf("expected GameState Error, got %d", unmarshalled.GameState)
	}

	_ = ctx // suppress unused
}

func TestButtonPhaseConstants(t *testing.T) {
	if Push != "push" {
		t.Error("Push should be 'push'")
	}
	if Hold != "hold" {
		t.Error("Hold should be 'hold'")
	}
	if Release != "release" {
		t.Error("Release should be 'release'")
	}
}

func TestGameStateConstants(t *testing.T) {
	if Update != 0 {
		t.Error("Update should be 0")
	}
	if Record != 1 {
		t.Error("Record should be 1")
	}
	if Error != 99 {
		t.Error("Error should be 99")
	}
}
