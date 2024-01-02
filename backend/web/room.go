package web

import (
	"errors"

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
	sessions map[protocol.UserID]*GameSession
}

// NewGameRoom creates a new GameRoom instance.
func NewGameRoom(clientId protocol.ClientID, roomId protocol.RoomID, db *db.DB) (*GameRoom, error) {
	sessions := make(map[protocol.UserID]*GameSession)
	msgLoc, err := localization.NewMessagesLocalization(clientId, roomId)
	return &GameRoom{
		ClientID: clientId,
		RoomID:   roomId,
		MsgLoc:   msgLoc,
		DB:       db,
		sessions: sessions,
	}, err
}

// Stats returns the statistics for the game room.
func (r *GameRoom) Stats() (protocol.GameRoomStats, error) {
	countActive, countActiveErr := r.DB.GetUsersCountInActiveSessions(r.ClientID, r.RoomID)
	countLeaderboard, countLeaderboardErr := r.DB.GetUsersCountInLeaderboard(r.ClientID, r.RoomID)
	bestDuration, bestDurationErr := r.DB.GetBestDurationInLeaderboard(r.ClientID, r.RoomID)
	err := errors.Join(countActiveErr, countLeaderboardErr, bestDurationErr)
	return protocol.NewGameRoomStats(&countActive, &countLeaderboard, &bestDuration), err
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
	UserLocale protocol.UserLocale,
	ws *websocket.Conn,
) error {
	session := NewGameSession(
		userID,
		UserLocale,
		r,
	)
	return session.MaintainGameSession(ws)
}
