package web

import (
	"testing"
)

func TestMetrics_Counters(t *testing.T) {
	m := NewMetrics()

	m.ActiveSessions.Add(1)
	m.ActiveSessions.Add(1)
	m.TotalSessions.Add(1)
	m.WebSocketErrors.Add(1)
	m.HTTPRequests.Add(5)
	m.ActiveWSConns.Add(3)

	if m.ActiveSessions.Load() != 2 {
		t.Errorf("expected ActiveSessions 2, got %d", m.ActiveSessions.Load())
	}
	if m.TotalSessions.Load() != 1 {
		t.Errorf("expected TotalSessions 1, got %d", m.TotalSessions.Load())
	}
	if m.WebSocketErrors.Load() != 1 {
		t.Errorf("expected WebSocketErrors 1, got %d", m.WebSocketErrors.Load())
	}
	if m.HTTPRequests.Load() != 5 {
		t.Errorf("expected HTTPRequests 5, got %d", m.HTTPRequests.Load())
	}
	if m.ActiveWSConns.Load() != 3 {
		t.Errorf("expected ActiveWSConns 3, got %d", m.ActiveWSConns.Load())
	}
}

func TestMetrics_Decrement(t *testing.T) {
	m := NewMetrics()

	m.ActiveSessions.Add(5)
	m.ActiveSessions.Add(-2)

	if m.ActiveSessions.Load() != 3 {
		t.Errorf("expected ActiveSessions 3 after decrement, got %d", m.ActiveSessions.Load())
	}
}
