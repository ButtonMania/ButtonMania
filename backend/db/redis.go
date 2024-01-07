package db

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"buttonmania.win/protocol"
	"github.com/barweiss/go-tuple"
	"github.com/go-redis/redis/v8"
)

type RedisKey string

// Redis represents the redis client.
type Redis struct {
	ctx    context.Context
	client *redis.Client
}

const (
	// RedisKey represents Redis custom keys.
	RedisKeyActiveSessions RedisKey = "sessions"
	RedisKeySessionTs      RedisKey = "sessionts"
	RedisKeyCustomRooms    RedisKey = "rooms"
	RedisKeyPayloads       RedisKey = "payloads"
	// Session ttl handling constants
	cleanupRandChance = 5
	sessionTtlSeconds = 40
)

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

// get list of best scored users payloads for gived room and client id's
func (r *Redis) getBestUsersPayloads(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	count int64,
) ([]protocol.UserPayload, error) {
	var err error
	payloads := make([]protocol.UserPayload, 0)
	activeSessionsKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		RedisKeyActiveSessions,
		roomId,
	)
	payloadsKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		RedisKeyPayloads,
		roomId,
	)
	// Get best scored users
	users, err := r.client.ZRange(
		r.ctx,
		activeSessionsKey,
		0,
		count,
	).Result()
	if err != nil && len(users) > 0 {
		payloadsStr, err := r.client.HMGet(
			r.ctx,
			payloadsKey,
			users...,
		).Result()
		if err != nil {
			for _, i := range payloadsStr {
				if i == nil {
					continue
				}
				payloads = append(payloads, protocol.UserPayload(fmt.Sprintf("%v", i)))
			}
		}
	}
	return payloads, err
}

// add user payload to set of given client and room id's
func (r *Redis) addUserPayload(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
	payload protocol.UserPayload,
) error {
	var err error
	userIdStr := string(userID)
	payloadStr := string(payload)
	payloadsKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		RedisKeyPayloads,
		roomId,
	)
	_, err = r.client.HSet(
		r.ctx,
		payloadsKey,
		userIdStr,
		payloadStr,
	).Result()
	return err
}

// remove user payload from set of given client and room id's
func (r *Redis) removeUserPayload(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) error {
	var err error
	userIdStr := string(userID)
	payloadsKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		RedisKeyPayloads,
		roomId,
	)
	// Remove record
	_, err = r.client.HDel(
		r.ctx,
		payloadsKey,
		userIdStr,
	).Result()
	return err
}

// list custom game rooms
func (r *Redis) listCustomGameRooms() ([]protocol.RoomKey, error) {
	var err error
	var roomList []protocol.RoomKey
	// Scan db and get all room hash sets
	prefix := fmt.Sprintf("*:%s", RedisKeyCustomRooms)
	trimStr := fmt.Sprintf(":%s", RedisKeyCustomRooms)
	iter := r.client.Scan(r.ctx, 0, prefix, 0).Iterator()
	for iter.Next(r.ctx) {
		// Extract client id from key
		keyStr := strings.TrimSuffix(iter.Val(), trimStr)
		keySplit := strings.Split(keyStr, ":")
		if len(keySplit) != 1 {
			err = fmt.Errorf("invalid key found: %s", keyStr)
			break
		}
		// Iterate over rooms in hash sets
		roomIter := r.client.HScan(r.ctx, iter.Val(), 0, "", 0).Iterator()
		for roomIter.Next(r.ctx) {
			// Add room key to slice
			clientId := protocol.ClientID(keySplit[0])
			roomId := protocol.RoomID(roomIter.Val())
			key := protocol.RoomKey(tuple.New2(clientId, roomId))
			roomList = append(roomList, key)
		}
	}
	return roomList, err
}

// add new user's custom game room.
func (r *Redis) createCustomRoom(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) error {
	var err error
	roomIdStr := string(roomId)
	userIdStr := string(userID)
	customRoomKey := fmt.Sprintf(
		"%s:%s",
		clientId,
		RedisKeyCustomRooms,
	)
	// Check if room already exist
	roomInRedis, err := r.client.HExists(
		r.ctx,
		customRoomKey,
		roomIdStr,
	).Result()
	if roomInRedis {
		err = errors.New("room exist")
	}
	// Create new room
	if err == nil {
		_, err = r.client.HSet(
			r.ctx,
			customRoomKey,
			roomIdStr,
			userIdStr,
		).Result()
	}
	return err
}

// remove user's custom game room.
func (r *Redis) removeCustomRoom(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) error {
	var err error
	roomIdStr := string(roomId)
	userIdStr := string(userID)
	customRoomKey := fmt.Sprintf(
		"%s:%s",
		clientId,
		RedisKeyCustomRooms,
	)
	// Check if room exist and belongs to user
	roomInRedis, err := r.client.HGet(
		r.ctx,
		customRoomKey,
		roomIdStr,
	).Result()
	if roomInRedis == "" {
		err = errors.New("room not exist")
	}
	if roomInRedis != userIdStr {
		err = errors.New("room does not belong to the user")
	}
	if err != nil {
		return nil
	}
	// Remove record
	_, err = r.client.HDel(
		r.ctx,
		customRoomKey,
		roomIdStr,
	).Result()
	return err
}
