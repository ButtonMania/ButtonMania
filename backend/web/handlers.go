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
	"buttonmania.win/protocol"
	"github.com/barweiss/go-tuple"
	"github.com/gin-gonic/gin"

	initdata "github.com/Telegram-Web-Apps/init-data-golang"
)

const (
	userPayloadCountInStats = 3
)

// parseTgInitData parse string to telegram InitData structure
func (w *Web) parseTgInitData(initDataStr string) (*initdata.InitData, error) {
	token := w.ctx.Value(bot.KeyTelegramToken).(string)
	expIn := 24 * time.Hour
	// Validate telegram init data
	if err := initdata.Validate(initDataStr, token, expIn); err != nil && gin.Mode() == gin.ReleaseMode {
		err = errors.Join(errors.New("invalid telegram init data"), err)
		return nil, err
	}
	// Parse telegram init data
	initData, err := initdata.Parse(initDataStr)
	if err != nil {
		err = errors.Join(errors.New("failed to parse telegram init data"), err)
		return nil, err
	}
	return initData, err
}

// wsHandler
//
//	@Summary	Handles WebSocket connections
//	@Param		clientId	query	string	true	"Client ID"
//	@Param		roomId		query	string	true	"Room ID"
//	@Param		userId		query	string	false	"User ID"
//	@Param		locale		query	string	false	"User locale"
//	@Param		payload		query	string	false	"User payload"
//	@Param		initData	query	string	false	"Telegram init data"
//	@Router		/ws [get]
func (w *Web) wsHandler(c *gin.Context) {
	clientIdStr := c.Query("clientId")
	roomIdStr := c.Query("roomId")
	userIdStr := c.Query("userId")
	localeStr := c.Query("locale")
	payloadStr := c.Query("payload")
	initDataStr := c.Query("initData")

	// Extract parameters from telegram init data
	if len(initDataStr) > 0 {
		initData, err := w.parseTgInitData(initDataStr)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
			return
		}
		userIdStr = strconv.FormatInt(initData.User.ID, 10)
		localeStr = initData.User.LanguageCode
	}

	// Check userId
	if len(userIdStr) == 0 {
		http.Error(
			c.Writer,
			"User id not provided",
			http.StatusNotFound,
		)
		return
	}

	// Convert incoming paramters
	userID := protocol.UserID(userIdStr)
	payload := protocol.UserPayload(payloadStr)
	locale := protocol.NewUserLocale(localeStr)
	clientId := protocol.ClientID(clientIdStr)
	roomId := protocol.RoomID(roomIdStr)
	roomKey := protocol.RoomKey(tuple.New2(clientId, roomId))

	// Search for room in map
	room, exists := w.rooms[roomKey]
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
// @Param		userId		query	string	false	"User ID"
// @Param		initData	query	string	false	"Telegram init data"
// @Success	200			"ok"
// @Failure	400			"User id not provided"
// @Failure	400			"Room id not provided"
// @Failure	400			"Room id is too long"
// @Failure	400			"Client not allowed"
// @Failure	400			"Room exists"
// @Router		/api/room/create [get]
func (w *Web) createHandler(c *gin.Context) {
	clientIdStr := c.Query("clientId")
	roomIdStr := c.Query("roomId")
	userIdStr := c.Query("userId")
	initDataStr := c.Query("initData")

	// Extract user id from telegram init data
	if len(initDataStr) > 0 {
		initData, err := w.parseTgInitData(initDataStr)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
			return
		}
		userIdStr = strconv.FormatInt(initData.User.ID, 10)
	}

	// Check userId
	if len(userIdStr) == 0 {
		http.Error(
			c.Writer,
			"User id not provided",
			http.StatusNotFound,
		)
		return
	}

	// Check room id
	if roomIdStr == "" {
		http.Error(
			c.Writer,
			"Room id not provided",
			http.StatusBadRequest,
		)
		return
	} else if len(roomIdStr) > 36 {
		http.Error(
			c.Writer,
			"Room id is too long",
			http.StatusBadRequest,
		)
		return
	}

	userID := protocol.UserID(userIdStr)
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

	// Check if the room is already created
	roomKey := protocol.RoomKey(tuple.New2(clientId, roomId))
	_, exists := w.rooms[roomKey]
	if exists {
		http.Error(
			c.Writer,
			"Room exists",
			http.StatusBadRequest,
		)
		return
	}

	// Create new record in db
	err := w.db.AddCustomGameRoom(clientId, roomId, userID)
	if err != nil {
		http.Error(
			c.Writer,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// Create room and add to map
	w.rooms[roomKey], _ = NewGameRoom(clientId, roomId, w.db, nil)
	c.String(http.StatusOK, "ok")
}

// deleteHandler
//
//	@Summary	Delete game room
//	@Produce	json
//	@Param		clientId	query	string	true	"Client ID"
//	@Param		roomId		query	string	true	"Room ID"
//	@Param		userId		query	string	false	"User ID"
//	@Param		initData	query	string	false	"Telegram init data"
//	@Success	200			"ok"
//	@Failure	400			"User id not provided"
//	@Failure	400			"Room id not provided"
//	@Failure	400			"Room id is too long"
//	@Failure	400			"Room cannot be deleted"
//	@Failure	404			"Room not found"
//	@Router		/api/room/delete [get]
func (w *Web) deleteHandler(c *gin.Context) {
	clientIdStr := c.Query("clientId")
	roomIdStr := c.Query("roomId")
	userIdStr := c.Query("userId")
	initDataStr := c.Query("initData")

	// Extract user id from telegram init data
	if len(initDataStr) > 0 {
		initData, err := w.parseTgInitData(initDataStr)
		if err != nil {
			http.Error(
				c.Writer,
				err.Error(),
				http.StatusBadRequest,
			)
			return
		}
		userIdStr = strconv.FormatInt(initData.User.ID, 10)
	}

	// Check userId
	if len(userIdStr) == 0 {
		http.Error(
			c.Writer,
			"User id not provided",
			http.StatusNotFound,
		)
		return
	}

	// Check room id
	if roomIdStr == "" {
		http.Error(
			c.Writer,
			"Room id not provided",
			http.StatusBadRequest,
		)
		return
	} else if len(roomIdStr) > 36 {
		http.Error(
			c.Writer,
			"Room id is too long",
			http.StatusBadRequest,
		)
		return
	}

	userID := protocol.UserID(userIdStr)
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

	// Remove record from db
	err := w.db.RemoveCustomGameRoom(clientId, roomId, userID)
	if err != nil {
		http.Error(
			c.Writer,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// Close room and delete from map
	roomKey := protocol.RoomKey(tuple.New2(clientId, roomId))
	room, exists := w.rooms[roomKey]
	if !exists {
		http.Error(
			c.Writer,
			"Room not found",
			http.StatusNotFound,
		)
		return
	} else {
		room.closed = true
		delete(w.rooms, roomKey)
	}

	c.String(http.StatusOK, "ok")
}

// statsHandler
//
//	@Summary	Get room stats
//	@Produce	json
//	@Param		clientId	query		string	true	"Client ID"
//	@Param		roomId		query		string	true	"Room ID"
//	@Success	200			{object}	protocol.GameRoomStats
//	@Failure	400			"Room id not provided"
//	@Failure	400			"Room id is too long"
//	@Failure	404			"Room not found"
//	@Router		/api/room/stats [get]
func (w *Web) statsHandler(c *gin.Context) {
	clientIdStr := c.Query("clientId")
	roomIdStr := c.Query("roomId")

	// Check room id
	if roomIdStr == "" {
		http.Error(
			c.Writer,
			"Room id not provided",
			http.StatusBadRequest,
		)
		return
	} else if len(roomIdStr) > 36 {
		http.Error(
			c.Writer,
			"Room id is too long",
			http.StatusBadRequest,
		)
		return
	}

	// Get room by key
	clientId := protocol.ClientID(clientIdStr)
	roomId := protocol.RoomID(roomIdStr)
	roomKey := protocol.RoomKey(tuple.New2(clientId, roomId))
	room, exists := w.rooms[roomKey]
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
