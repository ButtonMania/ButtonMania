package web

import (
	"testing"
)

func TestRateLimiter_AddConnection(t *testing.T) {
	rl := NewRateLimiter(2, 10)

	if !rl.AddConnection("user1") {
		t.Error("first connection should be allowed")
	}
	if !rl.AddConnection("user1") {
		t.Error("second connection should be allowed")
	}
	if rl.AddConnection("user1") {
		t.Error("third connection should be denied")
	}

	// Different user should be allowed
	if !rl.AddConnection("user2") {
		t.Error("different user should be allowed")
	}
}

func TestRateLimiter_RemoveConnection(t *testing.T) {
	rl := NewRateLimiter(1, 10)

	rl.AddConnection("user1")
	rl.RemoveConnection("user1")

	if !rl.AddConnection("user1") {
		t.Error("should allow connection after removal")
	}
}

func TestRateLimiter_AllowMessage(t *testing.T) {
	rl := NewRateLimiter(10, 3)

	for i := 0; i < 3; i++ {
		if !rl.AllowMessage("user1") {
			t.Errorf("message %d should be allowed", i+1)
		}
	}

	if rl.AllowMessage("user1") {
		t.Error("fourth message in same second should be denied")
	}
}

func TestRateLimiter_RemoveNonexistent(t *testing.T) {
	rl := NewRateLimiter(10, 10)
	// Should not panic
	rl.RemoveConnection("nonexistent")
}
