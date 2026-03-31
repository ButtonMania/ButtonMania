package web

import (
	"testing"

	"buttonmania.win/protocol"
)

func TestGameRoom_SessionManagement(t *testing.T) {
	room := &GameRoom{
		ClientID: "test",
		RoomID:   "testroom",
		sessions: make(map[protocol.UserID]*GameSession),
	}

	userID := protocol.UserID("user1")

	// Initially no session
	if room.HasGameSession(userID) {
		t.Error("expected no session for user1")
	}

	// Add session
	session := &GameSession{userID: userID}
	room.AddGameSession(userID, session)

	if !room.HasGameSession(userID) {
		t.Error("expected session for user1 after adding")
	}

	// Remove session
	room.RemoveGameSession(userID)

	if room.HasGameSession(userID) {
		t.Error("expected no session for user1 after removal")
	}
}

func TestGameRoom_MultipleUsers(t *testing.T) {
	room := &GameRoom{
		ClientID: "test",
		RoomID:   "testroom",
		sessions: make(map[protocol.UserID]*GameSession),
	}

	user1 := protocol.UserID("user1")
	user2 := protocol.UserID("user2")

	room.AddGameSession(user1, &GameSession{userID: user1})
	room.AddGameSession(user2, &GameSession{userID: user2})

	if !room.HasGameSession(user1) || !room.HasGameSession(user2) {
		t.Error("expected both users to have sessions")
	}

	room.RemoveGameSession(user1)

	if room.HasGameSession(user1) {
		t.Error("user1 should be removed")
	}
	if !room.HasGameSession(user2) {
		t.Error("user2 should still exist")
	}
}

func TestGameRoom_ClosedFlag(t *testing.T) {
	room := &GameRoom{
		ClientID: "test",
		RoomID:   "testroom",
		sessions: make(map[protocol.UserID]*GameSession),
		closed:   false,
	}

	if room.closed {
		t.Error("expected room to be open")
	}

	room.closed = true

	if !room.closed {
		t.Error("expected room to be closed")
	}
}
