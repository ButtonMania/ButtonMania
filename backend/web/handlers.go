package web

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"buttonmania.win/bot"
	"buttonmania.win/protocol"
	"github.com/barweiss/go-tuple"
	"github.com/gin-gonic/gin"

	initdata "github.com/Telegram-Web-Apps/init-data-golang"
)

// wsHandler
//
//	@Summary	Handles WebSocket connections
//	@Param		clientId	query	string	true	"Client ID"
//	@Param		roomId		query	string	true	"Room ID"
//	@Param		initData	query	string	true	"Telegram init data"
//	@Router		/ws [get]
func (w *Web) wsHandler(c *gin.Context) {
	clientIdStr := c.Query("clientId")
	roomIdStr := c.Query("roomId")
	initDataStr := c.Query("initData")
	// Check init data for empty value
	if len(initDataStr) == 0 {
		http.Error(c.Writer, "Empty telegram init data provided", http.StatusBadRequest)
		return
	}

	token := w.ctx.Value(bot.KeyTelegramToken).(string)
	expIn := 24 * time.Hour
	// Validate telegram init data
	if err := initdata.Validate(initDataStr, token, expIn); err != nil && gin.Mode() == gin.ReleaseMode {
		http.Error(c.Writer, "Invalid telegram init data", http.StatusBadRequest)
		return
	}

	// Parse telegram init data
	initData, err := initdata.Parse(initDataStr)
	if err != nil {
		http.Error(c.Writer, "Failed to parse telegram init data", http.StatusBadRequest)
		return
	}

	// Convert incoming paramters
	userLocale := initData.User.LanguageCode
	userID := strconv.FormatInt(initData.User.ID, 10)
	clientId := protocol.ClientID(clientIdStr)
	roomId := protocol.RoomID(roomIdStr)

	// Search for room in map
	room, exists := w.rooms[tuple.New2(clientId, roomId)]
	if !exists {
		http.Error(c.Writer, "Room with the provided id does not exist", http.StatusBadRequest)
		return
	}

	ws, err := w.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to create WebSocket connection", err)
		http.Error(c.Writer, "Failed to create WebSocket connection", http.StatusInternalServerError)
		return
	}
	defer ws.Close()

	tgID := protocol.UserID(userID)
	locale := protocol.NewUserLocale(userLocale)
	if err := room.MaintainGameSession(tgID, locale, ws); err != nil {
		log.Println("Error occurred while maintaining the game session:", err)
		return
	}
}

// statsHandler
//
//	@Summary	Handles API requests for getting room stats
//	@Produce	json
//	@Param		clientId	query		string	true	"Client ID"
//	@Param		roomId		query		string	true	"Room ID"
//	@Success	200			{object}	protocol.GameRoomStats
//	@Router		/api/room/stats [get]
func (w *Web) statsHandler(c *gin.Context) {
	clientIdStr := c.Query("clientId")
	roomIdStr := c.Query("roomId")

	clientId := protocol.ClientID(clientIdStr)
	roomId := protocol.RoomID(roomIdStr)
	room, exists := w.rooms[tuple.New2(clientId, roomId)]
	if !exists {
		http.Error(
			c.Writer,
			"Room with the provided id does not exist",
			http.StatusBadRequest,
		)
		return
	}

	stats, err := room.Stats()
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
