package db

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"buttonmania.win/protocol"
	"github.com/go-redis/redis/v8"
)

type ContextKey string
type RedisKey string

const trimRandChance = 5
const maxRecordsCount = 10

const (
	// ContextKey represents custom context keys for redis configuration.
	KeyRedisAddress  ContextKey = "redisaddress"
	KeyRedisUsername ContextKey = "redisusername"
	KeyRedisPassword ContextKey = "redispassword"
	KeyRedisDatabase ContextKey = "redisdatabase"
	KeyRedisTLS      ContextKey = "redistls"

	// RedisKey represents Redis custom keys.
	RedisKeyActiveSessions RedisKey = "sessions"
	RedisKeyLeaderboard    RedisKey = "leaderboard"
	RedisKeyRecords        RedisKey = "records"
	RedisKeyPredefined     RedisKey = "predefined"
)

// DB represents the database client.
type DB struct {
	ctx    context.Context
	client *redis.Client
}

// NewDB creates a new database instance.
func NewDB(ctx context.Context) (*DB, error) {
	redisaddress, _ := ctx.Value(KeyRedisAddress).(string)
	redisusername, _ := ctx.Value(KeyRedisUsername).(string)
	redispassword, _ := ctx.Value(KeyRedisPassword).(string)
	redisdatabase, _ := ctx.Value(KeyRedisDatabase).(int)
	redistls, _ := ctx.Value(KeyRedisTLS).(bool)

	var tlsConfig *tls.Config
	if redistls {
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:      redisaddress,
		Username:  redisusername,
		Password:  redispassword,
		DB:        redisdatabase,
		TLSConfig: tlsConfig,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &DB{
		ctx:    ctx,
		client: client,
	}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.client.Close()
}

// AddRecordToLeaderboard adds a gameplay record to the leaderboard.
func (db *DB) AddRecordToLeaderboard(
	buttonType protocol.ButtonType,
	userID protocol.UserID,
	record protocol.GameplayRecord,
) error {
	userIDStr := string(userID)
	recordKey := fmt.Sprintf(
		"%s:%s:%s:%s",
		RedisKeyRecords,
		RedisKeyPredefined,
		buttonType,
		userIDStr,
	)
	leaderboardKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyLeaderboard,
		RedisKeyPredefined,
		buttonType,
	)

	// Retrieve current record
	zScoreVal := db.client.ZScore(db.ctx, leaderboardKey, userIDStr).Val()

	// Write only better results
	floatDuration := float64(record.Duration)
	if floatDuration > zScoreVal {
		if err := db.client.ZAdd(db.ctx, leaderboardKey, &redis.Z{
			Score:  floatDuration,
			Member: userIDStr,
		}).Err(); err != nil {
			return err
		}
	}

	// Add record to list
	if err := db.client.LPush(db.ctx, recordKey, record).Err(); err != nil {
		return err
	}

	// Trim list
	if rand.Intn(trimRandChance) == 0 {
		if err := db.client.LTrim(db.ctx, recordKey, 0, maxRecordsCount).Err(); err != nil {
			return err
		}
	}

	return nil
}

// GetDurationPlaceInLeaderboard retrieves the duration place in the leaderboard.
func (db *DB) GetDurationPlaceInLeaderboard(
	buttonType protocol.ButtonType,
	duration int64,
) (int64, error) {
	// Define the leaderboard key for the specific buttonType.
	leaderboardKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyLeaderboard,
		RedisKeyPredefined,
		buttonType,
	)

	// Get the range of members with scores greater than or equal to the given duration.
	rng, zRangeErr := db.client.ZRangeByScore(db.ctx, leaderboardKey, &redis.ZRangeBy{
		Min:    strconv.FormatInt(duration, 10),
		Max:    "+inf",
		Offset: 0,
		Count:  1,
	}).Result()

	// Get the total count of members in the leaderboard.
	count, zCountErr := db.client.ZCount(db.ctx, leaderboardKey, "-inf", "+inf").Result()

	if len(rng) > 0 {
		// If a member with a score less than or equal to the given duration was found,
		// get its rank in the leaderboard.
		rank, zRankErr := db.client.ZRank(db.ctx, leaderboardKey, rng[0]).Result()

		// Calculate the place (count - rank) and return it along with any errors.
		return count - rank, errors.Join(zRangeErr, zCountErr, zRankErr)
	}

	// If no member with a score less than or equal to the given duration was found,
	// the place is one greater than the total count (since ranks are 0-based).
	return count + 1, errors.Join(zRangeErr, zCountErr)
}

// GetUserPlaceInLeaderboard retrieves the user's place in the leaderboard.
func (db *DB) GetUserPlaceInLeaderboard(
	buttonType protocol.ButtonType,
	userID protocol.UserID,
) (int64, error) {
	leaderboardKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyLeaderboard,
		RedisKeyPredefined,
		buttonType,
	)
	count, zCountErr := db.client.ZCount(db.ctx, leaderboardKey, "-inf", "+inf").Result()
	rank, zRankErr := db.client.ZRank(db.ctx, leaderboardKey, string(userID)).Result()
	return count - rank, errors.Join(zCountErr, zRankErr)
}

// GetUsersCountInLeaderboard retrieves the count of users in the leaderboard.
func (db *DB) GetUsersCountInLeaderboard(buttonType protocol.ButtonType) (int64, error) {
	leaderboardKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyLeaderboard,
		RedisKeyPredefined,
		buttonType,
	)
	return db.client.ZCount(db.ctx, leaderboardKey, "-inf", "+inf").Result()
}

// GetUserPlaceInActiveSessions retrieves the user's place in active sessions.
func (db *DB) GetUserPlaceInActiveSessions(
	buttonType protocol.ButtonType,
	userID protocol.UserID,
) (int64, error) {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyActiveSessions,
		RedisKeyPredefined,
		buttonType,
	)
	count, zCountErr := db.client.ZCount(db.ctx, activeSessionsKey, "-inf", "+inf").Result()
	rank, zRankErr := db.client.ZRank(db.ctx, activeSessionsKey, string(userID)).Result()
	return count - rank, errors.Join(zCountErr, zRankErr)
}

// GetUsersCountInActiveSessions retrieves the count of users in active sessions.
func (db *DB) GetUsersCountInActiveSessions(buttonType protocol.ButtonType) (int64, error) {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyActiveSessions,
		RedisKeyPredefined,
		buttonType,
	)
	return db.client.ZCount(db.ctx, activeSessionsKey, "-inf", "+inf").Result()
}

// SetUserDurationToActiveSessions sets the user's duration in active sessions.
func (db *DB) SetUserDurationToActiveSessions(
	buttonType protocol.ButtonType,
	userID protocol.UserID,
	duration int64,
) error {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyActiveSessions,
		RedisKeyPredefined,
		buttonType,
	)
	return db.client.ZAdd(db.ctx, activeSessionsKey, &redis.Z{
		Score:  float64(duration),
		Member: string(userID),
	}).Err()
}

// RemoveUserDurationFromActiveSessions removes the user's duration from active sessions.
func (db *DB) RemoveUserDurationFromActiveSessions(
	buttonType protocol.ButtonType,
	userID protocol.UserID,
) error {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyActiveSessions,
		RedisKeyPredefined,
		buttonType,
	)
	return db.client.ZRem(db.ctx, activeSessionsKey, string(userID)).Err()
}
