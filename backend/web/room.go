package web

import (
	"errors"

	"buttonmania.win/chat"
	"buttonmania.win/db"
	"buttonmania.win/localization"
	"buttonmania.win/protocol"
	"github.com/gorilla/websocket"
)

// GameRoom represents a room for managing game sessions.
type GameRoom struct {
	ClientID protocol.ClientID
	RoomID   protocol.RoomID
	MsgLoc   *localization.MessagesLocalization
	DB       *db.DB
	Chat     *chat.Chat
	sessions map[protocol.UserID]*GameSession
	closed   bool
}

// NewGameRoom creates a new GameRoom instance.
func NewGameRoom(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	db *db.DB,
	chat *chat.Chat,
	msgLoc *localization.MessagesLocalization,
) (*GameRoom, error) {
	sessions := make(map[protocol.UserID]*GameSession)
	err := chat.InitConsumerGroup(clientId, roomId)
	return &GameRoom{
		ClientID: clientId,
		RoomID:   roomId,
		MsgLoc:   msgLoc,
		DB:       db,
		Chat:     chat,
		sessions: sessions,
		closed:   false,
	}, err
}

// Stats returns the statistics for the game room.
func (r *GameRoom) Stats(payloadCount int64) (protocol.GameRoomStats, error) {
	countActive, countActiveErr := r.DB.GetUsersCountInActiveSessions(r.ClientID, r.RoomID)
	countLeaderboard, countLeaderboardErr := r.DB.GetUsersCountInLeaderboard(r.ClientID, r.RoomID)
	bestOverallDuration, bestOverallDurationErr := r.DB.GetBestOverallDurationInLeaderboard(r.ClientID, r.RoomID)
	bestTodaysDuration, bestTodaysDurationErr := r.DB.GetTodaysDurationInLeaderboard(r.ClientID, r.RoomID)
	bestUsersPayloads, bestUsersPayloadsErr := r.DB.GetBestUsersPayloads(r.ClientID, r.RoomID, payloadCount)
	err := errors.Join(
		countActiveErr,
		countLeaderboardErr,
		bestOverallDurationErr,
		bestTodaysDurationErr,
		bestUsersPayloadsErr,
	)
	return protocol.NewGameRoomStats(
		&countActive,
		&countLeaderboard,
		&bestOverallDuration,
		&bestTodaysDuration,
		&bestUsersPayloads,
	), err
}

// HasGameSession checks if a game session exists for a user.
func (r *GameRoom) HasGameSession(userID protocol.UserID) bool {
	_, exists := r.sessions[userID]
	return exists
}

// AddGameSession adds a game session to the room.
func (r *GameRoom) AddGameSession(userID protocol.UserID, session *GameSession) {
	r.sessions[userID] = session
}

// RemoveGameSession removes a game session from the room.
func (r *GameRoom) RemoveGameSession(userID protocol.UserID) {
	delete(r.sessions, userID)
}

// MaintainGameSession creates and maintains a game session for a user.
func (r *GameRoom) MaintainGameSession(
	userID protocol.UserID,
	UserPayload protocol.UserPayload,
	UserLocale protocol.UserLocale,
	ws *websocket.Conn,
) error {
	session := NewGameSession(
		userID,
		UserPayload,
		UserLocale,
		r,
		ws,
	)
	return session.MaintainGameSession()
}
