package web

import (
	"errors"
	"sync"

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

	// mu guards sessions and closed, which are touched by every
	// websocket handler goroutine serving this room.
	mu       sync.RWMutex
	sessions map[protocol.UserID]*GameSession
	closed   bool
}

// NewGameRoom creates a new GameRoom instance.
func NewGameRoom(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	db *db.DB,
	msgLoc *localization.MessagesLocalization,
) (*GameRoom, error) {
	sessions := make(map[protocol.UserID]*GameSession)
	err := db.InitChatConsumerGroup(clientId, roomId)
	return &GameRoom{
		ClientID: clientId,
		RoomID:   roomId,
		MsgLoc:   msgLoc,
		DB:       db,
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.sessions[userID]
	return exists
}

// TryAddGameSession registers a session for the user and reports whether it
// was added. It fails when the user already has a session in this room, so
// the check and the insert cannot be interleaved by a competing connection.
func (r *GameRoom) TryAddGameSession(userID protocol.UserID, session *GameSession) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.sessions[userID]; exists {
		return false
	}
	r.sessions[userID] = session
	return true
}

// RemoveGameSession removes the session from the room, but only when it is
// still the one registered for that user. A connection that was rejected as a
// duplicate must not evict the session that is actually running.
func (r *GameRoom) RemoveGameSession(userID protocol.UserID, session *GameSession) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if current, exists := r.sessions[userID]; exists && current == session {
		delete(r.sessions, userID)
	}
}

// IsClosed reports whether the room has been closed.
func (r *GameRoom) IsClosed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.closed
}

// Close marks the room as closed, so running sessions wind down.
func (r *GameRoom) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
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
