package web

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"buttonmania.win/chat"
	"buttonmania.win/conf"
	"buttonmania.win/db"
	"buttonmania.win/localization"
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

	_ "buttonmania.win/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	conf     conf.Conf
	db       *db.DB
	chat     *chat.Chat
	engine   *gin.Engine
	store    sessions.Store
	upgrader websocket.Upgrader
	clients  []protocol.ClientID
	rooms    map[protocol.RoomKey]*GameRoom
}

// NewWeb creates a new Web instance.
func NewWeb(
	ctx context.Context,
	conf conf.Conf,
	engine *gin.Engine,
	db *db.DB,
	chat *chat.Chat,
	debug bool,
) (*Web, error) {
	sessionName := ctx.Value(KeySessionName).(string)
	staticPath := ctx.Value(KeyStaticPath).(string)
	sessionSecret := ctx.Value(KeySessionSecret).(string)
	allowedOrigins := ctx.Value(KeyAllowedOrigins).(string)

	// Initialize router, session storage
	store := cookie.NewStore([]byte(sessionSecret))
	rooms := make(map[protocol.RoomKey]*GameRoom)
	clients := make([]protocol.ClientID, 0)

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

	// Initialize predefined game rooms
	for _, c := range conf.Clients {
		for _, r := range c.Rooms {
			msgLoc, err := localization.NewMessagesLocalization(c.ClientId, r)
			if err != nil {
				return nil, err
			}
			roomKey := protocol.RoomKey(tuple.New2(c.ClientId, r))
			rooms[roomKey], _ = NewGameRoom(c.ClientId, r, db, chat, msgLoc)
		}
		clients = append(clients, c.ClientId)
	}

	// Initialize user created game rooms
	customRooms, err := db.ListCustomGameRooms()
	for _, roomKey := range customRooms {
		rooms[roomKey], _ = NewGameRoom(roomKey.V1, roomKey.V2, db, chat, nil)
	}

	// Apply middlewares and other router parameters
	engine.SetTrustedProxies(nil)
	engine.Use(static.Serve("/", static.LocalFile(staticPath, true)))
	engine.Use(sessions.Sessions(sessionName, store))
	engine.Use(gzip.Gzip(gzip.DefaultCompression))
	engine.Use(cors.New(corsConfig))
	engine.Use(cachecontrol.New(cachecontrol.Config{
		MustRevalidate:       true,
		NoCache:              true,
		NoStore:              true,
		NoTransform:          false,
		Public:               true,
		Private:              false,
		ProxyRevalidate:      true,
		Immutable:            false,
		SMaxAge:              cachecontrol.Duration(30 * time.Minute),
		MaxAge:               cachecontrol.Duration(20 * time.Minute),
		StaleWhileRevalidate: cachecontrol.Duration(2 * time.Hour),
		StaleIfError:         cachecontrol.Duration(2 * time.Hour),
	}))

	return &Web{
		ctx:      ctx,
		conf:     conf,
		db:       db,
		engine:   engine,
		store:    store,
		upgrader: upgrader,
		clients:  clients,
		rooms:    rooms,
	}, err
}

// Run
//
//	@title			ButtonMania API
//	@version		1.0
//	@contact.name	ButtonMania Team
//	@contact.email	team@buttonmania.win
//	@host			buttonmania.win
//	@BasePath		/
func (w *Web) Run() error {
	serverPort := w.ctx.Value(KeyServerPort).(int)
	serverTLSCert := w.ctx.Value(KeyServerTLSCert).(string)
	serverTLSKey := w.ctx.Value(KeyServerTLSKey).(string)

	w.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	w.engine.GET("/ws", w.wsHandler)
	w.engine.GET("/api/room/create", w.createHandler)
	w.engine.GET("/api/room/delete", w.deleteHandler)
	w.engine.GET("/api/room/stats", w.statsHandler)

	if len(serverTLSCert) > 0 && len(serverTLSKey) > 0 {
		return w.engine.RunTLS(
			":"+strconv.Itoa(serverPort),
			serverTLSCert,
			serverTLSKey,
		)
	}

	// Run starts the web server.
	return w.engine.Run(":" + strconv.Itoa(serverPort))
}
