package chrome

import (
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewEventCollector(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	if ec == nil {
		t.Fatal("NewEventCollector() returned nil")
	}
	if ec.networkRequests == nil {
		t.Error("networkRequests map is nil")
	}
	if ec.consoleMessages == nil {
		t.Error("consoleMessages slice is nil")
	}
	if ec.jsErrors == nil {
		t.Error("jsErrors slice is nil")
	}
	if ec.lifecycleEvents == nil {
		t.Error("lifecycleEvents map is nil")
	}
}

func TestEventCollector_NetworkRequestLifecycle(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	// Simulate request start
	ec.mu.Lock()
	ec.networkRequests["req-1"] = &NetworkRequestData{
		RequestID:    "req-1",
		URL:          "https://example.com/api",
		Method:       "GET",
		ResourceType: "XHR",
		StartTime:    time.Now(),
	}
	ec.mu.Unlock()

	// Simulate response
	ec.mu.Lock()
	if req, ok := ec.networkRequests["req-1"]; ok {
		req.Status = 200
	}
	ec.mu.Unlock()

	// Simulate loading finished
	ec.mu.Lock()
	if req, ok := ec.networkRequests["req-1"]; ok {
		req.EndTime = time.Now()
		req.SizeBytes = 1024
	}
	ec.mu.Unlock()

	// Get results
	requests := ec.GetNetworkResults()

	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	req := requests[0]
	if req.URL != "https://example.com/api" {
		t.Errorf("URL = %q, want %q", req.URL, "https://example.com/api")
	}
	if req.Method != "GET" {
		t.Errorf("Method = %q, want %q", req.Method, "GET")
	}
	if req.Status != 200 {
		t.Errorf("Status = %d, want %d", req.Status, 200)
	}
	if req.Size != 1024 {
		t.Errorf("Size = %d, want %d", req.Size, 1024)
	}
}

func TestEventCollector_FailedRequest(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	ec.mu.Lock()
	ec.networkRequests["req-1"] = &NetworkRequestData{
		RequestID:     "req-1",
		URL:           "https://example.com/fail",
		Method:        "GET",
		ResourceType:  "Document",
		StartTime:     time.Now(),
		Failed:        true,
		FailureReason: "net::ERR_CONNECTION_REFUSED",
		EndTime:       time.Now(),
	}
	ec.mu.Unlock()

	requests := ec.GetNetworkResults()

	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	if !requests[0].Failed {
		t.Error("expected request to be marked as failed")
	}
}

func TestEventCollector_BlockedRequest(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	ec.mu.Lock()
	ec.networkRequests["req-1"] = &NetworkRequestData{
		RequestID:    "req-1",
		URL:          "https://google-analytics.com/collect",
		ResourceType: "Script",
		StartTime:    time.Now(),
		Blocked:      true,
		EndTime:      time.Now(),
	}
	ec.mu.Unlock()

	requests := ec.GetNetworkResults()

	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	if !requests[0].Blocked {
		t.Error("expected request to be marked as blocked")
	}
}

func TestEventCollector_ConsoleMessages(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	ec.mu.Lock()
	ec.consoleMessages = []ConsoleMessageData{
		{Level: "log", Message: "Hello world", Time: ec.startTime.Add(100 * time.Millisecond)},
		{Level: "warning", Message: "Deprecated API", Time: ec.startTime.Add(200 * time.Millisecond)},
		{Level: "error", Message: "Something failed", Time: ec.startTime.Add(300 * time.Millisecond)},
		{Level: "info", Message: "Info message", Time: ec.startTime.Add(400 * time.Millisecond)},
		{Level: "debug", Message: "Debug message", Time: ec.startTime.Add(500 * time.Millisecond)},
	}
	ec.mu.Unlock()

	messages := ec.GetConsoleResults()

	if len(messages) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(messages))
	}

	// Check timing (Timestamp is in seconds)
	if messages[0].Time < 0.1 || messages[0].Time > 0.15 {
		t.Errorf("first message Time = %f, expected around 0.1 seconds", messages[0].Time)
	}
}

func TestEventCollector_JSErrors(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	ec.mu.Lock()
	ec.jsErrors = []JSErrorData{
		{
			Message:    "TypeError: undefined is not a function",
			Source:     "https://example.com/app.js",
			Line:       42,
			Column:     10,
			StackTrace: "at foo (app.js:42:10)\nat bar (app.js:100:5)",
			Time:       ec.startTime.Add(500 * time.Millisecond),
		},
	}
	ec.mu.Unlock()

	errors := ec.GetJSErrors()

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	err := errors[0]
	if err.Message != "TypeError: undefined is not a function" {
		t.Errorf("Message = %q", err.Message)
	}
	if err.Source != "https://example.com/app.js" {
		t.Errorf("Source = %q", err.Source)
	}
	if err.Line != 42 {
		t.Errorf("Line = %d, want 42", err.Line)
	}
	if err.Column != 10 {
		t.Errorf("Column = %d, want 10", err.Column)
	}
	if err.StackTrace == "" {
		t.Error("StackTrace is empty")
	}
	if err.Timestamp < 0.5 || err.Timestamp > 0.55 {
		t.Errorf("Timestamp = %f, expected around 0.5 seconds", err.Timestamp)
	}
}

func TestEventCollector_LifecycleEvents(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	ec.mu.Lock()
	ec.lifecycleEvents["DOMContentLoaded"] = ec.startTime.Add(200 * time.Millisecond)
	ec.lifecycleEvents["load"] = ec.startTime.Add(500 * time.Millisecond)
	ec.lifecycleEvents["networkIdle"] = ec.startTime.Add(800 * time.Millisecond)
	ec.mu.Unlock()

	lifecycle := ec.GetLifecycleResults()

	if len(lifecycle) != 3 {
		t.Fatalf("expected 3 lifecycle events, got %d", len(lifecycle))
	}

	// Check DOMContentLoaded (first event)
	if lifecycle[0].Event != "DOMContentLoaded" {
		t.Errorf("expected first event to be DOMContentLoaded, got %s", lifecycle[0].Event)
	}
	if lifecycle[0].Time < 0.2 || lifecycle[0].Time > 0.25 {
		t.Errorf("DOMContentLoaded time = %f, expected around 0.2 seconds", lifecycle[0].Time)
	}

	// Check load (second event)
	if lifecycle[1].Event != "load" {
		t.Errorf("expected second event to be load, got %s", lifecycle[1].Event)
	}
	if lifecycle[1].Time < 0.5 || lifecycle[1].Time > 0.55 {
		t.Errorf("load time = %f, expected around 0.5 seconds", lifecycle[1].Time)
	}

	// Check networkIdle (third event)
	if lifecycle[2].Event != "networkIdle" {
		t.Errorf("expected third event to be networkIdle, got %s", lifecycle[2].Event)
	}
	if lifecycle[2].Time < 0.8 || lifecycle[2].Time > 0.85 {
		t.Errorf("networkIdle time = %f, expected around 0.8 seconds", lifecycle[2].Time)
	}
}

func TestEventCollector_ActiveRequestCount(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	ec.mu.Lock()
	// Active request (no EndTime)
	ec.networkRequests["req-1"] = &NetworkRequestData{
		RequestID: "req-1",
		URL:       "https://example.com/1",
		StartTime: time.Now(),
	}
	// Completed request
	ec.networkRequests["req-2"] = &NetworkRequestData{
		RequestID: "req-2",
		URL:       "https://example.com/2",
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}
	// Blocked request
	ec.networkRequests["req-3"] = &NetworkRequestData{
		RequestID: "req-3",
		URL:       "https://analytics.com",
		StartTime: time.Now(),
		Blocked:   true,
	}
	// Failed request
	ec.networkRequests["req-4"] = &NetworkRequestData{
		RequestID: "req-4",
		URL:       "https://failed.com",
		StartTime: time.Now(),
		Failed:    true,
	}
	// Another active request
	ec.networkRequests["req-5"] = &NetworkRequestData{
		RequestID: "req-5",
		URL:       "https://example.com/5",
		StartTime: time.Now(),
	}
	ec.mu.Unlock()

	count := ec.ActiveRequestCount()
	if count != 2 {
		t.Errorf("ActiveRequestCount() = %d, want 2", count)
	}
}

func TestEventCollector_ThreadSafety(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ec.mu.Lock()
			ec.networkRequests[string(rune('0'+id%10))] = &NetworkRequestData{
				RequestID: string(rune('0' + id%10)),
				URL:       "https://example.com",
				StartTime: time.Now(),
			}
			ec.consoleMessages = append(ec.consoleMessages, ConsoleMessageData{
				Level:   "log",
				Message: "test",
				Time:    time.Now(),
			})
			ec.mu.Unlock()
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ec.GetNetworkResults()
			ec.GetConsoleResults()
			ec.GetJSErrors()
			ec.GetLifecycleResults()
			ec.ActiveRequestCount()
		}()
	}

	wg.Wait()

	// Verify no panic occurred and data is accessible
	requests := ec.GetNetworkResults()
	if len(requests) == 0 {
		t.Error("expected some requests after concurrent access")
	}
}

func TestEventCollector_EmptyResults(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	requests := ec.GetNetworkResults()
	if len(requests) != 0 {
		t.Errorf("expected 0 requests, got %d", len(requests))
	}

	messages := ec.GetConsoleResults()
	if len(messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(messages))
	}

	errors := ec.GetJSErrors()
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}

	lifecycle := ec.GetLifecycleResults()
	if len(lifecycle) != 0 {
		t.Errorf("expected 0 lifecycle events, got %d", len(lifecycle))
	}
}

func TestEventCollector_NetworkRequestID(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	ec.mu.Lock()
	ec.networkRequests["req-123"] = &NetworkRequestData{
		RequestID:    "req-123",
		URL:          "https://example.com/api",
		Method:       "GET",
		ResourceType: "XHR",
		Status:       200,
		StartTime:    time.Now(),
		EndTime:      time.Now(),
	}
	ec.mu.Unlock()

	requests := ec.GetNetworkResults()

	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	// Verify ID is captured from RequestID
	if requests[0].ID != "req-123" {
		t.Errorf("ID = %q, want %q", requests[0].ID, "req-123")
	}
}

func TestEventCollector_NetworkRequestIsInternal(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	// Set page URL for internal detection
	ec.SetPageURL("https://example.com/page")

	ec.mu.Lock()
	// Internal request (same domain)
	ec.networkRequests["req-1"] = &NetworkRequestData{
		RequestID:    "req-1",
		URL:          "https://example.com/api",
		Method:       "GET",
		ResourceType: "XHR",
		StartTime:    time.Now(),
		EndTime:      time.Now(),
	}
	// Internal request (subdomain)
	ec.networkRequests["req-2"] = &NetworkRequestData{
		RequestID:    "req-2",
		URL:          "https://cdn.example.com/file.js",
		Method:       "GET",
		ResourceType: "Script",
		StartTime:    time.Now(),
		EndTime:      time.Now(),
	}
	// External request
	ec.networkRequests["req-3"] = &NetworkRequestData{
		RequestID:    "req-3",
		URL:          "https://other.com/api",
		Method:       "GET",
		ResourceType: "XHR",
		StartTime:    time.Now(),
		EndTime:      time.Now(),
	}
	// Third-party script
	ec.networkRequests["req-4"] = &NetworkRequestData{
		RequestID:    "req-4",
		URL:          "https://analytics.google.com/collect",
		Method:       "POST",
		ResourceType: "XHR",
		StartTime:    time.Now(),
		EndTime:      time.Now(),
	}
	ec.mu.Unlock()

	requests := ec.GetNetworkResults()

	// Find requests by ID
	requestMap := make(map[string]bool)
	for _, req := range requests {
		requestMap[req.ID] = req.IsInternal
	}

	// Verify internal detection
	if !requestMap["req-1"] {
		t.Error("Same domain request should be internal")
	}
	if !requestMap["req-2"] {
		t.Error("Subdomain request should be internal")
	}
	if requestMap["req-3"] {
		t.Error("Different domain request should be external")
	}
	if requestMap["req-4"] {
		t.Error("Third-party request should be external")
	}
}

func TestEventCollector_ConsoleMessageID(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	ec.mu.Lock()
	ec.consoleMessages = []ConsoleMessageData{
		{Level: "log", Message: "First", Time: ec.startTime.Add(100 * time.Millisecond)},
		{Level: "warn", Message: "Second", Time: ec.startTime.Add(200 * time.Millisecond)},
		{Level: "error", Message: "Third", Time: ec.startTime.Add(300 * time.Millisecond)},
	}
	ec.mu.Unlock()

	messages := ec.GetConsoleResults()

	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}

	// Verify sequential IDs
	if messages[0].ID != "console-1" {
		t.Errorf("messages[0].ID = %q, want %q", messages[0].ID, "console-1")
	}
	if messages[1].ID != "console-2" {
		t.Errorf("messages[1].ID = %q, want %q", messages[1].ID, "console-2")
	}
	if messages[2].ID != "console-3" {
		t.Errorf("messages[2].ID = %q, want %q", messages[2].ID, "console-3")
	}
}

func TestEventCollector_SetPageURL(t *testing.T) {
	logger := zap.NewNop()
	ec := NewEventCollector(logger)

	// Initially empty
	if ec.pageURL != "" {
		t.Errorf("pageURL should be empty initially, got %q", ec.pageURL)
	}

	// Set page URL
	ec.SetPageURL("https://example.com/page")

	ec.mu.RLock()
	if ec.pageURL != "https://example.com/page" {
		t.Errorf("pageURL = %q, want %q", ec.pageURL, "https://example.com/page")
	}
	ec.mu.RUnlock()
}

func TestUrlsMatchIgnoringFragment(t *testing.T) {
	tests := []struct {
		name     string
		url1     string
		url2     string
		expected bool
	}{
		{
			name:     "exact match",
			url1:     "https://example.com/page",
			url2:     "https://example.com/page",
			expected: true,
		},
		{
			name:     "trailing slash on root - url1 has slash",
			url1:     "https://example.com/",
			url2:     "https://example.com",
			expected: true,
		},
		{
			name:     "trailing slash on root - url2 has slash",
			url1:     "https://example.com",
			url2:     "https://example.com/",
			expected: true,
		},
		{
			name:     "both have trailing slash on root",
			url1:     "https://example.com/",
			url2:     "https://example.com/",
			expected: true,
		},
		{
			name:     "different paths",
			url1:     "https://example.com/page1",
			url2:     "https://example.com/page2",
			expected: false,
		},
		{
			name:     "same path different hosts",
			url1:     "https://example.com/page",
			url2:     "https://other.com/page",
			expected: false,
		},
		{
			name:     "with fragment - should match",
			url1:     "https://example.com/page#section",
			url2:     "https://example.com/page",
			expected: true,
		},
		{
			name:     "different schemes",
			url1:     "https://example.com/page",
			url2:     "http://example.com/page",
			expected: false,
		},
		{
			name:     "with query params - same",
			url1:     "https://example.com/page?a=1",
			url2:     "https://example.com/page?a=1",
			expected: true,
		},
		{
			name:     "with query params - different",
			url1:     "https://example.com/page?a=1",
			url2:     "https://example.com/page?a=2",
			expected: false,
		},
		{
			name:     "root with query - missing slash before query",
			url1:     "https://example.com?query=1",
			url2:     "https://example.com/?query=1",
			expected: true,
		},
		{
			name:     "root with query - missing slash before query (reversed)",
			url1:     "https://example.com/?query=1",
			url2:     "https://example.com?query=1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := urlsMatchIgnoringFragment(tt.url1, tt.url2)
			if result != tt.expected {
				t.Errorf("urlsMatchIgnoringFragment(%q, %q) = %v, want %v",
					tt.url1, tt.url2, result, tt.expected)
			}
		})
	}
}
