package db

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"buttonmania.win/protocol"

	tuple "github.com/barweiss/go-tuple"
	redis "github.com/go-redis/redis/v8"
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
	RedisKeyChat           RedisKey = "chat"
	// Session ttl handling constants
	cleanupRandChance      = 5
	sessionTtlSeconds      = 40
	maxChatMessageInStream = 5
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

// retrieves the count of online users of gived client.
func (r *Redis) getOnlineUsersCount(
	clientId protocol.ClientID,
) (int64, error) {
	var total int64
	pattern := fmt.Sprintf(
		"%s:%s:*",
		clientId,
		RedisKeyActiveSessions,
	)
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return total, err
	}
	for _, key := range keys {
		count, err := r.client.ZCount(
			r.ctx,
			key,
			"-inf",
			"+inf",
		).Result()
		if err != nil {
			return total, err
		}
		total += count
	}
	return total, nil
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
	if err == nil && len(users) > 0 {
		payloadsStr, err := r.client.HMGet(
			r.ctx,
			payloadsKey,
			users...,
		).Result()
		if err == nil {
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

// removeOwnedRoomScript deletes a room only when it belongs to the caller, so
// the ownership check and the delete cannot be interleaved by another request.
// Returns 1 on delete, 0 when the room is not owned, -1 when it does not exist.
var removeOwnedRoomScript = redis.NewScript(`
	local owner = redis.call('HGET', KEYS[1], ARGV[1])
	if not owner then
		return -1
	end
	if owner ~= ARGV[2] then
		return 0
	end
	redis.call('HDEL', KEYS[1], ARGV[1])
	return 1
`)

// add new user's custom game room.
func (r *Redis) createCustomRoom(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) error {
	roomIdStr := string(roomId)
	userIdStr := string(userID)
	customRoomKey := fmt.Sprintf(
		"%s:%s",
		clientId,
		RedisKeyCustomRooms,
	)
	// Create the room only if it does not exist yet, in a single operation.
	created, err := r.client.HSetNX(
		r.ctx,
		customRoomKey,
		roomIdStr,
		userIdStr,
	).Result()
	if err != nil {
		return err
	}
	if !created {
		return ErrRoomExists
	}
	return nil
}

// remove user's custom game room.
func (r *Redis) removeCustomRoom(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) error {
	roomIdStr := string(roomId)
	userIdStr := string(userID)
	customRoomKey := fmt.Sprintf(
		"%s:%s",
		clientId,
		RedisKeyCustomRooms,
	)
	result, err := removeOwnedRoomScript.Run(
		r.ctx,
		r.client,
		[]string{customRoomKey},
		roomIdStr,
		userIdStr,
	).Int64()
	if err != nil {
		return err
	}
	switch result {
	case -1:
		return ErrRoomNotExist
	case 0:
		return ErrRoomNotOwned
	}
	return nil
}

// init user's chat consumer group
func (r *Redis) initChatConsumerGroup(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) error {
	streamKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		roomId,
		RedisKeyChat,
	)
	return r.client.XGroupCreateMkStream(
		r.ctx,
		streamKey,
		string(roomId),
		"$",
	).Err()
}

// add user consumer to consumer group of chat stream
func (r *Redis) addChatConsumerToGroup(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userId protocol.UserID,
) error {
	streamKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		roomId,
		RedisKeyChat,
	)
	return r.client.XGroupCreateConsumer(
		r.ctx,
		streamKey,
		string(roomId),
		string(userId),
	).Err()
}

// push user's chat message.
func (r *Redis) pushChatMessage(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	chatMessage protocol.ChatMessage,
) error {
	streamKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		roomId,
		RedisKeyChat,
	)
	return r.client.XAdd(r.ctx, &redis.XAddArgs{
		MaxLen: maxChatMessageInStream,
		Stream: streamKey,
		Values: []interface{}{
			"UserID", string(chatMessage.UserID),
			"Message", string(chatMessage.Message),
		},
	}).Err()
}

// pop user's chat message
func (r *Redis) popChatMessage(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userId protocol.UserID,
) (protocol.ChatMessage, error) {
	var msg protocol.ChatMessage
	streamKey := fmt.Sprintf(
		"%s:%s:%s",
		clientId,
		roomId,
		RedisKeyChat,
	)
	result, err := r.client.XReadGroup(r.ctx, &redis.XReadGroupArgs{
		Streams:  []string{streamKey, ">"},
		Group:    string(roomId),
		Consumer: string(userId),
		Block:    time.Second,
		Count:    10,
		NoAck:    true,
	}).Result()
	if err == nil {
		for _, s := range result {
			for _, message := range s.Messages {
				// Stream values are untyped: a malformed entry must not panic
				// the session goroutine, which would take the process down.
				message_, msgOk := message.Values["Message"].(string)
				userIdRaw, userOk := message.Values["UserID"].(string)
				err = r.client.XAck(
					r.ctx,
					streamKey,
					string(roomId),
					message.ID,
				).Err()
				if !msgOk || !userOk {
					continue
				}

				userid_ := protocol.UserID(userIdRaw)
				if userid_ != userId {
					msg.Message = message_
					msg.UserID = userid_
					break
				}
			}
		}
	}
	return msg, err
}
