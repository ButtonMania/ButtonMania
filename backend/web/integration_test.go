//go:build integration

package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"buttonmania.win/bot"
	"buttonmania.win/conf"
	"buttonmania.win/db"
	"buttonmania.win/protocol"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	initdata "github.com/Telegram-Web-Apps/init-data-golang"
)

const (
	testToken     = "123456:fake-local-integration-test-token"
	testClientId  = protocol.ClientID("buttonmania")
	testRoomId    = protocol.RoomID("love")
	testUserId    = int64(279058397)
	testStranger  = int64(111111111)
	testCustomRid = protocol.RoomID("customroom")
)

// signedInitData builds init data signed with the test bot token.
func signedInitData(t *testing.T, userId int64) string {
	t.Helper()
	user := fmt.Sprintf(`{"id":%d,"first_name":"Test","language_code":"en"}`, userId)
	payload := url.Values{"user": {user}}.Encode()
	authDate := time.Now()

	hash, err := initdata.SignQueryString(payload, testToken, authDate)
	if err != nil {
		t.Fatalf("failed to sign init data: %v", err)
	}
	return fmt.Sprintf("%s&auth_date=%d&hash=%s", payload, authDate.Unix(), hash)
}

// newTestServer starts a server backed by the integration databases.
func newTestServer(t *testing.T) (*httptest.Server, *Web) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	database, err := db.NewDB(db.IntegrationContext())
	if err != nil {
		t.Fatalf("failed to connect to test databases: %v", err)
	}
	if err := database.FlushForTests(); err != nil {
		t.Fatalf("failed to flush test databases: %v", err)
	}

	ctx := db.IntegrationContext()
	ctx = context.WithValue(ctx, KeySessionName, "session")
	ctx = context.WithValue(ctx, KeySessionSecret, "secret")
	ctx = context.WithValue(ctx, KeyStaticPath, t.TempDir())
	ctx = context.WithValue(ctx, KeyServerPort, 0)
	ctx = context.WithValue(ctx, KeyServerTLSCert, "")
	ctx = context.WithValue(ctx, KeyServerTLSKey, "")
	ctx = context.WithValue(ctx, KeyAllowedOrigins, "*")
	ctx = context.WithValue(ctx, bot.KeyTelegramToken, testToken)

	configuration := conf.Conf{
		Clients: []conf.ClientConf{
			{ClientId: testClientId, Rooms: []protocol.RoomID{testRoomId}},
		},
	}

	engine := gin.New()
	web, err := NewWeb(ctx, configuration, engine, database, false)
	if err != nil {
		t.Fatalf("failed to create web: %v", err)
	}
	web.RegisterRoutes()

	server := httptest.NewServer(engine)
	t.Cleanup(func() {
		server.Close()
		database.Close()
	})
	return server, web
}

func get(t *testing.T, server *httptest.Server, path string) (int, string) {
	t.Helper()
	response, err := http.Get(server.URL + path)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer response.Body.Close()

	body := make([]byte, 4096)
	n, _ := response.Body.Read(body)
	return response.StatusCode, string(body[:n])
}

// dialWs opens a websocket connection to the given room.
func dialWs(t *testing.T, server *httptest.Server, room protocol.RoomID, initData string) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	endpoint := "ws" + strings.TrimPrefix(server.URL, "http")
	target := fmt.Sprintf(
		"%s/ws?clientId=%s&roomId=%s&initData=%s",
		endpoint,
		testClientId,
		room,
		url.QueryEscape(initData),
	)
	return websocket.DefaultDialer.Dial(target, nil)
}

func readMessage(t *testing.T, ws *websocket.Conn) protocol.GameplayMessage {
	t.Helper()
	var msg protocol.GameplayMessage
	if err := ws.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("failed to set deadline: %v", err)
	}
	if err := ws.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	return msg
}

func TestWsRequiresInitData(t *testing.T) {
	server, _ := newTestServer(t)

	endpoint := "ws" + strings.TrimPrefix(server.URL, "http")
	target := fmt.Sprintf("%s/ws?clientId=%s&roomId=%s&userId=%d", endpoint, testClientId, testRoomId, testUserId)
	_, response, err := websocket.DefaultDialer.Dial(target, nil)
	if err == nil {
		t.Fatal("connection without init data was accepted")
	}
	if response == nil || response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %v", response)
	}
}

func TestWsRejectsInvalidInitData(t *testing.T) {
	server, _ := newTestServer(t)

	_, response, err := dialWs(t, server, testRoomId, "user=%7B%22id%22%3A1%7D&auth_date=1&hash=deadbeef")
	if err == nil {
		t.Fatal("connection with a forged signature was accepted")
	}
	if response == nil || response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %v", response)
	}
}

func TestWsGameplayRoundTrip(t *testing.T) {
	server, _ := newTestServer(t)

	ws, _, err := dialWs(t, server, testRoomId, signedInitData(t, testUserId))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	start := readMessage(t, ws)
	if start.GameState != protocol.Update || start.Context == nil {
		t.Fatalf("expected an update message, got %+v", start)
	}
	if start.Context.ButtonPhase != protocol.Push {
		t.Fatalf("expected push phase, got %q", start.Context.ButtonPhase)
	}

	hold := protocol.GameplayContext{ButtonPhase: protocol.Hold}
	if err := ws.WriteJSON(hold); err != nil {
		t.Fatalf("failed to send hold: %v", err)
	}
	update := readMessage(t, ws)
	if update.GameState != protocol.Update {
		t.Fatalf("expected an update message, got %+v", update)
	}
	if update.CountActive == nil || *update.CountActive != 1 {
		t.Fatalf("expected one active session, got %+v", update.CountActive)
	}

	release := protocol.GameplayContext{ButtonPhase: protocol.Release}
	if err := ws.WriteJSON(release); err != nil {
		t.Fatalf("failed to send release: %v", err)
	}
	record := readMessage(t, ws)
	if record.GameState != protocol.Record || record.Record == nil {
		t.Fatalf("expected a record message, got %+v", record)
	}
}

func TestWsRejectsSecondSessionOfSameUser(t *testing.T) {
	server, _ := newTestServer(t)
	initData := signedInitData(t, testUserId)

	first, _, err := dialWs(t, server, testRoomId, initData)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer first.Close()
	readMessage(t, first)

	second, _, err := dialWs(t, server, testRoomId, initData)
	if err != nil {
		t.Fatalf("failed to open the second connection: %v", err)
	}
	defer second.Close()

	msg := readMessage(t, second)
	if msg.GameState != protocol.Error || msg.Error == nil {
		t.Fatalf("expected an error message, got %+v", msg)
	}

	hold := protocol.GameplayContext{ButtonPhase: protocol.Hold}
	if err := first.WriteJSON(hold); err != nil {
		t.Fatalf("failed to send hold on the first connection: %v", err)
	}
	if update := readMessage(t, first); update.GameState != protocol.Update {
		t.Fatalf("the running session was disturbed, got %+v", update)
	}
}

// Oversized frames must be rejected, the read limit protects server memory.
func TestWsEnforcesReadLimit(t *testing.T) {
	server, _ := newTestServer(t)

	ws, _, err := dialWs(t, server, testRoomId, signedInitData(t, testUserId))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()
	readMessage(t, ws)

	oversized := protocol.GameplayContext{
		ButtonPhase: protocol.Hold,
		ChatMessage: &protocol.ChatMessage{Message: strings.Repeat("a", maxWsMessageBytes*2)},
	}
	if err := ws.WriteJSON(oversized); err != nil {
		t.Fatalf("failed to send oversized message: %v", err)
	}

	if err := ws.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("failed to set deadline: %v", err)
	}
	var msg protocol.GameplayMessage
	if err := ws.ReadJSON(&msg); err == nil {
		t.Fatal("oversized frame was accepted")
	}
}

func TestCustomRoomHandlers(t *testing.T) {
	server, _ := newTestServer(t)
	owner := url.QueryEscape(signedInitData(t, testUserId))
	stranger := url.QueryEscape(signedInitData(t, testStranger))

	status, _ := get(t, server, fmt.Sprintf("/api/room/create?clientId=%s&roomId=%s", testClientId, testCustomRid))
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401 without init data, got %d", status)
	}

	status, body := get(t, server, fmt.Sprintf("/api/room/create?clientId=%s&roomId=%s&initData=%s", testClientId, testCustomRid, owner))
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}

	status, _ = get(t, server, fmt.Sprintf("/api/room/create?clientId=%s&roomId=%s&initData=%s", testClientId, testCustomRid, owner))
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 for a duplicate room, got %d", status)
	}

	status, _ = get(t, server, fmt.Sprintf("/api/room/delete?clientId=%s&roomId=%s&initData=%s", testClientId, testCustomRid, stranger))
	if status != http.StatusForbidden {
		t.Fatalf("expected 403 for a stranger, got %d", status)
	}

	status, _ = get(t, server, fmt.Sprintf("/api/room/stats?clientId=%s&roomId=%s", testClientId, testCustomRid))
	if status != http.StatusOK {
		t.Fatalf("the room was removed by a stranger, stats returned %d", status)
	}

	status, _ = get(t, server, fmt.Sprintf("/api/room/delete?clientId=%s&roomId=%s&initData=%s", testClientId, testCustomRid, owner))
	if status != http.StatusOK {
		t.Fatalf("expected 200 for the owner, got %d", status)
	}

	status, _ = get(t, server, fmt.Sprintf("/api/room/stats?clientId=%s&roomId=%s", testClientId, testCustomRid))
	if status != http.StatusNotFound {
		t.Fatalf("expected 404 for a deleted room, got %d", status)
	}
}

func TestPredefinedRoomCannotBeDeleted(t *testing.T) {
	server, _ := newTestServer(t)
	owner := url.QueryEscape(signedInitData(t, testUserId))

	status, _ := get(t, server, fmt.Sprintf("/api/room/delete?clientId=%s&roomId=%s&initData=%s", testClientId, testRoomId, owner))
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", status)
	}
}

func TestStatsHandlers(t *testing.T) {
	server, _ := newTestServer(t)

	status, body := get(t, server, fmt.Sprintf("/api/room/stats?clientId=%s&roomId=%s", testClientId, testRoomId))
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}
	var roomStats protocol.GameRoomStats
	if err := json.Unmarshal([]byte(body), &roomStats); err != nil {
		t.Fatalf("failed to decode room stats: %v", err)
	}

	status, body = get(t, server, fmt.Sprintf("/api/stats?clientId=%s", testClientId))
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}
	var clientStats protocol.ClientStats
	if err := json.Unmarshal([]byte(body), &clientStats); err != nil {
		t.Fatalf("failed to decode client stats: %v", err)
	}
	if clientStats.UsersOnline == nil || clientStats.RoomsCount == nil {
		t.Fatalf("expected populated client stats, got %+v", clientStats)
	}

	status, _ = get(t, server, "/api/stats?clientId=unknown")
	if status != http.StatusNotFound {
		t.Fatalf("expected 404 for an unknown client, got %d", status)
	}
}
