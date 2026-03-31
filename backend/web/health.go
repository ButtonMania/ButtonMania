package web

import (
	"net/http"

	"buttonmania.win/db"
	"github.com/gin-gonic/gin"
)

// HealthStatus represents the health check response.
type HealthStatus struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

// HealthHandler returns a gin handler for health checks.
func HealthHandler(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := HealthStatus{
			Status:   "ok",
			Services: make(map[string]string),
		}
		httpStatus := http.StatusOK

		// Check Redis
		if err := database.PingRedis(); err != nil {
			status.Services["redis"] = "unhealthy: " + err.Error()
			status.Status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		} else {
			status.Services["redis"] = "healthy"
		}

		// Check Postgres
		if err := database.PingPostgres(); err != nil {
			status.Services["postgres"] = "unhealthy: " + err.Error()
			status.Status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		} else {
			status.Services["postgres"] = "healthy"
		}

		c.JSON(httpStatus, status)
	}
}
