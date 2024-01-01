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

type RedisKey string

const trimRandChance = 5
const maxRecordsCount = 100

const (
	// RedisKey represents Redis custom keys.
	RedisKeyActiveSessions RedisKey = "sessions"
	RedisKeyLeaderboard    RedisKey = "leaderboard"
	RedisKeyRecords        RedisKey = "records"
	RedisKeyPredefined     RedisKey = "predefined"
)

// Redis represents the redis client.
type Redis struct {
	ctx    context.Context
	client *redis.Client
}

// NewDB creates a new database instance.
func NewRedis(ctx context.Context) (*Redis, error) {
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

	return &Redis{
		ctx:    ctx,
		client: client,
	}, nil
}

// Close closes the database connection.
func (r *Redis) Close() error {
	return r.client.Close()
}

// AddRecordToLeaderboard adds a gameplay record to the leaderboard.
func (r *Redis) AddRecordToLeaderboard(
	roomId protocol.RoomID,
	userID protocol.UserID,
	record protocol.GameplayRecord,
) error {
	userIDStr := string(userID)
	recordKey := fmt.Sprintf(
		"%s:%s:%s:%s",
		RedisKeyRecords,
		RedisKeyPredefined,
		roomId,
		userIDStr,
	)
	leaderboardKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyLeaderboard,
		RedisKeyPredefined,
		roomId,
	)

	// Retrieve current record
	zScoreVal := r.client.ZScore(r.ctx, leaderboardKey, userIDStr).Val()

	// Write only better results
	floatDuration := float64(record.Duration)
	if floatDuration > zScoreVal {
		if err := r.client.ZAdd(r.ctx, leaderboardKey, &redis.Z{
			Score:  floatDuration,
			Member: userIDStr,
		}).Err(); err != nil {
			return err
		}
	}

	// Add record to list
	if err := r.client.LPush(r.ctx, recordKey, record).Err(); err != nil {
		return err
	}

	// Trim list
	if rand.Intn(trimRandChance) == 0 {
		if err := r.client.LTrim(r.ctx, recordKey, 0, maxRecordsCount).Err(); err != nil {
			return err
		}
	}

	return nil
}

// GetDurationPlaceInLeaderboard retrieves the duration place in the leaderboard.
func (r *Redis) GetDurationPlaceInLeaderboard(
	roomId protocol.RoomID,
	duration int64,
) (int64, error) {
	// Define the leaderboard key for the specific roomId.
	leaderboardKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyLeaderboard,
		RedisKeyPredefined,
		roomId,
	)

	// Get the range of members with scores greater than or equal to the given duration.
	rng, zRangeErr := r.client.ZRangeByScore(r.ctx, leaderboardKey, &redis.ZRangeBy{
		Min:    strconv.FormatInt(duration, 10),
		Max:    "+inf",
		Offset: 0,
		Count:  1,
	}).Result()

	// Get the total count of members in the leaderboard.
	count, zCountErr := r.client.ZCount(r.ctx, leaderboardKey, "-inf", "+inf").Result()

	if len(rng) > 0 {
		// If a member with a score less than or equal to the given duration was found,
		// get its rank in the leaderboard.
		rank, zRankErr := r.client.ZRank(r.ctx, leaderboardKey, rng[0]).Result()

		// Calculate the place (count - rank) and return it along with any errors.
		return count - rank, errors.Join(zRangeErr, zCountErr, zRankErr)
	}

	// If no member with a score less than or equal to the given duration was found,
	// the place is one greater than the total count (since ranks are 0-based).
	return count + 1, errors.Join(zRangeErr, zCountErr)
}

// GetUserPlaceInLeaderboard retrieves the user's place in the leaderboard.
func (r *Redis) GetUserPlaceInLeaderboard(
	roomId protocol.RoomID,
	userID protocol.UserID,
) (int64, error) {
	leaderboardKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyLeaderboard,
		RedisKeyPredefined,
		roomId,
	)
	count, zCountErr := r.client.ZCount(r.ctx, leaderboardKey, "-inf", "+inf").Result()
	rank, zRankErr := r.client.ZRank(r.ctx, leaderboardKey, string(userID)).Result()
	return count - rank, errors.Join(zCountErr, zRankErr)
}

// GetUsersCountInLeaderboard retrieves the count of users in the leaderboard.
func (r *Redis) GetUsersCountInLeaderboard(roomId protocol.RoomID) (int64, error) {
	leaderboardKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyLeaderboard,
		RedisKeyPredefined,
		roomId,
	)
	return r.client.ZCount(r.ctx, leaderboardKey, "-inf", "+inf").Result()
}

// GetUserPlaceInActiveSessions retrieves the user's place in active sessions.
func (r *Redis) GetUserPlaceInActiveSessions(
	roomId protocol.RoomID,
	userID protocol.UserID,
) (int64, error) {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyActiveSessions,
		RedisKeyPredefined,
		roomId,
	)
	count, zCountErr := r.client.ZCount(r.ctx, activeSessionsKey, "-inf", "+inf").Result()
	rank, zRankErr := r.client.ZRank(r.ctx, activeSessionsKey, string(userID)).Result()
	return count - rank, errors.Join(zCountErr, zRankErr)
}

// GetUsersCountInActiveSessions retrieves the count of users in active sessions.
func (r *Redis) GetUsersCountInActiveSessions(roomId protocol.RoomID) (int64, error) {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyActiveSessions,
		RedisKeyPredefined,
		roomId,
	)
	return r.client.ZCount(r.ctx, activeSessionsKey, "-inf", "+inf").Result()
}

// GetBestDurationInLeaderboard retrieves the best duration achieved by a player in the leaderboard.
func (r *Redis) GetBestDurationInLeaderboard(roomId protocol.RoomID) (int64, error) {
	leaderboardKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyLeaderboard,
		RedisKeyPredefined,
		roomId,
	)
	rng, err := r.client.ZRangeWithScores(r.ctx, leaderboardKey, -1, -1).Result()
	if len(rng) == 0 || err != nil {
		return 0, err
	}
	return int64(rng[0].Score), nil
}

// SetUserDurationToActiveSessions sets the user's duration in active sessions.
func (r *Redis) SetUserDurationToActiveSessions(
	roomId protocol.RoomID,
	userID protocol.UserID,
	duration int64,
) error {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyActiveSessions,
		RedisKeyPredefined,
		roomId,
	)
	return r.client.ZAdd(r.ctx, activeSessionsKey, &redis.Z{
		Score:  float64(duration),
		Member: string(userID),
	}).Err()
}

// RemoveUserDurationFromActiveSessions removes the user's duration from active sessions.
func (r *Redis) RemoveUserDurationFromActiveSessions(
	roomId protocol.RoomID,
	userID protocol.UserID,
) error {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		RedisKeyActiveSessions,
		RedisKeyPredefined,
		roomId,
	)
	return r.client.ZRem(r.ctx, activeSessionsKey, string(userID)).Err()
}
