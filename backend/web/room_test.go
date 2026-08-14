package web

import (
	"strconv"
	"sync"
	"testing"

	"buttonmania.win/protocol"
	tuple "github.com/barweiss/go-tuple"
)

// newTestRoom builds a room without touching the database, which NewGameRoom
// would need for the chat consumer group.
func newTestRoom() *GameRoom {
	return &GameRoom{
		ClientID: "testclient",
		RoomID:   "testroom",
		sessions: make(map[protocol.UserID]*GameSession),
	}
}

func TestTryAddGameSessionRejectsDuplicate(t *testing.T) {
	room := newTestRoom()
	first := &GameSession{userID: "user"}
	second := &GameSession{userID: "user"}

	if !room.TryAddGameSession("user", first) {
		t.Fatal("first session should be accepted")
	}
	if room.TryAddGameSession("user", second) {
		t.Fatal("second session for the same user should be rejected")
	}
	if !room.HasGameSession("user") {
		t.Fatal("room should still report the running session")
	}
}

// A connection rejected as a duplicate still runs its close path, which must
// not evict the session that is actually playing.
func TestRemoveGameSessionKeepsForeignSession(t *testing.T) {
	room := newTestRoom()
	running := &GameSession{userID: "user"}
	rejected := &GameSession{userID: "user"}

	room.TryAddGameSession("user", running)
	room.TryAddGameSession("user", rejected)

	room.RemoveGameSession("user", rejected)
	if !room.HasGameSession("user") {
		t.Fatal("rejected session evicted the running one")
	}

	room.RemoveGameSession("user", running)
	if room.HasGameSession("user") {
		t.Fatal("running session was not removed by its owner")
	}
}

// Run with -race: concurrent map access used to abort the process with
// "fatal error: concurrent map writes".
func TestGameRoomConcurrentSessionAccess(t *testing.T) {
	room := newTestRoom()
	const goroutines = 32
	const iterations = 100

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			userID := protocol.UserID(strconv.Itoa(id))
			session := &GameSession{userID: userID}
			for j := 0; j < iterations; j++ {
				room.TryAddGameSession(userID, session)
				room.HasGameSession(userID)
				room.IsClosed()
				room.RemoveGameSession(userID, session)
			}
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		room.Close()
	}()

	wg.Wait()

	if !room.IsClosed() {
		t.Fatal("room should be closed")
	}
}

func TestWebRoomRegistry(t *testing.T) {
	web := &Web{rooms: make(map[protocol.RoomKey]*GameRoom)}
	key := protocol.RoomKey(tuple.New2(protocol.ClientID("c"), protocol.RoomID("r")))
	room := newTestRoom()

	if _, exists := web.getRoom(key); exists {
		t.Fatal("empty registry returned a room")
	}
	if !web.addRoom(key, room) {
		t.Fatal("first add should succeed")
	}
	if web.addRoom(key, newTestRoom()) {
		t.Fatal("duplicate add should be rejected")
	}

	got, exists := web.getRoom(key)
	if !exists || got != room {
		t.Fatal("registry returned the wrong room")
	}

	removed, exists := web.removeRoom(key)
	if !exists || removed != room {
		t.Fatal("remove returned the wrong room")
	}
	if _, exists := web.removeRoom(key); exists {
		t.Fatal("second remove should report nothing")
	}
}

// Run with -race: handlers create, read and delete rooms concurrently.
func TestWebRoomRegistryConcurrentAccess(t *testing.T) {
	web := &Web{rooms: make(map[protocol.RoomKey]*GameRoom)}
	const goroutines = 32
	const iterations = 100

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := protocol.RoomKey(tuple.New2(
				protocol.ClientID("c"),
				protocol.RoomID(strconv.Itoa(id)),
			))
			for j := 0; j < iterations; j++ {
				web.addRoom(key, newTestRoom())
				web.getRoom(key)
				web.removeRoom(key)
			}
		}(i)
	}
	wg.Wait()

	if len(web.rooms) != 0 {
		t.Fatalf("expected an empty registry, got %d rooms", len(web.rooms))
	}
}
