package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"buttonmania.win/bot"
	"buttonmania.win/db"
	"buttonmania.win/web"
	"github.com/alecthomas/kingpin"
	"github.com/gin-gonic/gin"
)

var (
	redisAddress   = kingpin.Flag("redisaddress", "Redis server address.").Envar("REDIS_ADDRESS").Required().String()
	redisUsername  = kingpin.Flag("redisusername", "Redis server username.").Envar("REDIS_USERNAME").Default("").String()
	redisPassword  = kingpin.Flag("redispassword", "Redis server password.").Envar("REDIS_PASSWORD").Default("").String()
	redisDatabase  = kingpin.Flag("redisdatabase", "Redis server database number.").Envar("REDIS_DB").Default("0").Int()
	redisTLS       = kingpin.Flag("redistls", "Redis server tls connection.").Envar("REDIS_TLS").Default("0").Bool()
	staticPath     = kingpin.Flag("staticpath", "Static assets folder path.").Envar("STATIC_PATH").Required().String()
	sessionName    = kingpin.Flag("sessionname", "Server session name.").Envar("SESSION_NAME").Default("session").String()
	sessionSecret  = kingpin.Flag("sessionsecret", "Server session secret phrase.").Envar("SESSION_SECRET").Default("secret").String()
	serverPort     = kingpin.Flag("serverport", "Server port.").Envar("SERVER_PORT").Default("8080").Int()
	allowedOrigins = kingpin.Flag("allowedorigins", "Allowed CORS origins.").Envar("CORS_ORIGINS").Default("*").String()
	tgAppURL       = kingpin.Flag("telegramappurl", "Telegram app url.").Envar("TG_APP_URL").Required().String()
	tgToken        = kingpin.Flag("telegramtoken", "Telegram bot token.").Envar("TG_BOT_TOKEN").Required().String()
	tgWebhook      = kingpin.Flag("telegramwebhook", "Telegram webhook url.").Envar("TG_WEBHOOK_URL").Default("").String()
	tgWebhookPort  = kingpin.Flag("telegramwhport", "Telegram webhook listen port.").Envar("TG_WEBHOOK_PORT").Default("8080").Int()
	tgDonateTon    = kingpin.Flag("telegramdonateton", "TON address for Telegram bot donation feature.").Envar("TG_DONATION_TON").Default("").String()
	tgDonateEth    = kingpin.Flag("telegramdonateeth", "Ethereum address for Telegram bot donation feature.").Envar("TG_DONATION_ETH").Default("").String()
	tgDonateXmr    = kingpin.Flag("telegramdonatexmr", "Monero address for Telegram bot donation feature.").Envar("TG_DONATION_XMR").Default("").String()
)

func setupContext() context.Context {
	ctx := context.TODO()
	ctx = context.WithValue(ctx, db.KeyRedisAddress, *redisAddress)
	ctx = context.WithValue(ctx, db.KeyRedisUsername, *redisUsername)
	ctx = context.WithValue(ctx, db.KeyRedisPassword, *redisPassword)
	ctx = context.WithValue(ctx, db.KeyRedisDatabase, *redisDatabase)
	ctx = context.WithValue(ctx, db.KeyRedisTLS, *redisTLS)
	ctx = context.WithValue(ctx, web.KeySessionSecret, *sessionSecret)
	ctx = context.WithValue(ctx, web.KeySessionName, *sessionName)
	ctx = context.WithValue(ctx, web.KeyStaticPath, *staticPath)
	ctx = context.WithValue(ctx, web.KeyServerPort, *serverPort)
	ctx = context.WithValue(ctx, web.KeyAllowedOrigins, *allowedOrigins)
	ctx = context.WithValue(ctx, bot.KeyTelegramAppUrl, *tgAppURL)
	ctx = context.WithValue(ctx, bot.KeyTelegramToken, *tgToken)
	ctx = context.WithValue(ctx, bot.KeyTelegramWebhook, *tgWebhook)
	ctx = context.WithValue(ctx, bot.KeyTelegramWhPork, *tgWebhookPort)
	ctx = context.WithValue(ctx, bot.KeyTelegramDonateTonAddress, *tgDonateTon)
	ctx = context.WithValue(ctx, bot.KeyTelegramDonateEthAddress, *tgDonateEth)
	ctx = context.WithValue(ctx, bot.KeyTelegramDonateXmrAddress, *tgDonateXmr)
	return ctx
}
func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Panic: %v", r)
		}
	}()

	// Initialize context
	ctx := setupContext()
	debug := gin.Mode() == gin.DebugMode

	// Initialize and check errors for each component
	db, err := db.NewDB(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize db: %v", err)
	}
	defer db.Close()

	web, err := web.NewWeb(ctx, db, debug)
	if err != nil {
		log.Fatalf("Failed to initialize web: %v", err)
	}

	bot, err := bot.NewBot(ctx, db, debug)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	// Start the web and bot components
	go web.Run()
	go bot.Run()

	// Handle CTRL-C
	sigIntHandler()
}

func sigIntHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	log.Println("CTRL-C; exiting")
	os.Exit(0)
}
