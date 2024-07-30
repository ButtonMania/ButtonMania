package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"buttonmania.win/bot"
	"buttonmania.win/chat"
	"buttonmania.win/conf"
	"buttonmania.win/db"
	"buttonmania.win/web"
	"github.com/alecthomas/kingpin"
	"github.com/gin-gonic/gin"
	"github.com/gookit/config/v2"
)

var (
	configPath     = kingpin.Flag(string(conf.KeyConfigPath), "Config file path.").Envar("CONFIG_PATH").Required().String()
	staticPath     = kingpin.Flag(string(web.KeyStaticPath), "Static assets folder path.").Envar("STATIC_PATH").Required().String()
	sessionName    = kingpin.Flag(string(web.KeySessionName), "Server session name.").Envar("SESSION_NAME").Default("session").String()
	sessionSecret  = kingpin.Flag(string(web.KeySessionSecret), "Server session secret phrase.").Envar("SESSION_SECRET").Default("secret").String()
	serverPort     = kingpin.Flag(string(web.KeyServerPort), "Server port.").Envar("SERVER_PORT").Default("8080").Int()
	serverTLSCert  = kingpin.Flag(string(web.KeyServerTLSCert), "Server tls cert file.").Envar("SERVER_TLS_CERT").String()
	serverTLSKey   = kingpin.Flag(string(web.KeyServerTLSKey), "Server tls key file.").Envar("SERVER_TLS_KEY").String()
	allowedOrigins = kingpin.Flag(string(web.KeyAllowedOrigins), "Allowed CORS origins.").Envar("CORS_ORIGINS").Default("*").String()
	postgresUrl    = kingpin.Flag(string(db.KeyPostgresUrl), "Postgres server url.").Envar("POSTGRES_URL").Required().String()
	redisAddress   = kingpin.Flag(string(db.KeyRedisAddress), "Redis server address.").Envar("REDIS_ADDRESS").Required().String()
	redisUsername  = kingpin.Flag(string(db.KeyRedisUsername), "Redis server username.").Envar("REDIS_USERNAME").Default("").String()
	redisPassword  = kingpin.Flag(string(db.KeyRedisPassword), "Redis server password.").Envar("REDIS_PASSWORD").Default("").String()
	redisDatabase  = kingpin.Flag(string(db.KeyRedisDatabase), "Redis server database number.").Envar("REDIS_DB").Default("0").Int()
	redisTLS       = kingpin.Flag(string(db.KeyRedisTLS), "Redis server tls connection.").Envar("REDIS_TLS").Default("0").Bool()
	tgAppURL       = kingpin.Flag(string(bot.KeyTelegramAppUrl), "Telegram app url.").Envar("TG_APP_URL").Required().String()
	tgToken        = kingpin.Flag(string(bot.KeyTelegramToken), "Telegram bot token.").Envar("TG_BOT_TOKEN").Required().String()
	tgWebhook      = kingpin.Flag(string(bot.KeyTelegramWebhook), "Telegram webhook url.").Envar("TG_WEBHOOK_URL").Default("").String()
	tgDonateTon    = kingpin.Flag(string(bot.KeyTelegramDonateTonAddress), "TON address for Telegram bot donation feature.").Envar("TG_DONATION_TON").Default("").String()
	tgDonateEth    = kingpin.Flag(string(bot.KeyTelegramDonateEthAddress), "Ethereum address for Telegram bot donation feature.").Envar("TG_DONATION_ETH").Default("").String()
	tgDonateXmr    = kingpin.Flag(string(bot.KeyTelegramDonateXmrAddress), "Monero address for Telegram bot donation feature.").Envar("TG_DONATION_XMR").Default("").String()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Panic: %v", r)
		}
	}()

	// Load config file
	err := config.LoadFiles(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	// Bind config struct
	conf := conf.Conf{}
	err = config.Decode(&conf)
	if err != nil {
		log.Fatalf("Failed to bind config struct: %v", err)
	}

	// Initialize context
	ctx := setupContext()
	engine := gin.Default()
	debug := gin.Mode() == gin.DebugMode

	// Initialize and check errors for each component
	db, err := db.NewDB(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize db: %v", err)
	}
	defer db.Close()

	chat, err := chat.NewChat(ctx, db)
	if err != nil {
		log.Fatalf("Failed to initialize chat: %v", err)
	}

	bot, err := bot.NewBot(ctx, engine, db, debug)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	web, err := web.NewWeb(ctx, conf, engine, db, chat, debug)
	if err != nil {
		log.Fatalf("Failed to initialize web: %v", err)
	}

	// Start the web and bot components
	go web.Run()
	go bot.Run()

	// Handle CTRL-C
	sigIntHandler()
}

func setupContext() context.Context {
	ctx := context.TODO()
	ctx = context.WithValue(ctx, db.KeyPostgresUrl, *postgresUrl)
	ctx = context.WithValue(ctx, db.KeyRedisAddress, *redisAddress)
	ctx = context.WithValue(ctx, db.KeyRedisUsername, *redisUsername)
	ctx = context.WithValue(ctx, db.KeyRedisPassword, *redisPassword)
	ctx = context.WithValue(ctx, db.KeyRedisDatabase, *redisDatabase)
	ctx = context.WithValue(ctx, db.KeyRedisTLS, *redisTLS)
	ctx = context.WithValue(ctx, web.KeySessionSecret, *sessionSecret)
	ctx = context.WithValue(ctx, web.KeySessionName, *sessionName)
	ctx = context.WithValue(ctx, web.KeyStaticPath, *staticPath)
	ctx = context.WithValue(ctx, web.KeyServerPort, *serverPort)
	ctx = context.WithValue(ctx, web.KeyServerTLSCert, *serverTLSCert)
	ctx = context.WithValue(ctx, web.KeyServerTLSKey, *serverTLSKey)
	ctx = context.WithValue(ctx, web.KeyAllowedOrigins, *allowedOrigins)
	ctx = context.WithValue(ctx, bot.KeyTelegramAppUrl, *tgAppURL)
	ctx = context.WithValue(ctx, bot.KeyTelegramToken, *tgToken)
	ctx = context.WithValue(ctx, bot.KeyTelegramWebhook, *tgWebhook)
	ctx = context.WithValue(ctx, bot.KeyTelegramDonateTonAddress, *tgDonateTon)
	ctx = context.WithValue(ctx, bot.KeyTelegramDonateEthAddress, *tgDonateEth)
	ctx = context.WithValue(ctx, bot.KeyTelegramDonateXmrAddress, *tgDonateXmr)
	return ctx
}

func sigIntHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	log.Println("CTRL-C; exiting")
	os.Exit(0)
}
