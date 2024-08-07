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

// GetBestOverallDurationInLeaderboard retrieves the best duration achieved by a player in the leaderboard.
func (db *DB) GetBestOverallDurationInLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) (int64, error) {
	return db.postgres.getBestOverallDurationInLeaderboard(
		clientId,
		roomId,
	)
}

// GetTodaysDurationInLeaderboard retrieves today's best duration from  the leaderboard.
func (db *DB) GetTodaysDurationInLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) (int64, error) {
	return db.postgres.getTodaysDurationInLeaderboard(
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

// ListCustomGameRooms returs identifiers of custom game rooms
func (db *DB) ListCustomGameRooms() ([]protocol.RoomKey, error) {
	return db.redis.listCustomGameRooms()
}

// AddCustomGameRoom add new custom game room.
func (db *DB) AddCustomGameRoom(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) error {
	return db.redis.createCustomRoom(
		clientId,
		roomId,
		userID,
	)
}

// AddCustomGameRoom add new custom game room.
func (db *DB) RemoveCustomGameRoom(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) error {
	return db.redis.removeCustomRoom(
		clientId,
		roomId,
		userID,
	)
}

// GetBestUsersPayloads get payloads from redis
func (db *DB) GetBestUsersPayloads(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	count int64,
) ([]protocol.UserPayload, error) {
	return db.redis.getBestUsersPayloads(
		clientId,
		roomId,
		count,
	)
}

// AddUserPayload add payload to redis
func (db *DB) AddUserPayload(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
	payload protocol.UserPayload,
) error {
	return db.redis.addUserPayload(
		clientId,
		roomId,
		userID,
		payload,
	)
}

// RemoveUserPayload remove payload from redis
func (db *DB) RemoveUserPayload(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) error {
	return db.redis.removeUserPayload(
		clientId,
		roomId,
		userID,
	)
}

// InitChatConsumerGroup initialize consumer group for chat stream
func (db *DB) InitChatConsumerGroup(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) error {
	return db.redis.initChatConsumerGroup(
		clientId,
		roomId,
	)
}

// AddConsumerToGroup add new consumer to group
func (db *DB) AddConsumerToGroup(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userId protocol.UserID,
) error {
	return db.redis.addChatConsumerToGroup(
		clientId,
		roomId,
		userId,
	)
}

// PushChatMessage push user's chat message.
func (db *DB) PushChatMessage(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	chatMessage protocol.ChatMessage,
) error {
	return db.redis.pushChatMessage(
		clientId,
		roomId,
		chatMessage,
	)
}

// PopChatMessage pop user's chat message
func (db *DB) PopChatMessage(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) (protocol.ChatMessage, error) {
	return db.redis.popChatMessage(
		clientId,
		roomId,
		userID,
	)
}
