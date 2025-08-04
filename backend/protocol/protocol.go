package protocol

import (
	"encoding/json"
	"time"

	tuple "github.com/barweiss/go-tuple"
)

type ButtonPhase string
type UserLocale string
type UserID string
type UserPayload string
type GameMessage string
type GameState int
type ClientID string
type RoomID string
type RoomKey tuple.T2[ClientID, RoomID]

const (
	// Button phases
	Push    ButtonPhase = "push"
	Hold    ButtonPhase = "hold"
	Release ButtonPhase = "release"
	// User locale
	EN UserLocale = "en"
	RU UserLocale = "ru"
	// Game state
	Update GameState = 0
	Record GameState = 1
	Error  GameState = 99
)

// GameplayGameState represents the base struct of game, which contains only current game state
type GameplayGameState struct {
	GameState GameState `json:"gameState"`
}

// GameplayContext represents the context of a game session.
type GameplayContext struct {
	ButtonPhase ButtonPhase  `json:"buttonPhase"`
	ChatMessage *ChatMessage `json:"chat,omitempty"`
	Timestamp   *int64       `json:"timestamp,omitempty"`
	Duration    *int64       `json:"duration,omitempty"`
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
	Message GameMessage `json:"message"`
}

// NewGameplayError creates a new GameplayError.
func NewGameplayError(msg GameMessage) GameplayError {
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

// ClientStats
type ClientStats struct {
	UsersOnline *int64 `json:"usersOnline,omitempty"`
	RoomsCount  *int64 `json:"roomsCount,omitempty"`
}

// NewClientStats creates a new ClientStats.
func NewClientStats(
	usersOnline *int64,
	roomsCount *int64,
) ClientStats {
	return ClientStats{
		UsersOnline: usersOnline,
		RoomsCount:  roomsCount,
	}
}

// GameRoomStats represents statistics for a game room.
type GameRoomStats struct {
	CountActive         *int64         `json:"countActive,omitempty"`
	CountLeaderboard    *int64         `json:"countLeaderboard,omitempty"`
	BestOverallDuration *int64         `json:"bestOverallDuration,omitempty"`
	BestTodaysDuration  *int64         `json:"bestTodaysDuration,omitempty"`
	BestUsersPayloads   *[]UserPayload `json:"bestUsersPayloads,omitempty"`
}

// NewGameRoomStats creates a new GameRoomStats.
func NewGameRoomStats(
	totalCountActive *int64,
	totalCountLeaderboard *int64,
	bestOverallDuration *int64,
	bestTodaysDuration *int64,
	bestUsersPayloads *[]UserPayload,
) GameRoomStats {
	return GameRoomStats{
		CountActive:         totalCountActive,
		CountLeaderboard:    totalCountLeaderboard,
		BestOverallDuration: bestOverallDuration,
		BestTodaysDuration:  bestTodaysDuration,
		BestUsersPayloads:   bestUsersPayloads,
	}
}

// ChatMessage represents the struct, which contains string message with optional user id
type ChatMessage struct {
	UserID  UserID `json:"userID,omitempty"`
	Message string `json:"message"`
}

// MarshalBinary marshals a ChatMessage to binary data.
func (m ChatMessage) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

// GameplayMessage represents an update sent to the client during gameplay.
type GameplayMessage struct {
	GameplayGameState
	GameRoomStats
	Context          *GameplayContext `json:"context,omitempty"`
	ChatMessage      *ChatMessage     `json:"chat,omitempty"`
	Record           *GameplayRecord  `json:"record,omitempty"`
	Error            *GameplayError   `json:"error,omitempty"`
	GameMessage      *GameMessage     `json:"message,omitempty"`
	PlaceActive      *int64           `json:"placeActive,omitempty"`
	PlaceLeaderboard *int64           `json:"placeLeaderboard,omitempty"`
	WorldRecord      *bool            `json:"worldRecord,omitempty"`
}

// NewGameplayMessage creates a new GameplayMessage.
func NewGameplayMessage(
	gameplayContext *GameplayContext,
	gameplayRecord *GameplayRecord,
	gameplayError *GameplayError,
	chatMessage *ChatMessage,
	gameMessage *GameMessage,
	placeActive *int64,
	totalCountActive *int64,
	placeLeaderboard *int64,
	totalCountLeaderboard *int64,
	worldRecord *bool,
	gameState GameState,
) GameplayMessage {
	return GameplayMessage{
		Context:          gameplayContext,
		Record:           gameplayRecord,
		Error:            gameplayError,
		ChatMessage:      chatMessage,
		GameMessage:      gameMessage,
		PlaceActive:      placeActive,
		PlaceLeaderboard: placeLeaderboard,
		WorldRecord:      worldRecord,
		GameplayGameState: GameplayGameState{
			GameState: gameState,
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
