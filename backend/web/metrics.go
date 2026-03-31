package web

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics tracks application-level metrics in a lightweight way.
type Metrics struct {
	ActiveSessions  atomic.Int64
	TotalSessions   atomic.Int64
	WebSocketErrors atomic.Int64
	HTTPRequests    atomic.Int64
	ActiveWSConns   atomic.Int64
	startTime       time.Time
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

// MetricsHandler returns a gin handler that exposes Prometheus-compatible metrics.
func (m *Metrics) MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		uptime := time.Since(m.startTime).Seconds()
		activeSessions := m.ActiveSessions.Load()
		totalSessions := m.TotalSessions.Load()
		wsErrors := m.WebSocketErrors.Load()
		httpRequests := m.HTTPRequests.Load()
		activeWSConns := m.ActiveWSConns.Load()

		body := fmt.Sprintf(`# HELP buttonmania_active_sessions Number of currently active game sessions
# TYPE buttonmania_active_sessions gauge
buttonmania_active_sessions %d
# HELP buttonmania_total_sessions Total number of game sessions started
# TYPE buttonmania_total_sessions counter
buttonmania_total_sessions %d
# HELP buttonmania_websocket_errors Total number of WebSocket errors
# TYPE buttonmania_websocket_errors counter
buttonmania_websocket_errors %d
# HELP buttonmania_http_requests_total Total number of HTTP requests
# TYPE buttonmania_http_requests_total counter
buttonmania_http_requests_total %d
# HELP buttonmania_active_ws_connections Number of active WebSocket connections
# TYPE buttonmania_active_ws_connections gauge
buttonmania_active_ws_connections %d
# HELP buttonmania_uptime_seconds Server uptime in seconds
# TYPE buttonmania_uptime_seconds gauge
buttonmania_uptime_seconds %.2f
`, activeSessions, totalSessions, wsErrors, httpRequests, activeWSConns, uptime)

		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(body))
	}
}

// MetricsMiddleware returns a middleware that counts HTTP requests.
func (m *Metrics) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.HTTPRequests.Add(1)
		c.Next()
	}
}
