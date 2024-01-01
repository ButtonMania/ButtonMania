package db

import (
	"context"

	"buttonmania.win/protocol"
)

type ContextKey string

const (
	// ContextKey represents custom context keys for redis configuration.
	KeyRedisAddress  ContextKey = "redisaddress"
	KeyRedisUsername ContextKey = "redisusername"
	KeyRedisPassword ContextKey = "redispassword"
	KeyRedisDatabase ContextKey = "redisdatabase"
	KeyRedisTLS      ContextKey = "redistls"
)

// DB represents the database client.
type DB struct {
	redis *Redis
}

// NewDB creates a new database instance.
func NewDB(ctx context.Context) (*DB, error) {
	r, rErr := NewRedis(ctx)
	return &DB{
		redis: r,
	}, rErr
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.redis.Close()
}

// AddRecordToLeaderboard adds a gameplay record to the leaderboard.
func (db *DB) AddRecordToLeaderboard(
	roomId protocol.RoomID,
	userID protocol.UserID,
	record protocol.GameplayRecord,
) error {
	return db.redis.AddRecordToLeaderboard(roomId, userID, record)
}

// GetDurationPlaceInLeaderboard retrieves the duration place in the leaderboard.
func (db *DB) GetDurationPlaceInLeaderboard(
	roomId protocol.RoomID,
	duration int64,
) (int64, error) {
	return db.redis.GetDurationPlaceInLeaderboard(roomId, duration)
}

// GetUserPlaceInLeaderboard retrieves the user's place in the leaderboard.
func (db *DB) GetUserPlaceInLeaderboard(
	roomId protocol.RoomID,
	userID protocol.UserID,
) (int64, error) {
	return db.redis.GetUserPlaceInLeaderboard(roomId, userID)
}

// GetUsersCountInLeaderboard retrieves the count of users in the leaderboard.
func (db *DB) GetUsersCountInLeaderboard(
	roomId protocol.RoomID,
) (int64, error) {
	return db.redis.GetUsersCountInLeaderboard(roomId)
}

// GetUserPlaceInActiveSessions retrieves the user's place in active sessions.
func (db *DB) GetUserPlaceInActiveSessions(
	roomId protocol.RoomID,
	userID protocol.UserID,
) (int64, error) {
	return db.redis.GetUserPlaceInActiveSessions(roomId, userID)
}

// GetUsersCountInActiveSessions retrieves the count of users in active sessions.
func (db *DB) GetUsersCountInActiveSessions(
	roomId protocol.RoomID,
) (int64, error) {
	return db.redis.GetUsersCountInActiveSessions(roomId)
}

// GetBestDurationInLeaderboard retrieves the best duration achieved by a player in the leaderboard.
func (db *DB) GetBestDurationInLeaderboard(
	roomId protocol.RoomID,
) (int64, error) {
	return db.redis.GetBestDurationInLeaderboard(roomId)
}

// SetUserDurationToActiveSessions sets the user's duration in active sessions.
func (db *DB) SetUserDurationToActiveSessions(
	roomId protocol.RoomID,
	userID protocol.UserID,
	duration int64,
) error {
	return db.redis.SetUserDurationToActiveSessions(roomId, userID, duration)
}

// RemoveUserDurationFromActiveSessions removes the user's duration from active sessions.
func (db *DB) RemoveUserDurationFromActiveSessions(
	roomId protocol.RoomID,
	userID protocol.UserID,
) error {
	return db.redis.RemoveUserDurationFromActiveSessions(roomId, userID)
}
