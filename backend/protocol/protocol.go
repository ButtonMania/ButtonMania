package protocol

import (
	"encoding/json"
	"time"
)

type ButtonType string
type ButtonPhase string
type UserLocale string
type UserID string
type MessageType int

const (
	// Button types
	Peace    ButtonType = "peace"
	Love     ButtonType = "love"
	Fortune  ButtonType = "fortune"
	Prestige ButtonType = "prestige"
	// Button phases
	Push    ButtonPhase = "push"
	Hold    ButtonPhase = "hold"
	Release ButtonPhase = "release"
	// User locale
	EN UserLocale = "en"
	RU UserLocale = "ru"
	// Message type
	Update MessageType = 0
	Record MessageType = 1
	Error  MessageType = 2
)

// GameplayMessageType represents the base struct of game, which contains only message type
type GameplayMessageType struct {
	MessageType MessageType `json:"messageType"`
}

// GameplayContext represents the context of a game session.
type GameplayContext struct {
	ButtonPhase ButtonPhase `json:"buttonPhase"`
	Timestamp   *int64      `json:"timestamp,omitempty"`
	Duration    *int64      `json:"duration,omitempty"`
}

// NewGameplayContext creates a new GameplayContext.
func NewGameplayContext() GameplayContext {
	pushTimestamp := time.Now().Unix()
	holdDuration := int64(0)
	return GameplayContext{
		ButtonPhase: Push,
		Timestamp:   &pushTimestamp,
		Duration:    &holdDuration,
	}
}

// GameplayError represents an error that can occur during gameplay.
type GameplayError struct {
	Message string `json:"message"`
}

// NewGameplayError creates a new GameplayError.
func NewGameplayError(msg string) GameplayError {
	return GameplayError{
		Message: msg,
	}
}

// GameplayRecord represents a record of a completed game session.
type GameplayRecord struct {
	Timestamp int64 `json:"timestamp"`
	Duration  int64 `json:"duration"`
}

// NewGameplayRecord creates a new GameplayRecord.
func NewGameplayRecord(ctx GameplayContext) GameplayRecord {
	duration := *ctx.Duration
	timestamp := *ctx.Timestamp
	return GameplayRecord{
		Timestamp: time.Unix(timestamp, 0).Add(time.Duration(duration) * time.Second).Unix(),
		Duration:  duration,
	}
}

// MarshalBinary marshals a GameplayRecord to binary data.
func (r GameplayRecord) MarshalBinary() ([]byte, error) {
	return json.Marshal(r)
}

// GameRoomStats represents statistics for a game room.
type GameRoomStats struct {
	CountActive      *int64 `json:"countActive,omitempty"`
	CountLeaderboard *int64 `json:"countLeaderboard,omitempty"`
	BestDuration     *int64 `json:"bestDuration,omitempty"`
}

// NewGameRoomStats creates a new GameRoomStats.
func NewGameRoomStats(
	totalCountActive,
	totalCountLeaderboard,
	bestDuration *int64,
) GameRoomStats {
	return GameRoomStats{
		CountActive:      totalCountActive,
		CountLeaderboard: totalCountLeaderboard,
		BestDuration:     bestDuration,
	}
}

// GameplayMessage represents an update sent to the client during gameplay.
type GameplayMessage struct {
	GameplayMessageType
	GameRoomStats
	Context          *GameplayContext `json:"context,omitempty"`
	Record           *GameplayRecord  `json:"record,omitempty"`
	Error            *GameplayError   `json:"error,omitempty"`
	Message          *string          `json:"message,omitempty"`
	PlaceActive      *int64           `json:"placeActive,omitempty"`
	PlaceLeaderboard *int64           `json:"placeLeaderboard,omitempty"`
	WorldRecord      *bool            `json:"worldRecord,omitempty"`
}

// NewGameplayMessage creates a new GameplayMessage.
func NewGameplayMessage(
	gameplayContext *GameplayContext,
	gameplayRecord *GameplayRecord,
	gameplayError *GameplayError,
	message *string,
	placeActive, totalCountActive, placeLeaderboard, totalCountLeaderboard *int64,
	worldRecord *bool,
	messageType MessageType,
) GameplayMessage {
	return GameplayMessage{
		Context:          gameplayContext,
		Record:           gameplayRecord,
		Error:            gameplayError,
		Message:          message,
		PlaceActive:      placeActive,
		PlaceLeaderboard: placeLeaderboard,
		WorldRecord:      worldRecord,
		GameplayMessageType: GameplayMessageType{
			MessageType: messageType,
		},
		GameRoomStats: GameRoomStats{
			CountActive:      totalCountActive,
			CountLeaderboard: totalCountLeaderboard,
		},
	}
}

// NewUserLocale retrieves the supported user locale string based on the user's input.
func NewUserLocale(locale string) UserLocale {
	if locale == string(RU) {
		return RU
	}
	return EN
}
