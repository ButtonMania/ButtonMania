package bot

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWebhookRejectsInvalidSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	called := false

	srv := GinWebhookServer{
		Server:      engine,
		SecretToken: "expected-secret",
	}
	if err := srv.RegisterHandler("/bot", func(_ []byte) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("RegisterHandler failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/bot", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without secret header, got %d", rec.Code)
	}
	if called {
		t.Fatal("handler must not be invoked without valid secret")
	}

	reqOK := httptest.NewRequest(http.MethodPost, "/bot", strings.NewReader(`{}`))
	reqOK.Header.Set(telegramSecretTokenHeader, "expected-secret")
	recOK := httptest.NewRecorder()
	engine.ServeHTTP(recOK, reqOK)
	if recOK.Code != http.StatusOK || !called {
		t.Fatalf("expected 200 and handler call with valid secret, got %d (called=%v)", recOK.Code, called)
	}
}
