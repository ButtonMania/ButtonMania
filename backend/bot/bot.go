package bot

import (
	"context"
	"log"
	"net/url"

	"buttonmania.win/db"
	"buttonmania.win/localization"
	"buttonmania.win/protocol"
	"github.com/gin-gonic/gin"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

// ContextKey is used for context keys.
type ContextKey string

const (
	// Context keys for configuration.
	KeyTelegramToken   ContextKey = ContextKey("telegramtoken")
	KeyTelegramAppUrl  ContextKey = ContextKey("telegramappurl")
	KeyTelegramWebhook ContextKey = ContextKey("telegramwebhook")
	// Context keys for donation cryptocurrency addresses
	KeyTelegramDonateTonAddress ContextKey = ContextKey("telegramdonateton")
	KeyTelegramDonateEthAddress ContextKey = ContextKey("telegramdonateeth")
	KeyTelegramDonateXmrAddress ContextKey = ContextKey("telegramdonatexmr")
)

// Bot represents a Telegram bot.
type Bot struct {
	ctx    context.Context
	db     *db.DB
	engine *gin.Engine
	bot    *telego.Bot
	loc    *localization.BotLocalization
}

// NewBot creates a new instance of Bot.
func NewBot(ctx context.Context, engine *gin.Engine, db *db.DB, debug bool) (*Bot, error) {
	var options telego.BotOption
	if debug {
		options = telego.WithDefaultDebugLogger()
	} else {
		options = telego.WithDiscardLogger()
	}

	telegramToken := ctx.Value(KeyTelegramToken).(string)
	bot, err := telego.NewBot(telegramToken, options)
	if err != nil {
		return nil, err
	}

	loc, err := localization.NewBotLocalization()
	if err != nil {
		return nil, err
	}

	return &Bot{
		ctx:    ctx,
		db:     db,
		engine: engine,
		bot:    bot,
		loc:    loc,
	}, nil
}

// handleStartCommand handles the "/start" command.
func (b *Bot) handleStartCommand(bot *telego.Bot, update telego.Update) {
	telegramAppUrl := b.ctx.Value(KeyTelegramAppUrl).(string)

	locale := protocol.NewUserLocale(update.Message.From.LanguageCode)
	commandString := b.loc.LocalizedCommandString(locale, "start")

	keyboard := telegoutil.InlineKeyboard(
		telegoutil.InlineKeyboardRow(
			telegoutil.InlineKeyboardButton("Push the Button").WithURL(telegramAppUrl),
		),
	)
	message := telegoutil.MessageWithEntities(
		telegoutil.ID(update.Message.Chat.ID),
		telegoutil.Entityf(commandString, update.Message.From.Username),
	).WithParseMode("HTML").WithReplyMarkup(keyboard)

	_, _ = b.bot.SendMessage(message)
}

// handleStartCommand handles the "/donate" command.
func (b *Bot) handleDonateCommand(bot *telego.Bot, update telego.Update) {
	tonAddress := b.ctx.Value(KeyTelegramDonateTonAddress).(string)
	ethAddress := b.ctx.Value(KeyTelegramDonateEthAddress).(string)
	xmrAddress := b.ctx.Value(KeyTelegramDonateXmrAddress).(string)

	locale := protocol.NewUserLocale(update.Message.From.LanguageCode)
	commandString := b.loc.LocalizedCommandString(locale, "donate")

	message := telegoutil.MessageWithEntities(
		telegoutil.ID(update.Message.Chat.ID),
		telegoutil.Entityf(commandString, tonAddress, ethAddress, xmrAddress),
	).WithParseMode("HTML")

	_, _ = b.bot.SendMessage(message)
}

// handleUnknownCommand handles unknown commands.
func (b *Bot) handleUnknownCommand(bot *telego.Bot, update telego.Update) {
	_, _ = b.bot.SendMessage(telegoutil.Message(
		telegoutil.ID(update.Message.Chat.ID),
		"Unknown command, use /start",
	))
}

// RunWithUpdates starts the bot using update channels.
func (b *Bot) RunWithUpdates(updates <-chan telego.Update) error {
	bh, err := telegohandler.NewBotHandler(b.bot, updates)
	if err != nil {
		return err
	}
	defer bh.Stop()

	bh.Handle(b.handleStartCommand, telegohandler.CommandEqual("start"))
	bh.Handle(b.handleDonateCommand, telegohandler.CommandEqual("donate"))
	bh.Handle(b.handleUnknownCommand, telegohandler.AnyCommand())
	bh.Start()
	return nil
}

// RunWithWebhook starts the bot using a webhook.
func (b *Bot) RunWithWebhook(webhookUrl string) error {
	bot := b.bot
	token := bot.Token()

	u, err := url.Parse(webhookUrl)
	if err != nil {
		return err
	}

	err = bot.SetWebhook(&telego.SetWebhookParams{
		URL: webhookUrl + token,
	})
	if err != nil {
		return err
	}

	if info, err := bot.GetWebhookInfo(); err == nil {
		log.Println("Webhook Info:", info)
	} else {
		return err
	}

	webhookOpts := telego.WithWebhookServer(&GinWebhookServer{
		Server: b.engine,
	})

	updates, err := bot.UpdatesViaWebhook(u.Path+token, webhookOpts)
	if err != nil {
		return err
	}

	go func() {
		_ = bot.StartWebhook("")
	}()

	defer func() {
		_ = bot.StopWebhook()
	}()

	return b.RunWithUpdates(updates)
}

// RunWithLongPolling starts the bot using long polling.
func (b *Bot) RunWithLongPolling() error {
	// Remove webhook
	err := b.bot.DeleteWebhook(&telego.DeleteWebhookParams{
		DropPendingUpdates: false,
	})
	if err != nil {
		return err
	}
	// Setup long polling
	updates, err := b.bot.UpdatesViaLongPolling(nil)
	if err != nil {
		return err
	}
	defer b.bot.StopLongPolling()
	return b.RunWithUpdates(updates)
}

// Run starts the bot using the appropriate method (webhook or long polling).
func (b *Bot) Run() error {
	telegramWebhook := b.ctx.Value(KeyTelegramWebhook).(string)
	if len(telegramWebhook) > 0 {
		return b.RunWithWebhook(telegramWebhook)
	} else {
		return b.RunWithLongPolling()
	}
}
