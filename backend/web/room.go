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
	ButtonType protocol.ButtonType
	MsgLoc     *localization.MessagesLocalization
	DB         *db.DB
	sessions   map[protocol.UserID]*GameSession
}

// NewGameRoom creates a new GameRoom instance.
func NewGameRoom(buttonType protocol.ButtonType, db *db.DB) (*GameRoom, error) {
	sessions := make(map[protocol.UserID]*GameSession)
	msgLoc, err := localization.NewMessagesLocalization(buttonType)
	if err != nil {
		return nil, err
	}
	return &GameRoom{
		ButtonType: buttonType,
		MsgLoc:     msgLoc,
		DB:         db,
		sessions:   sessions,
	}, nil
}

// Stats returns the statistics for the game room.
func (r *GameRoom) Stats() (protocol.GameRoomStats, error) {
	var err error

	countActive, err_ := r.DB.GetUsersCountInActiveSessions(r.ButtonType)
	if err_ != nil {
		err = errors.Join(err, err_)
	}

	countLeaderboard, err_ := r.DB.GetUsersCountInLeaderboard(r.ButtonType)
	if err_ != nil {
		err = errors.Join(err, err_)
	}

	return protocol.NewGameRoomStats(&countActive, &countLeaderboard), err
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
