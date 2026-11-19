package web

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func testContext(target string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", target, nil)
	return c
}

// Outside debug mode a plain userId query parameter must not authenticate
// anyone: it would let a caller act as any telegram user.
func TestAuthenticateRejectsQueryUserId(t *testing.T) {
	web := &Web{debug: false}

	_, _, err := web.authenticate(testContext("/ws?userId=42&roomId=love"))
	if !errors.Is(err, ErrInitDataMissing) {
		t.Fatalf("expected ErrInitDataMissing, got %v", err)
	}
}

func TestAuthenticateRejectsMissingInitData(t *testing.T) {
	web := &Web{debug: false}

	_, _, err := web.authenticate(testContext("/ws?roomId=love"))
	if !errors.Is(err, ErrInitDataMissing) {
		t.Fatalf("expected ErrInitDataMissing, got %v", err)
	}
}

// Local development has no bot token, so debug mode keeps the query fallback.
func TestAuthenticateAllowsQueryUserIdInDebug(t *testing.T) {
	web := &Web{debug: true}

	userID, locale, err := web.authenticate(testContext("/ws?userId=42&locale=ru"))
	if err != nil {
		t.Fatalf("debug mode should accept a query user id, got %v", err)
	}
	if userID != "42" {
		t.Fatalf("expected user id 42, got %q", userID)
	}
	if locale != "ru" {
		t.Fatalf("expected locale ru, got %q", locale)
	}
}

func TestAuthenticateRequiresUserIdInDebug(t *testing.T) {
	web := &Web{debug: true}

	_, _, err := web.authenticate(testContext("/ws?roomId=love"))
	if !errors.Is(err, ErrInitDataMissing) {
		t.Fatalf("expected ErrInitDataMissing, got %v", err)
	}
}
