package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"buttonmania.win/bot"
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

	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic recovered", "panic", r)
			os.Exit(1)
		}
	}()

	// Load config file
	err := config.LoadFiles(*configPath)
	if err != nil {
		slog.Error("failed to load config file", "error", err)
		os.Exit(1)
	}

	// Bind config struct
	conf := conf.Conf{}
	err = config.Decode(&conf)
	if err != nil {
		slog.Error("failed to bind config struct", "error", err)
		os.Exit(1)
	}

	// Initialize context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = setupContext(ctx)
	engine := gin.Default()
	debug := gin.Mode() == gin.DebugMode

	// Initialize and check errors for each component
	db, err := db.NewDB(ctx)
	if err != nil {
		slog.Error("failed to initialize db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	bot, err := bot.NewBot(ctx, engine, db, debug)
	if err != nil {
		slog.Error("failed to initialize bot", "error", err)
		os.Exit(1)
	}

	web, err := web.NewWeb(ctx, conf, engine, db, debug)
	if err != nil {
		slog.Error("failed to initialize web", "error", err)
		os.Exit(1)
	}

	// Start the web and bot components
	go func() {
		if err := web.Run(); err != nil {
			slog.Error("web server error", "error", err)
		}
	}()
	go func() {
		if err := bot.Run(); err != nil {
			slog.Error("bot error", "error", err)
		}
	}()

	slog.Info("server started", "debug", debug)

	// Graceful shutdown
	gracefulShutdown(cancel, db)
}

func setupContext(ctx context.Context) context.Context {
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

func gracefulShutdown(cancel context.CancelFunc, database interface{ Close() error }) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	sig := <-ch
	slog.Info("shutdown signal received", "signal", sig)

	// Cancel context to stop background workers
	cancel()

	// Close database connections
	if err := database.Close(); err != nil {
		slog.Error("error closing database", "error", err)
	}

	slog.Info("server stopped gracefully")
	os.Exit(0)
}
