//go:build integration

package db

import (
	"errors"
	"testing"
	"time"

	"buttonmania.win/protocol"
)

const (
	testClientId = protocol.ClientID("testclient")
	testRoomId   = protocol.RoomID("testroom")
)

// newTestDB connects to the databases and flushes leftover state.
func newTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := NewDB(IntegrationContext())
	if err != nil {
		t.Fatalf("failed to connect to test databases: %v", err)
	}
	if err := db.FlushForTests(); err != nil {
		t.Fatalf("failed to flush test databases: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCustomRoomLifecycle(t *testing.T) {
	db := newTestDB(t)
	owner := protocol.UserID("owner")

	if err := db.AddCustomGameRoom(testClientId, testRoomId, owner); err != nil {
		t.Fatalf("failed to create room: %v", err)
	}
	if err := db.AddCustomGameRoom(testClientId, testRoomId, owner); !errors.Is(err, ErrRoomExists) {
		t.Fatalf("expected ErrRoomExists, got %v", err)
	}

	rooms, err := db.ListCustomGameRooms()
	if err != nil {
		t.Fatalf("failed to list rooms: %v", err)
	}
	if len(rooms) != 1 || rooms[0].V1 != testClientId || rooms[0].V2 != testRoomId {
		t.Fatalf("unexpected room list: %v", rooms)
	}

	if err := db.RemoveCustomGameRoom(testClientId, testRoomId, owner); err != nil {
		t.Fatalf("owner failed to remove room: %v", err)
	}
	if err := db.RemoveCustomGameRoom(testClientId, testRoomId, owner); !errors.Is(err, ErrRoomNotExist) {
		t.Fatalf("expected ErrRoomNotExist, got %v", err)
	}
}

func TestCustomRoomRemoveRejectsForeignUser(t *testing.T) {
	db := newTestDB(t)
	owner := protocol.UserID("owner")
	stranger := protocol.UserID("stranger")

	if err := db.AddCustomGameRoom(testClientId, testRoomId, owner); err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	err := db.RemoveCustomGameRoom(testClientId, testRoomId, stranger)
	if !errors.Is(err, ErrRoomNotOwned) {
		t.Fatalf("expected ErrRoomNotOwned, got %v", err)
	}

	rooms, err := db.ListCustomGameRooms()
	if err != nil {
		t.Fatalf("failed to list rooms: %v", err)
	}
	if len(rooms) != 1 {
		t.Fatalf("room was removed by a stranger, got %d rooms", len(rooms))
	}
}

func TestActiveSessionsCountAndPlace(t *testing.T) {
	db := newTestDB(t)
	now := time.Now().Unix()

	if err := db.SetUserDurationToActiveSessions(testClientId, testRoomId, "first", 10, now); err != nil {
		t.Fatalf("failed to add first session: %v", err)
	}
	if err := db.SetUserDurationToActiveSessions(testClientId, testRoomId, "second", 20, now); err != nil {
		t.Fatalf("failed to add second session: %v", err)
	}

	count, err := db.GetUsersCountInActiveSessions(testClientId, testRoomId)
	if err != nil {
		t.Fatalf("failed to count active sessions: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 active sessions, got %d", count)
	}

	place, err := db.GetUserPlaceInActiveSessions(testClientId, testRoomId, "second")
	if err != nil {
		t.Fatalf("failed to get place: %v", err)
	}
	if place != 1 {
		t.Fatalf("expected the longest hold to be first, got place %d", place)
	}

	online, err := db.GetOnlineUsersCount(testClientId)
	if err != nil {
		t.Fatalf("failed to count online users: %v", err)
	}
	if online != 2 {
		t.Fatalf("expected 2 online users, got %d", online)
	}

	if err := db.RemoveUserDurationFromActiveSessions(testClientId, testRoomId, "first", now); err != nil {
		t.Fatalf("failed to remove session: %v", err)
	}
	count, err = db.GetUsersCountInActiveSessions(testClientId, testRoomId)
	if err != nil {
		t.Fatalf("failed to count active sessions: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 active session, got %d", count)
	}
}

func TestUserPayloads(t *testing.T) {
	db := newTestDB(t)
	now := time.Now().Unix()
	payload := protocol.UserPayload("payload")

	if err := db.SetUserDurationToActiveSessions(testClientId, testRoomId, "user", 5, now); err != nil {
		t.Fatalf("failed to add session: %v", err)
	}
	if err := db.AddUserPayload(testClientId, testRoomId, "user", payload); err != nil {
		t.Fatalf("failed to add payload: %v", err)
	}

	payloads, err := db.GetBestUsersPayloads(testClientId, testRoomId, 3)
	if err != nil {
		t.Fatalf("failed to get payloads: %v", err)
	}
	if len(payloads) != 1 || payloads[0] != payload {
		t.Fatalf("unexpected payloads: %v", payloads)
	}

	if err := db.RemoveUserPayload(testClientId, testRoomId, "user"); err != nil {
		t.Fatalf("failed to remove payload: %v", err)
	}
	payloads, err = db.GetBestUsersPayloads(testClientId, testRoomId, 3)
	if err != nil {
		t.Fatalf("failed to get payloads: %v", err)
	}
	if len(payloads) != 0 {
		t.Fatalf("expected no payloads, got %v", payloads)
	}
}

func TestChatMessageRoundTrip(t *testing.T) {
	db := newTestDB(t)
	if err := db.InitChatConsumerGroup(testClientId, testRoomId); err != nil {
		t.Fatalf("failed to init consumer group: %v", err)
	}

	sender := protocol.ChatMessage{UserID: "sender", Message: "hello"}
	if err := db.PushChatMessage(testClientId, testRoomId, sender); err != nil {
		t.Fatalf("failed to push message: %v", err)
	}

	got, err := db.PopChatMessage(testClientId, testRoomId, "reader")
	if err != nil {
		t.Fatalf("failed to pop message: %v", err)
	}
	if got.Message != "hello" || got.UserID != "sender" {
		t.Fatalf("unexpected message: %+v", got)
	}
}

// PopChatMessage must not block, it runs on every gameplay update.
func TestChatPopDoesNotBlockWhenEmpty(t *testing.T) {
	db := newTestDB(t)
	if err := db.InitChatConsumerGroup(testClientId, testRoomId); err != nil {
		t.Fatalf("failed to init consumer group: %v", err)
	}

	start := time.Now()
	if _, err := db.PopChatMessage(testClientId, testRoomId, "reader"); err != nil {
		t.Fatalf("failed to pop message: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("pop blocked for %v on an empty stream", elapsed)
	}
}

func TestLeaderboardRecords(t *testing.T) {
	db := newTestDB(t)
	now := time.Now().Unix()

	records := []struct {
		user     protocol.UserID
		duration int64
	}{
		{"first", 30},
		{"second", 20},
		{"third", 10},
	}
	for _, r := range records {
		record := protocol.GameplayRecord{Timestamp: now, Duration: r.duration}
		if err := db.AddRecordToLeaderboard(testClientId, testRoomId, r.user, record); err != nil {
			t.Fatalf("failed to add record: %v", err)
		}
	}

	count, err := db.GetUsersCountInLeaderboard(testClientId, testRoomId)
	if err != nil {
		t.Fatalf("failed to count leaderboard: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 users, got %d", count)
	}

	best, err := db.GetBestOverallDurationInLeaderboard(testClientId, testRoomId)
	if err != nil {
		t.Fatalf("failed to get best duration: %v", err)
	}
	if best != 30 {
		t.Fatalf("expected best duration 30, got %d", best)
	}

	todays, err := db.GetTodaysDurationInLeaderboard(testClientId, testRoomId)
	if err != nil {
		t.Fatalf("failed to get todays duration: %v", err)
	}
	if todays != 30 {
		t.Fatalf("expected todays duration 30, got %d", todays)
	}

	place, err := db.GetDurationPlaceInLeaderboard(testClientId, testRoomId, 30)
	if err != nil {
		t.Fatalf("failed to get duration place: %v", err)
	}
	if place != 0 {
		t.Fatalf("expected the best duration to lead, got place %d", place)
	}

	userPlace, err := db.GetUserPlaceInLeaderboard(testClientId, testRoomId, "third")
	if err != nil {
		t.Fatalf("failed to get user place: %v", err)
	}
	if userPlace != 2 {
		t.Fatalf("expected 2 users ahead, got %d", userPlace)
	}
}

// Leaderboard records must survive durations above int32, duration was SERIAL.
func TestLeaderboardAcceptsLargeDuration(t *testing.T) {
	db := newTestDB(t)
	const largeDuration int64 = 3_000_000_000

	record := protocol.GameplayRecord{Timestamp: time.Now().Unix(), Duration: largeDuration}
	if err := db.AddRecordToLeaderboard(testClientId, testRoomId, "user", record); err != nil {
		t.Fatalf("failed to add large record: %v", err)
	}

	best, err := db.GetBestOverallDurationInLeaderboard(testClientId, testRoomId)
	if err != nil {
		t.Fatalf("failed to get best duration: %v", err)
	}
	if best != largeDuration {
		t.Fatalf("expected %d, got %d", largeDuration, best)
	}
}

// Records of other rooms must not affect the user place.
func TestUserPlaceIsScopedToRoom(t *testing.T) {
	db := newTestDB(t)
	now := time.Now().Unix()
	otherRoom := protocol.RoomID("otherroom")

	own := protocol.GameplayRecord{Timestamp: now, Duration: 10}
	if err := db.AddRecordToLeaderboard(testClientId, testRoomId, "user", own); err != nil {
		t.Fatalf("failed to add record: %v", err)
	}
	foreign := protocol.GameplayRecord{Timestamp: now, Duration: 500}
	if err := db.AddRecordToLeaderboard(testClientId, otherRoom, "stranger", foreign); err != nil {
		t.Fatalf("failed to add foreign record: %v", err)
	}

	place, err := db.GetUserPlaceInLeaderboard(testClientId, testRoomId, "user")
	if err != nil {
		t.Fatalf("failed to get user place: %v", err)
	}
	if place != 0 {
		t.Fatalf("expected no users ahead in this room, got %d", place)
	}
}
