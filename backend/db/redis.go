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

const (
	// RedisKey represents Redis custom keys.
	RedisKeyActiveSessions RedisKey = "sessions"
	RedisKeySessionTs      RedisKey = "sessionts"
	// Session ttl handling constants
	cleanupRandChance = 5
	sessionTtlSeconds = 40
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

// closes the redis connection.
func (r *Redis) close() error {
	return r.client.Close()
}

// cleanup expired user sessions
func (r *Redis) cleanupExpiredSessions(
	activeSessionsKey string,
	sessionTsKey string,
	now int64,
) error {
	var err error
	var expiredTtlsMembers []interface{}
	expiredTtlScore := strconv.FormatInt(now-sessionTtlSeconds, 10)
	expiredTtlsResult, zrgageTsErr := r.client.ZRangeByScore(
		r.ctx,
		sessionTsKey,
		&redis.ZRangeBy{
			Min: "-inf",
			Max: expiredTtlScore,
		},
	).Result()
	for _, v := range expiredTtlsResult {
		expiredTtlsMembers = append(expiredTtlsMembers, v)
	}
	if len(expiredTtlsMembers) > 0 {
		remExpiredSessionsErr := r.client.ZRem(
			r.ctx,
			activeSessionsKey,
			expiredTtlsMembers...,
		).Err()
		remExpiredTsErr := r.client.ZRemRangeByScore(
			r.ctx,
			sessionTsKey,
			"-inf",
			expiredTtlScore,
		).Err()
		err = errors.Join(
			zrgageTsErr,
			remExpiredSessionsErr,
			remExpiredTsErr,
		)
	}
	return err
}

// retrieves the user's place in active sessions.
func (r *Redis) getUserPlaceInActiveSessions(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) (int64, error) {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		RedisKeyActiveSessions,
		roomId,
	)
	count, zCountErr := r.client.ZCount(
		r.ctx,
		activeSessionsKey,
		"-inf",
		"+inf",
	).Result()
	rank, zRankErr := r.client.ZRank(
		r.ctx,
		activeSessionsKey,
		string(userID),
	).Result()
	return count - rank, errors.Join(zCountErr, zRankErr)
}

// retrieves the count of users in active sessions.
func (r *Redis) getUsersCountInActiveSessions(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) (int64, error) {
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		RedisKeyActiveSessions,
		roomId,
	)
	return r.client.ZCount(
		r.ctx,
		activeSessionsKey,
		"-inf",
		"+inf",
	).Result()
}

// sets the user's duration in active sessions.
func (r *Redis) setUserDurationToActiveSessions(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
	duration int64,
	now int64,
) error {
	var err error
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		RedisKeyActiveSessions,
		roomId,
	)
	sessionTsKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		RedisKeySessionTs,
		roomId,
	)
	addActiveSessionsErr := r.client.ZAdd(
		r.ctx,
		activeSessionsKey,
		&redis.Z{
			Score:  float64(duration),
			Member: string(userID),
		},
	).Err()
	addSessionTsErr := r.client.ZAdd(
		r.ctx,
		sessionTsKey,
		&redis.Z{
			Score:  float64(now),
			Member: string(userID),
		},
	).Err()
	// Remove expired sessions
	if rand.Intn(cleanupRandChance) == 0 {
		err = errors.Join(
			err,
			r.cleanupExpiredSessions(activeSessionsKey, sessionTsKey, now),
		)
	}
	return errors.Join(
		err,
		addActiveSessionsErr,
		addSessionTsErr,
	)
}

// removes the user's duration from active sessions.
func (r *Redis) removeUserDurationFromActiveSessions(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
	now int64,
) error {
	var err error
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		RedisKeyActiveSessions,
		roomId,
	)
	sessionTsKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		RedisKeySessionTs,
		roomId,
	)
	remActiveSessionsErr := r.client.ZRem(
		r.ctx,
		activeSessionsKey,
		string(userID),
	).Err()
	remSessionTsErr := r.client.ZRem(
		r.ctx,
		sessionTsKey,
		string(userID),
	).Err()
	// Remove expired sessions
	if rand.Intn(cleanupRandChance) == 0 {
		err = errors.Join(
			err,
			r.cleanupExpiredSessions(activeSessionsKey, sessionTsKey, now),
		)
	}
	return errors.Join(
		err,
		remActiveSessionsErr,
		remSessionTsErr,
	)
}
