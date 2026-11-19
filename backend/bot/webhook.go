package bot

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mymmrac/telego"
)

const (
	telegramSecretTokenHeader = "X-Telegram-Bot-Api-Secret-Token"
	maxWebhookBodyBytes       = 1 << 20 // 1 MiB
)

type GinWebhookServer struct {
	Server      *gin.Engine
	SecretToken string
}

// Start does nothing
func (g GinWebhookServer) Start(_ string) error {
	return nil
}

// Stop does nothing
func (g GinWebhookServer) Stop(_ context.Context) error {
	return nil
}

// RegisterHandler using func or server's method
func (g GinWebhookServer) RegisterHandler(path string, handler telego.WebhookHandler) error {
	g.Server.POST(path, func(ctx *gin.Context) {
		if !validWebhookSecret(ctx.GetHeader(telegramSecretTokenHeader), g.SecretToken) {
			http.Error(ctx.Writer, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxWebhookBodyBytes)
		data, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			slog.Warn("failed to read webhook body", "error", err)
			http.Error(ctx.Writer, "Bad request", http.StatusBadRequest)
			return
		}
		if err := handler(data); err != nil {
			slog.Error("webhook handler failed", "error", err)
			http.Error(ctx.Writer, "Internal server error", http.StatusInternalServerError)
			return
		}
		ctx.Writer.WriteHeader(http.StatusOK)
	})
	return nil
}
