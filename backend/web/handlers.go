package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"time"

	"buttonmania.win/bot"
	"buttonmania.win/db"
	"buttonmania.win/protocol"
	"github.com/gin-gonic/gin"

	initdata "github.com/Telegram-Web-Apps/init-data-golang"
	tuple "github.com/barweiss/go-tuple"
)

const (
	userPayloadCountInStats = 3
	maxRoomIdLength         = 36
)

// Define authentication errors
var (
	ErrInitDataMissing = errors.New("telegram init data not provided")
	ErrInitDataInvalid = errors.New("invalid telegram init data")
	ErrInitDataNoUser  = errors.New("telegram init data contains no user")
)

// parseTgInitData parse string to telegram InitData structure
func (w *Web) parseTgInitData(initDataStr string) (*initdata.InitData, error) {
	token := w.ctx.Value(bot.KeyTelegramToken).(string)
	expIn := 24 * time.Hour
	// Validate telegram init data. The signature check is only skipped in
	// debug mode, so a development client can be used without a bot token.
	if err := initdata.Validate(initDataStr, token, expIn); err != nil && !w.debug {
		return nil, errors.Join(ErrInitDataInvalid, err)
	}
	// Parse telegram init data
	initData, err := initdata.Parse(initDataStr)
	if err != nil {
		return nil, errors.Join(ErrInitDataInvalid, err)
	}
	if initData.User == nil {
		return nil, ErrInitDataNoUser
	}
	return initData, nil
}

// authenticate resolves the caller identity from signed telegram init data.
// The user id is never taken from a plain query parameter outside debug mode:
// doing so would let anyone act as any user.
func (w *Web) authenticate(c *gin.Context) (protocol.UserID, protocol.UserLocale, error) {
	initDataStr := c.Query("initData")
	if len(initDataStr) == 0 {
		if w.debug {
			userIdStr := c.Query("userId")
			if len(userIdStr) == 0 {
				return "", "", ErrInitDataMissing
			}
			return protocol.UserID(userIdStr), protocol.NewUserLocale(c.Query("locale")), nil
		}
		return "", "", ErrInitDataMissing
	}

	initData, err := w.parseTgInitData(initDataStr)
	if err != nil {
		return "", "", err
	}

	userID := protocol.UserID(strconv.FormatInt(initData.User.ID, 10))
	locale := protocol.NewUserLocale(initData.User.LanguageCode)
	return userID, locale, nil
}

// validateRoomId checks the room id and writes an error response if invalid.
func validateRoomId(c *gin.Context, roomIdStr string) bool {
	if roomIdStr == "" {
		http.Error(c.Writer, "Room id not provided", http.StatusBadRequest)
		return false
	}
	if len(roomIdStr) > maxRoomIdLength {
		http.Error(c.Writer, "Room id is too long", http.StatusBadRequest)
		return false
	}
	return true
}

// @Summary	Handles WebSocket connections
// @Param		clientId	query	string	true	"Client ID"
// @Param		roomId		query	string	true	"Room ID"
// @Param		payload		query	string	false	"User payload"
// @Param		initData	query	string	true	"Telegram init data"
// @Router		/ws [get]
func (w *Web) wsHandler(c *gin.Context) {
	clientIdStr := c.Query("clientId")
	roomIdStr := c.Query("roomId")
	payloadStr := c.Query("payload")

	userID, locale, err := w.authenticate(c)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusUnauthorized)
		return
	}

	// Convert incoming paramters
	payload := protocol.UserPayload(payloadStr)
	clientId := protocol.ClientID(clientIdStr)
	roomId := protocol.RoomID(roomIdStr)
	roomKey := protocol.RoomKey(tuple.New2(clientId, roomId))

	// Search for room in map
	room, exists := w.getRoom(roomKey)
	if !exists {
		http.Error(
			c.Writer,
			"Room not found",
			http.StatusNotFound,
		)
		return
	}

	ws, err := w.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to create WebSocket connection", err)
		http.Error(c.Writer, "Failed to create WebSocket connection", http.StatusInternalServerError)
		return
	}
	defer ws.Close()

	if err := room.MaintainGameSession(userID, payload, locale, ws); err != nil {
		log.Println("Error occurred while maintaining the game session:", err)
		return
	}
}

// @Summary	Create game room
// @Produce	json
// @Param		clientId	query	string	true	"Client ID"
// @Param		roomId		query	string	true	"Room ID"
// @Param		initData	query	string	true	"Telegram init data"
// @Success	200			"ok"
// @Failure	400			"Room id not provided"
// @Failure	400			"Room id is too long"
// @Failure	400			"Client not allowed"
// @Failure	400			"Room exists"
// @Failure	401			"Telegram init data not provided"
// @Router		/api/room/create [get]
func (w *Web) createRoomHandler(c *gin.Context) {
	clientIdStr := c.Query("clientId")
	roomIdStr := c.Query("roomId")

	userID, _, err := w.authenticate(c)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusUnauthorized)
		return
	}

	if !validateRoomId(c, roomIdStr) {
		return
	}

	clientId := protocol.ClientID(clientIdStr)
	roomId := protocol.RoomID(roomIdStr)
	// Check if client allowed
	if !slices.Contains(w.clients, clientId) {
		http.Error(
			c.Writer,
			"Client not allowed",
			http.StatusBadRequest,
		)
		return
	}

	// Create new record in db. The store rejects a duplicate itself, so no
	// separate existence check is needed here.
	if err := w.db.AddCustomGameRoom(clientId, roomId, userID); err != nil {
		if errors.Is(err, db.ErrRoomExists) {
			http.Error(c.Writer, "Room exists", http.StatusBadRequest)
			return
		}
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create room and add to map
	room, err := NewGameRoom(clientId, roomId, w.db, nil)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	roomKey := protocol.RoomKey(tuple.New2(clientId, roomId))
	if !w.addRoom(roomKey, room) {
		http.Error(c.Writer, "Room exists", http.StatusBadRequest)
		return
	}

	c.String(http.StatusOK, "ok")
}

// @Summary	Delete game room
// @Produce	json
// @Param		clientId	query	string	true	"Client ID"
// @Param		roomId		query	string	true	"Room ID"
// @Param		initData	query	string	true	"Telegram init data"
// @Success	200			"ok"
// @Failure	400			"Room id not provided"
// @Failure	400			"Room id is too long"
// @Failure	400			"Room cannot be deleted"
// @Failure	401			"Telegram init data not provided"
// @Failure	403			"Room does not belong to the user"
// @Failure	404			"Room not found"
// @Router		/api/room/delete [get]
func (w *Web) deleteRoomHandler(c *gin.Context) {
	clientIdStr := c.Query("clientId")
	roomIdStr := c.Query("roomId")

	userID, _, err := w.authenticate(c)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusUnauthorized)
		return
	}

	if !validateRoomId(c, roomIdStr) {
		return
	}

	clientId := protocol.ClientID(clientIdStr)
	roomId := protocol.RoomID(roomIdStr)
	// Check room key (predefined rooms cannot be deleted)
	for _, clientsConf := range w.conf.Clients {
		for _, r := range clientsConf.Rooms {
			if clientId == clientsConf.ClientId && r == roomId {
				http.Error(
					c.Writer,
					"Room cannot be deleted",
					http.StatusBadRequest,
				)
				return
			}
		}
	}

	// Remove record from db. Ownership is enforced by the store.
	if err := w.db.RemoveCustomGameRoom(clientId, roomId, userID); err != nil {
		switch {
		case errors.Is(err, db.ErrRoomNotExist):
			http.Error(c.Writer, "Room not found", http.StatusNotFound)
		case errors.Is(err, db.ErrRoomNotOwned):
			http.Error(c.Writer, "Room does not belong to the user", http.StatusForbidden)
		default:
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Close room and delete from map
	roomKey := protocol.RoomKey(tuple.New2(clientId, roomId))
	room, exists := w.removeRoom(roomKey)
	if !exists {
		http.Error(
			c.Writer,
			"Room not found",
			http.StatusNotFound,
		)
		return
	}
	room.Close()

	c.String(http.StatusOK, "ok")
}

// @Summary	Get room stats
// @Produce	json
// @Param		clientId	query		string	true	"Client ID"
// @Param		roomId		query		string	true	"Room ID"
// @Success	200			{object}	protocol.GameRoomStats
// @Failure	400			"Room id not provided"
// @Failure	400			"Room id is too long"
// @Failure	404			"Room not found"
// @Router		/api/room/stats [get]
func (w *Web) statsRoomHandler(c *gin.Context) {
	clientIdStr := c.Query("clientId")
	roomIdStr := c.Query("roomId")

	if !validateRoomId(c, roomIdStr) {
		return
	}

	// Get room by key
	clientId := protocol.ClientID(clientIdStr)
	roomId := protocol.RoomID(roomIdStr)
	roomKey := protocol.RoomKey(tuple.New2(clientId, roomId))
	room, exists := w.getRoom(roomKey)
	if !exists {
		http.Error(
			c.Writer,
			"Room not found",
			http.StatusNotFound,
		)
		return
	}

	// Retrive room stats
	stats, err := room.Stats(userPayloadCountInStats)
	if err != nil {
		http.Error(
			c.Writer,
			fmt.Sprintln("Failed to get room stats:", err),
			http.StatusInternalServerError,
		)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary	Get client stats
// @Produce	json
// @Param		clientId	query		string	true	"Client ID"
// @Success	200			{object}	protocol.ClientStats
// @Failure	404			"Client not found"
// @Router		/api/stats [get]
func (w *Web) statsHandler(c *gin.Context) {
	exists := false
	clientIdStr := c.Query("clientId")
	clientId := protocol.ClientID(clientIdStr)

	for _, clientsConf := range w.conf.Clients {
		if clientId == clientsConf.ClientId {
			exists = true
			break
		}
	}

	if !exists {
		http.Error(
			c.Writer,
			"Client not found",
			http.StatusNotFound,
		)
		return
	}

	// Get rooms by key
	customRooms, err := w.db.ListCustomGameRooms()
	if err != nil {
		http.Error(
			c.Writer,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	onlineUsersCount, err := w.db.GetOnlineUsersCount(clientId)
	if err != nil {
		http.Error(
			c.Writer,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	usersOnline := int64(onlineUsersCount)
	roomsCount := int64(len(customRooms))
	stats := protocol.NewClientStats(&usersOnline, &roomsCount)
	c.JSON(http.StatusOK, stats)
}
