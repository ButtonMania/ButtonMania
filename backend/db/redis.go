package db

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"buttonmania.win/protocol"
	"github.com/go-redis/redis/v8"
)

type RedisKey string

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

// NewRedis creates a new redis instance.
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

// Close closes the redis connection.
func (r *Redis) Close() error {
	return r.client.Close()
}

// GetUserPlaceInActiveSessions retrieves the user's place in active sessions.
func (r *Redis) GetUserPlaceInActiveSessions(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) (int64, error) {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s:%s",
		clientId,
		RedisKeyActiveSessions,
		RedisKeyPredefined,
		roomId,
	)
	count, zCountErr := r.client.ZCount(r.ctx, activeSessionsKey, "-inf", "+inf").Result()
	rank, zRankErr := r.client.ZRank(r.ctx, activeSessionsKey, string(userID)).Result()
	return count - rank, errors.Join(zCountErr, zRankErr)
}

// GetUsersCountInActiveSessions retrieves the count of users in active sessions.
func (r *Redis) GetUsersCountInActiveSessions(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) (int64, error) {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s:%s",
		clientId,
		RedisKeyActiveSessions,
		RedisKeyPredefined,
		roomId,
	)
	return r.client.ZCount(r.ctx, activeSessionsKey, "-inf", "+inf").Result()
}

// SetUserDurationToActiveSessions sets the user's duration in active sessions.
func (r *Redis) SetUserDurationToActiveSessions(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
	duration int64,
) error {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s:%s",
		clientId,
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
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) error {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s:%s",
		clientId,
		RedisKeyActiveSessions,
		RedisKeyPredefined,
		roomId,
	)
	return r.client.ZRem(r.ctx, activeSessionsKey, string(userID)).Err()
}
