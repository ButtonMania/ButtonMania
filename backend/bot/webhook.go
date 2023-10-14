package bot

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mymmrac/telego"
)

type GinWebhookServer struct {
	Server *gin.Engine
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
		var data []byte
		var err error
		if data, err = io.ReadAll(ctx.Request.Body); err != nil {
			errString := fmt.Sprintf("Webhook handler: %s", err)
			log.Println(errString)
			http.Error(ctx.Writer, errString, http.StatusInternalServerError)
			return
		}
		if err = handler(data); err != nil {
			errString := fmt.Sprintf("Webhook handler: %s", err)
			log.Println(errString)
			http.Error(ctx.Writer, errString, http.StatusInternalServerError)
			return
		}
		ctx.Writer.WriteHeader(http.StatusOK)
	})
	return nil
}
