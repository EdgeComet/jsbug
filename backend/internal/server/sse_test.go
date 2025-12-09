package server

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewSSEManager(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)

	if manager == nil {
		t.Fatal("NewSSEManager() returned nil")
	}
	if manager.channels == nil {
		t.Error("channels map not initialized")
	}
}

func TestSSEManager_Subscribe(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)

	requestID := "test-123"
	ch := manager.Subscribe(requestID)

	if ch == nil {
		t.Fatal("Subscribe() returned nil channel")
	}

	if !manager.HasSubscriber(requestID) {
		t.Error("HasSubscriber() should return true after Subscribe()")
	}
}

func TestSSEManager_Unsubscribe(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)

	requestID := "test-123"
	manager.Subscribe(requestID)
	manager.Unsubscribe(requestID)

	if manager.HasSubscriber(requestID) {
		t.Error("HasSubscriber() should return false after Unsubscribe()")
	}
}

func TestSSEManager_Unsubscribe_NonExistent(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)

	// Should not panic
	manager.Unsubscribe("non-existent")
}

func TestSSEManager_Publish(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)

	requestID := "test-123"
	ch := manager.Subscribe(requestID)

	event := SSEEvent{
		Type: SSEEventStarted,
		Data: map[string]interface{}{"url": "https://example.com"},
	}

	manager.Publish(requestID, event)

	select {
	case received := <-ch:
		if received.Type != event.Type {
			t.Errorf("event type = %s, want %s", received.Type, event.Type)
		}
		if received.Data["url"] != "https://example.com" {
			t.Errorf("event data url = %v, want %s", received.Data["url"], "https://example.com")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected event not received")
	}
}

func TestSSEManager_Publish_NoSubscriber(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)

	// Should not panic
	manager.Publish("non-existent", SSEEvent{Type: SSEEventStarted})
}

func TestSSEManager_Publish_ChannelFull(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)

	requestID := "test-123"
	manager.Subscribe(requestID)

	// Fill the channel
	for i := 0; i < sseChannelBuffer+5; i++ {
		manager.Publish(requestID, SSEEvent{Type: SSEEventWaiting})
	}

	// Should not block or panic - extra events are dropped
}

func TestSSEManager_ResubscribeSameID(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)

	requestID := "test-123"

	ch1 := manager.Subscribe(requestID)
	ch2 := manager.Subscribe(requestID)

	// ch1 should be closed
	select {
	case _, ok := <-ch1:
		if ok {
			t.Error("old channel should be closed")
		}
	default:
		// Channel not closed yet, might happen due to timing
	}

	// ch2 should be active
	manager.Publish(requestID, SSEEvent{Type: SSEEventStarted})

	select {
	case <-ch2:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("new channel should receive events")
	}
}

func TestSSEManager_PublishHelpers(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)

	requestID := "test-123"
	ch := manager.Subscribe(requestID)

	tests := []struct {
		name     string
		publish  func()
		expected string
	}{
		{
			name: "PublishStarted",
			publish: func() {
				manager.PublishStarted(requestID, "https://example.com")
			},
			expected: SSEEventStarted,
		},
		{
			name: "PublishNavigating",
			publish: func() {
				manager.PublishNavigating(requestID, "https://example.com")
			},
			expected: SSEEventNavigating,
		},
		{
			name: "PublishWaiting",
			publish: func() {
				manager.PublishWaiting(requestID, "load", 500)
			},
			expected: SSEEventWaiting,
		},
		{
			name: "PublishCapturing",
			publish: func() {
				manager.PublishCapturing(requestID, 10)
			},
			expected: SSEEventCapturing,
		},
		{
			name: "PublishParsing",
			publish: func() {
				manager.PublishParsing(requestID)
			},
			expected: SSEEventParsing,
		},
		{
			name: "PublishComplete",
			publish: func() {
				manager.PublishComplete(requestID, 1500)
			},
			expected: SSEEventComplete,
		},
		{
			name: "PublishError",
			publish: func() {
				manager.PublishError(requestID, "RENDER_FAILED", "Test error")
			},
			expected: SSEEventError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.publish()

			select {
			case event := <-ch:
				if event.Type != tt.expected {
					t.Errorf("event type = %s, want %s", event.Type, tt.expected)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("expected event not received")
			}
		})
	}
}

func TestNewSSEHandler(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)
	handler := NewSSEHandler(manager, logger)

	if handler == nil {
		t.Fatal("NewSSEHandler() returned nil")
	}
}

func TestSSEHandler_MethodNotAllowed(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)
	handler := NewSSEHandler(manager, logger)

	req := httptest.NewRequest(http.MethodPost, "/api/render/stream?request_id=123", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestSSEHandler_MissingRequestID(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)
	handler := NewSSEHandler(manager, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/render/stream", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestSSEHandler_Headers(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)
	handler := NewSSEHandler(manager, logger)

	// Use a context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := httptest.NewRequest(http.MethodGet, "/api/render/stream?request_id=123", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %s, want text/event-stream", ct)
	}
	if cc := w.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("Cache-Control = %s, want no-cache", cc)
	}
	if conn := w.Header().Get("Connection"); conn != "keep-alive" {
		t.Errorf("Connection = %s, want keep-alive", conn)
	}
}

func TestSSEHandler_EventFormat(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)
	handler := NewSSEHandler(manager, logger)

	requestID := "test-456"

	// Create a server for testing
	server := httptest.NewServer(handler)
	defer server.Close()

	// Start a goroutine to send events
	go func() {
		time.Sleep(50 * time.Millisecond)
		manager.PublishStarted(requestID, "https://example.com")
		time.Sleep(50 * time.Millisecond)
		manager.PublishComplete(requestID, 1000)
	}()

	// Make request
	resp, err := http.Get(server.URL + "?request_id=" + requestID)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read events
	scanner := bufio.NewScanner(resp.Body)
	var events []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event:") {
			events = append(events, strings.TrimSpace(strings.TrimPrefix(line, "event:")))
		}
	}

	if len(events) < 2 {
		t.Errorf("expected at least 2 events, got %d", len(events))
	}

	if len(events) >= 1 && events[0] != SSEEventStarted {
		t.Errorf("first event = %s, want %s", events[0], SSEEventStarted)
	}

	if len(events) >= 2 && events[1] != SSEEventComplete {
		t.Errorf("second event = %s, want %s", events[1], SSEEventComplete)
	}
}

func TestSSEHandler_ClientDisconnect(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)
	handler := NewSSEHandler(manager, logger)

	requestID := "disconnect-test"

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/render/stream?request_id="+requestID, nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Run handler in goroutine
	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(w, req)
		close(done)
	}()

	// Wait for subscription
	time.Sleep(50 * time.Millisecond)

	// Cancel context (simulate disconnect)
	cancel()

	// Handler should exit
	select {
	case <-done:
		// Expected
	case <-time.After(500 * time.Millisecond):
		t.Error("handler did not exit after client disconnect")
	}

	// Subscription should be cleaned up
	time.Sleep(50 * time.Millisecond)
	if manager.HasSubscriber(requestID) {
		t.Error("subscription should be cleaned up after disconnect")
	}
}

func TestSSEHandler_CompleteEvent_ClosesConnection(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)
	handler := NewSSEHandler(manager, logger)

	requestID := "complete-test"

	req := httptest.NewRequest(http.MethodGet, "/api/render/stream?request_id="+requestID, nil)
	w := httptest.NewRecorder()

	// Run handler in goroutine
	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(w, req)
		close(done)
	}()

	// Wait for subscription
	time.Sleep(50 * time.Millisecond)

	// Send complete event
	manager.PublishComplete(requestID, 1000)

	// Handler should exit
	select {
	case <-done:
		// Expected
	case <-time.After(500 * time.Millisecond):
		t.Error("handler did not exit after complete event")
	}
}

func TestSSEHandler_ErrorEvent_ClosesConnection(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSSEManager(logger)
	handler := NewSSEHandler(manager, logger)

	requestID := "error-test"

	req := httptest.NewRequest(http.MethodGet, "/api/render/stream?request_id="+requestID, nil)
	w := httptest.NewRecorder()

	// Run handler in goroutine
	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(w, req)
		close(done)
	}()

	// Wait for subscription
	time.Sleep(50 * time.Millisecond)

	// Send error event
	manager.PublishError(requestID, "TEST_ERROR", "Test error message")

	// Handler should exit
	select {
	case <-done:
		// Expected
	case <-time.After(500 * time.Millisecond):
		t.Error("handler did not exit after error event")
	}
}
