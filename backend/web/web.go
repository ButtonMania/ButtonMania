package web

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"buttonmania.win/bot"
	"buttonmania.win/conf"
	"buttonmania.win/db"
	"buttonmania.win/protocol"
	"github.com/barweiss/go-tuple"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/glob"
	"github.com/gorilla/websocket"

	initdata "github.com/Telegram-Web-Apps/init-data-golang"
	cachecontrol "go.eigsys.de/gin-cachecontrol/v2"
)

// ContextKey is used for context keys.
type ContextKey string

// Web represents the web server.
const (
	KeySessionName    ContextKey = "sessionname"
	KeySessionSecret  ContextKey = "sessionsecret"
	KeyStaticPath     ContextKey = "staticpath"
	KeyServerPort     ContextKey = "serverport"
	KeyServerTLSCert  ContextKey = "servertlscert"
	KeyServerTLSKey   ContextKey = "servertlskey"
	KeyAllowedOrigins ContextKey = "allowedorigins"
)

type Web struct {
	ctx      context.Context
	db       *db.DB
	engine   *gin.Engine
	store    sessions.Store
	upgrader websocket.Upgrader
	rooms    map[tuple.T2[protocol.ClientID, protocol.ButtonType]]*GameRoom
}

// NewWeb creates a new Web instance.
func NewWeb(ctx context.Context, conf conf.Conf, engine *gin.Engine, db *db.DB, debug bool) (*Web, error) {
	sessionName := ctx.Value(KeySessionName).(string)
	staticPath := ctx.Value(KeyStaticPath).(string)
	sessionSecret := ctx.Value(KeySessionSecret).(string)
	allowedOrigins := ctx.Value(KeyAllowedOrigins).(string)

	// Initialize router, session storage
	store := cookie.NewStore([]byte(sessionSecret))
	rooms := make(map[tuple.T2[protocol.ClientID, protocol.ButtonType]]*GameRoom)

	// Initialize WebSocket upgrader
	originChecker := glob.MustCompile(allowedOrigins)
	headerOrigin := "Origin"
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		if debug {
			return true
		}
		origin := r.Header.Get(headerOrigin)
		return originChecker.Match(origin)
	}

	// Initialize CORS config
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOriginFunc = func(origin string) bool {
		// Don't forget to change mode in the production environment!
		if debug {
			return true
		}
		return originChecker.Match(origin)
	}

	// Initialize game rooms
	for _, c := range conf.Clients {
		for _, r := range c.Rooms {
			room, err := NewGameRoom(c.ClientId, r, db)
			if err != nil {
				return nil, err
			}
			rooms[tuple.New2(c.ClientId, r)] = room
		}
	}

	// Apply middlewares and other router parameters
	engine.SetTrustedProxies(nil)
	engine.Use(gzip.Gzip(gzip.DefaultCompression))
	engine.Use(cors.New(corsConfig))
	engine.Use(cachecontrol.New(cachecontrol.Config{
		MustRevalidate:       true,
		NoCache:              false,
		NoStore:              false,
		NoTransform:          false,
		Public:               true,
		Private:              false,
		ProxyRevalidate:      true,
		MaxAge:               cachecontrol.Duration(30 * time.Minute),
		SMaxAge:              nil,
		Immutable:            false,
		StaleWhileRevalidate: cachecontrol.Duration(2 * time.Hour),
		StaleIfError:         cachecontrol.Duration(2 * time.Hour),
	}))
	engine.Use(sessions.Sessions(sessionName, store))
	engine.Use(static.Serve("/", static.LocalFile(staticPath, true)))

	return &Web{
		ctx:      ctx,
		db:       db,
		engine:   engine,
		store:    store,
		upgrader: upgrader,
		rooms:    rooms,
	}, nil
}

// wsEndpoint handles WebSocket connections.
func (w *Web) wsEndpoint(c *gin.Context) {
	clientIdStr := c.Query("clientId")
	buttonTypeStr := c.Query("buttonType")
	telegramInitData := c.Query("initData")

	token := w.ctx.Value(bot.KeyTelegramToken).(string)
	expIn := 24 * time.Hour
	// Validate telegram init data
	if err := initdata.Validate(telegramInitData, token, expIn); err != nil && gin.Mode() == gin.ReleaseMode {
		http.Error(c.Writer, "Invalid telegram init data", http.StatusBadRequest)
		return
	}

	// Parse telegram init data
	tgData, err := initdata.Parse(telegramInitData)
	if err != nil {
		http.Error(c.Writer, "Failed to parse telegram init data", http.StatusBadRequest)
		return
	}

	// Convert incoming paramters
	userLocale := tgData.User.LanguageCode
	userID := strconv.FormatInt(tgData.User.ID, 10)
	clientId := protocol.ClientID(clientIdStr)
	buttonType := protocol.ButtonType(buttonTypeStr)

	// Search for room in map
	room, exists := w.rooms[tuple.New2(clientId, buttonType)]
	if !exists {
		http.Error(c.Writer, "Room with the provided type does not exist", http.StatusBadRequest)
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

// statsEndpoint handles API requests for getting room stats.
func (w *Web) statsEndpoint(c *gin.Context) {
	clientIdStr := c.Query("clientId")
	buttonTypeStr := c.Query("buttonType")

	clientId := protocol.ClientID(clientIdStr)
	buttonType := protocol.ButtonType(buttonTypeStr)
	room, exists := w.rooms[tuple.New2(clientId, buttonType)]
	if !exists {
		http.Error(c.Writer, "Room with the provided type does not exist", http.StatusBadRequest)
		return
	}

	stats, err := room.Stats()
	if err != nil {
		http.Error(c.Writer, "Failed to get room stats", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Run starts the web server.
func (w *Web) Run() error {
	serverPort := w.ctx.Value(KeyServerPort).(int)
	serverTLSCert := w.ctx.Value(KeyServerTLSCert).(string)
	serverTLSKey := w.ctx.Value(KeyServerTLSKey).(string)

	w.engine.GET("/ws", w.wsEndpoint)
	w.engine.GET("/api/stats", w.statsEndpoint)

	if len(serverTLSCert) > 0 && len(serverTLSKey) > 0 {
		return w.engine.RunTLS(
			":"+strconv.Itoa(serverPort),
			serverTLSCert,
			serverTLSKey,
		)
	}
	return w.engine.Run(":" + strconv.Itoa(serverPort))
}
