package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

// SSE event types
const (
	SSEEventStarted    = "started"
	SSEEventNavigating = "navigating"
	SSEEventWaiting    = "waiting"
	SSEEventCapturing  = "capturing"
	SSEEventParsing    = "parsing"
	SSEEventComplete   = "complete"
	SSEEventError      = "error"
)

// SSE channel buffer size
const sseChannelBuffer = 16

// SSEEvent represents a server-sent event
type SSEEvent struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// SSEManager manages SSE subscriptions
type SSEManager struct {
	channels map[string]chan SSEEvent
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewSSEManager creates a new SSEManager
func NewSSEManager(logger *zap.Logger) *SSEManager {
	return &SSEManager{
		channels: make(map[string]chan SSEEvent),
		logger:   logger,
	}
}

// Subscribe creates a subscription for the given request ID
func (m *SSEManager) Subscribe(requestID string) <-chan SSEEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close existing channel if present
	if ch, exists := m.channels[requestID]; exists {
		close(ch)
	}

	ch := make(chan SSEEvent, sseChannelBuffer)
	m.channels[requestID] = ch

	m.logger.Debug("SSE subscription created", zap.String("request_id", requestID))

	return ch
}

// Unsubscribe removes a subscription for the given request ID
func (m *SSEManager) Unsubscribe(requestID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ch, exists := m.channels[requestID]; exists {
		close(ch)
		delete(m.channels, requestID)
		m.logger.Debug("SSE subscription removed", zap.String("request_id", requestID))
	}
}

// Publish sends an event to subscribers for the given request ID
func (m *SSEManager) Publish(requestID string, event SSEEvent) {
	m.mu.RLock()
	ch, exists := m.channels[requestID]
	m.mu.RUnlock()

	if !exists {
		return
	}

	// Non-blocking send
	select {
	case ch <- event:
		m.logger.Debug("SSE event published",
			zap.String("request_id", requestID),
			zap.String("event_type", event.Type),
		)
	default:
		m.logger.Warn("SSE channel full, dropping event",
			zap.String("request_id", requestID),
			zap.String("event_type", event.Type),
		)
	}
}

// HasSubscriber checks if there's an active subscriber for the request ID
func (m *SSEManager) HasSubscriber(requestID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.channels[requestID]
	return exists
}

// SSEHandler handles SSE connections
type SSEHandler struct {
	manager *SSEManager
	logger  *zap.Logger
}

// NewSSEHandler creates a new SSEHandler
func NewSSEHandler(manager *SSEManager, logger *zap.Logger) *SSEHandler {
	return &SSEHandler{
		manager: manager,
		logger:  logger,
	}
}

// ServeHTTP handles GET /api/render/stream requests
func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestID := r.URL.Query().Get("request_id")
	if requestID == "" {
		http.Error(w, "request_id query parameter is required", http.StatusBadRequest)
		return
	}

	// Check if response writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Subscribe to events
	events := h.manager.Subscribe(requestID)
	defer h.manager.Unsubscribe(requestID)

	h.logger.Debug("SSE connection established", zap.String("request_id", requestID))

	// Stream events
	for {
		select {
		case <-r.Context().Done():
			h.logger.Debug("SSE client disconnected", zap.String("request_id", requestID))
			return

		case event, ok := <-events:
			if !ok {
				// Channel closed
				return
			}

			if err := h.writeEvent(w, event); err != nil {
				h.logger.Error("Failed to write SSE event",
					zap.String("request_id", requestID),
					zap.Error(err),
				)
				return
			}
			flusher.Flush()

			// Close connection on complete or error
			if event.Type == SSEEventComplete || event.Type == SSEEventError {
				return
			}
		}
	}
}

// writeEvent writes an SSE event to the response writer
func (h *SSEHandler) writeEvent(w http.ResponseWriter, event SSEEvent) error {
	// Format: event: {type}\ndata: {json}\n\n
	data, err := json.Marshal(event.Data)
	if err != nil {
		data = []byte("{}")
	}

	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(data))
	return err
}

// PublishStarted publishes a started event
func (m *SSEManager) PublishStarted(requestID, url string) {
	m.Publish(requestID, SSEEvent{
		Type: SSEEventStarted,
		Data: map[string]interface{}{
			"url": url,
		},
	})
}

// PublishNavigating publishes a navigating event
func (m *SSEManager) PublishNavigating(requestID, url string) {
	m.Publish(requestID, SSEEvent{
		Type: SSEEventNavigating,
		Data: map[string]interface{}{
			"url": url,
		},
	})
}

// PublishWaiting publishes a waiting event
func (m *SSEManager) PublishWaiting(requestID, waitEvent string, elapsedMS int64) {
	m.Publish(requestID, SSEEvent{
		Type: SSEEventWaiting,
		Data: map[string]interface{}{
			"wait_event": waitEvent,
			"elapsed_ms": elapsedMS,
		},
	})
}

// PublishCapturing publishes a capturing event
func (m *SSEManager) PublishCapturing(requestID string, requestCount int) {
	m.Publish(requestID, SSEEvent{
		Type: SSEEventCapturing,
		Data: map[string]interface{}{
			"request_count": requestCount,
		},
	})
}

// PublishParsing publishes a parsing event
func (m *SSEManager) PublishParsing(requestID string) {
	m.Publish(requestID, SSEEvent{
		Type: SSEEventParsing,
	})
}

// PublishComplete publishes a complete event
func (m *SSEManager) PublishComplete(requestID string, renderTime float64) {
	m.Publish(requestID, SSEEvent{
		Type: SSEEventComplete,
		Data: map[string]interface{}{
			"render_time": renderTime,
		},
	})
}

// PublishError publishes an error event
func (m *SSEManager) PublishError(requestID, code, message string) {
	m.Publish(requestID, SSEEvent{
		Type: SSEEventError,
		Data: map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}
