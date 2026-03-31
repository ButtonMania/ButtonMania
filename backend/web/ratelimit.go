package web

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter tracks per-user connection and message rates.
type RateLimiter struct {
	mu              sync.Mutex
	connections     map[string]int
	maxConnPerUser  int
	maxMsgPerSec    int
	messageCounts   map[string]*rateBucket
	cleanupInterval time.Duration
}

type rateBucket struct {
	count    int
	lastTime time.Time
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(maxConnPerUser, maxMsgPerSec int) *RateLimiter {
	rl := &RateLimiter{
		connections:     make(map[string]int),
		maxConnPerUser:  maxConnPerUser,
		maxMsgPerSec:    maxMsgPerSec,
		messageCounts:   make(map[string]*rateBucket),
		cleanupInterval: 5 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

// AddConnection increments connection count for a user. Returns false if limit exceeded.
func (rl *RateLimiter) AddConnection(userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.connections[userID] >= rl.maxConnPerUser {
		slog.Warn("rate limit: max connections exceeded",
			"user_id", userID,
			"current", rl.connections[userID],
			"max", rl.maxConnPerUser,
		)
		return false
	}
	rl.connections[userID]++
	return true
}

// RemoveConnection decrements connection count for a user.
func (rl *RateLimiter) RemoveConnection(userID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.connections[userID] > 0 {
		rl.connections[userID]--
	}
	if rl.connections[userID] == 0 {
		delete(rl.connections, userID)
	}
}

// AllowMessage checks if a user can send a message based on rate limits.
func (rl *RateLimiter) AllowMessage(userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.messageCounts[userID]
	if !exists {
		rl.messageCounts[userID] = &rateBucket{count: 1, lastTime: now}
		return true
	}

	elapsed := now.Sub(bucket.lastTime)
	if elapsed >= time.Second {
		bucket.count = 1
		bucket.lastTime = now
		return true
	}

	bucket.count++
	if bucket.count > rl.maxMsgPerSec {
		slog.Warn("rate limit: message rate exceeded",
			"user_id", userID,
			"count", bucket.count,
			"max", rl.maxMsgPerSec,
		)
		return false
	}
	return true
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, v := range rl.messageCounts {
			if now.Sub(v.lastTime) > rl.cleanupInterval {
				delete(rl.messageCounts, k)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware returns a gin middleware that rate-limits requests by IP.
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	var mu sync.Mutex
	requests := make(map[string]*rateBucket)

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for k, v := range requests {
				if now.Sub(v.lastTime) > time.Minute {
					delete(requests, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		bucket, exists := requests[ip]
		if !exists {
			requests[ip] = &rateBucket{count: 1, lastTime: time.Now()}
			mu.Unlock()
			c.Next()
			return
		}

		elapsed := time.Since(bucket.lastTime)
		if elapsed >= time.Minute {
			bucket.count = 1
			bucket.lastTime = time.Now()
			mu.Unlock()
			c.Next()
			return
		}

		bucket.count++
		if bucket.count > requestsPerMinute {
			mu.Unlock()
			slog.Warn("rate limit: HTTP request rate exceeded",
				"ip", ip,
				"count", bucket.count,
			)
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		mu.Unlock()
		c.Next()
	}
}
