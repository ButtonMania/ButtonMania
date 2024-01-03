package db

import (
	"context"
	"errors"

	"buttonmania.win/protocol"
)

type ContextKey string

const (
	// Context keys for db configuration
	KeyPostgresUrl   ContextKey = "postgresurl"
	KeyRedisAddress  ContextKey = "redisaddress"
	KeyRedisUsername ContextKey = "redisusername"
	KeyRedisPassword ContextKey = "redispassword"
	KeyRedisDatabase ContextKey = "redisdatabase"
	KeyRedisTLS      ContextKey = "redistls"
)

// DB represents the database client.
type DB struct {
	redis    *Redis
	postgres *Postgres
}

// NewDB creates a new database instance.
func NewDB(ctx context.Context) (*DB, error) {
	r, rErr := NewRedis(ctx)
	p, pErr := NewPostgres(ctx)
	return &DB{
		redis:    r,
		postgres: p,
	}, errors.Join(rErr, pErr)
}

// Close closes the database connection.
func (db *DB) Close() error {
	rErr := db.redis.close()
	pErr := db.postgres.close()
	return errors.Join(rErr, pErr)
}

// AddRecordToLeaderboard adds a gameplay record to the leaderboard.
func (db *DB) AddRecordToLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
	record protocol.GameplayRecord,
) error {
	return db.postgres.addRecordToLeaderboard(
		clientId,
		roomId,
		userID,
		record,
	)
}

// GetDurationPlaceInLeaderboard retrieves the duration place in the leaderboard.
func (db *DB) GetDurationPlaceInLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	duration int64,
) (int64, error) {
	return db.postgres.getDurationPlaceInLeaderboard(
		clientId,
		roomId,
		duration,
	)
}

// GetUserPlaceInLeaderboard retrieves the user's place in the leaderboard.
func (db *DB) GetUserPlaceInLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) (int64, error) {
	return db.postgres.getUserPlaceInLeaderboard(
		clientId,
		roomId,
		userID,
	)
}

// GetUsersCountInLeaderboard retrieves the count of users in the leaderboard.
func (db *DB) GetUsersCountInLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) (int64, error) {
	return db.postgres.getUsersCountInLeaderboard(
		clientId,
		roomId,
	)
}

// GetBestDurationInLeaderboard retrieves the best duration achieved by a player in the leaderboard.
func (db *DB) GetBestDurationInLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) (int64, error) {
	return db.postgres.getBestDurationInLeaderboard(
		clientId,
		roomId,
	)
}

// GetTodaysRecordInLeaderboard retrieves today's best duration from  the leaderboard.
func (db *DB) GetTodaysRecordInLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) (int64, error) {
	return db.postgres.getTodaysRecordInLeaderboard(
		clientId,
		roomId,
	)
}

// GetUserPlaceInActiveSessions retrieves the user's place in active sessions.
func (db *DB) GetUserPlaceInActiveSessions(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) (int64, error) {
	return db.redis.getUserPlaceInActiveSessions(
		clientId,
		roomId,
		userID,
	)
}

// GetUsersCountInActiveSessions retrieves the count of users in active sessions.
func (db *DB) GetUsersCountInActiveSessions(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) (int64, error) {
	return db.redis.getUsersCountInActiveSessions(
		clientId,
		roomId,
	)
}

// SetUserDurationToActiveSessions sets the user's duration in active sessions.
func (db *DB) SetUserDurationToActiveSessions(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
	duration int64,
	timestamp int64,
) error {
	return db.redis.setUserDurationToActiveSessions(
		clientId,
		roomId,
		userID,
		duration,
		timestamp,
	)
}

// RemoveUserDurationFromActiveSessions removes the user's duration from active sessions.
func (db *DB) RemoveUserDurationFromActiveSessions(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
	timestamp int64,
) error {
	return db.redis.removeUserDurationFromActiveSessions(
		clientId,
		roomId,
		userID,
		timestamp,
	)
}
