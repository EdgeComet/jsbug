package chrome

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"

	"github.com/user/jsbug/internal/parser"
	"github.com/user/jsbug/internal/types"
)

// NetworkRequestData holds data for a single network request
type NetworkRequestData struct {
	RequestID     string
	URL           string
	Method        string
	ResourceType  string
	Status        int
	SizeBytes     int64
	ReceivedBytes int64 // Accumulated from dataReceived events (fallback for size)
	StartTime     time.Time
	EndTime       time.Time
	Blocked       bool
	Failed        bool
	FailureReason string
}

// ConsoleMessageData holds data for a console message
type ConsoleMessageData struct {
	Level   string
	Message string
	Time    time.Time
}

// JSErrorData holds data for a JavaScript error
type JSErrorData struct {
	Message    string
	Source     string
	Line       int
	Column     int
	StackTrace string
	Time       time.Time
}

// EventCollector collects Chrome DevTools Protocol events
type EventCollector struct {
	networkRequests map[string]*NetworkRequestData
	consoleMessages []ConsoleMessageData
	jsErrors        []JSErrorData
	lifecycleEvents map[string]time.Time
	startTime       time.Time
	pageURL         string // For internal/external request detection
	mu              sync.RWMutex
	logger          *zap.Logger

	// Navigation tracking for lifecycle event matching
	frameID  string
	loaderID string

	// Redirect tracking
	redirectURL    string
	redirectStatus int

	// Fetch handler tracking
	fetchHandlerCount int64
}

// NewEventCollector creates a new EventCollector
func NewEventCollector(logger *zap.Logger) *EventCollector {
	return &EventCollector{
		networkRequests: make(map[string]*NetworkRequestData),
		consoleMessages: make([]ConsoleMessageData, 0),
		jsErrors:        make([]JSErrorData, 0),
		lifecycleEvents: make(map[string]time.Time),
		startTime:       time.Now(),
		logger:          logger,
	}
}

// SetPageURL sets the page URL for internal/external request detection
func (ec *EventCollector) SetPageURL(pageURL string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.pageURL = pageURL
}

// SetNavigationIDs sets the frame and loader IDs for lifecycle event matching
func (ec *EventCollector) SetNavigationIDs(frameID, loaderID string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.frameID = frameID
	ec.loaderID = loaderID
}

// SetupListeners configures Chrome DevTools Protocol event listeners
func (ec *EventCollector) SetupListeners(ctx context.Context, blocklist *Blocklist) error {
	// Enable required domains and disable cache
	if err := chromedp.Run(ctx,
		fetch.Enable(),
		network.Enable(),
		network.ClearBrowserCookies(),
		network.SetCacheDisabled(true),
		page.Enable(),
		runtime.Enable(),
	); err != nil {
		return err
	}

	// Enable fetch interception if blocklist is configured
	if blocklist != nil && !blocklist.IsEmpty() {
		patterns := []*fetch.RequestPattern{
			{RequestStage: fetch.RequestStageRequest},
		}
		if err := chromedp.Run(ctx, fetch.Enable().WithPatterns(patterns)); err != nil {
			return err
		}
	}

	// Set up event listeners
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventRequestWillBeSent:
			ec.handleRequestWillBeSent(e)

		case *network.EventResponseReceived:
			ec.handleResponseReceived(e)

		case *network.EventLoadingFinished:
			ec.handleLoadingFinished(e)

		case *network.EventLoadingFailed:
			ec.handleLoadingFailed(e)

		case *network.EventDataReceived:
			ec.handleDataReceived(e)

		case *fetch.EventRequestPaused:
			ec.handleRequestPaused(ctx, e, blocklist)

		case *runtime.EventConsoleAPICalled:
			ec.handleConsoleAPICalled(e)

		case *runtime.EventExceptionThrown:
			ec.handleExceptionThrown(e)

		case *page.EventLifecycleEvent:
			ec.handleLifecycleEvent(e)
		}
	})

	return nil
}

func (ec *EventCollector) handleRequestWillBeSent(e *network.EventRequestWillBeSent) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	reqID := string(e.RequestID)
	ec.networkRequests[reqID] = &NetworkRequestData{
		RequestID:    reqID,
		URL:          e.Request.URL,
		Method:       e.Request.Method,
		ResourceType: e.Type.String(),
		StartTime:    time.Now(),
	}

	// Capture redirect information
	if e.RedirectResponse != nil &&
		urlsMatchIgnoringFragment(e.RedirectResponse.URL, ec.pageURL) &&
		e.DocumentURL == e.Request.URL &&
		e.RedirectResponse.Status != 0 {
		ec.redirectStatus = int(e.RedirectResponse.Status)
		ec.redirectURL = e.Request.URL
	}
}

func (ec *EventCollector) handleResponseReceived(e *network.EventResponseReceived) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	reqID := string(e.RequestID)
	if req, ok := ec.networkRequests[reqID]; ok {
		req.Status = int(e.Response.Status)
		// Capture size from response (may be partial, but good fallback for cached)
		if e.Response.EncodedDataLength > 0 {
			req.SizeBytes = int64(e.Response.EncodedDataLength)
		}
		// Set EndTime as fallback (LoadingFinished may not fire for cached)
		if req.EndTime.IsZero() {
			req.EndTime = time.Now()
		}
	}
}

func (ec *EventCollector) handleLoadingFinished(e *network.EventLoadingFinished) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	reqID := string(e.RequestID)
	if req, ok := ec.networkRequests[reqID]; ok {
		req.EndTime = time.Now()
		// LoadingFinished has the authoritative final size
		if e.EncodedDataLength > 0 {
			req.SizeBytes = int64(e.EncodedDataLength)
		}
	}
}

func (ec *EventCollector) handleLoadingFailed(e *network.EventLoadingFailed) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	reqID := string(e.RequestID)
	if req, ok := ec.networkRequests[reqID]; ok {
		req.Failed = true
		req.FailureReason = e.ErrorText
		req.EndTime = time.Now()
	}
}

func (ec *EventCollector) handleDataReceived(e *network.EventDataReceived) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	reqID := string(e.RequestID)
	if req, ok := ec.networkRequests[reqID]; ok {
		req.ReceivedBytes += e.DataLength
	}
}

func (ec *EventCollector) handleRequestPaused(ctx context.Context, e *fetch.EventRequestPaused, blocklist *Blocklist) {
	reqID := string(e.RequestID)
	reqURL := e.Request.URL
	resourceType := e.ResourceType.String()

	if blocklist != nil && blocklist.ShouldBlock(reqURL, resourceType) {
		ec.mu.Lock()
		// Mark as blocked in network requests
		if req, ok := ec.networkRequests[string(e.NetworkID)]; ok {
			req.Blocked = true
		} else {
			ec.networkRequests[string(e.NetworkID)] = &NetworkRequestData{
				RequestID:    string(e.NetworkID),
				URL:          reqURL,
				ResourceType: resourceType,
				Blocked:      true,
				StartTime:    time.Now(),
				EndTime:      time.Now(),
			}
		}
		ec.mu.Unlock()

		// Fail the request - track goroutine for completion
		atomic.AddInt64(&ec.fetchHandlerCount, 1)
		go func() {
			defer atomic.AddInt64(&ec.fetchHandlerCount, -1)
			_ = chromedp.Run(ctx, fetch.FailRequest(fetch.RequestID(reqID), network.ErrorReasonBlockedByClient))
		}()
	} else {
		// Continue the request - track goroutine for completion
		atomic.AddInt64(&ec.fetchHandlerCount, 1)
		go func() {
			defer atomic.AddInt64(&ec.fetchHandlerCount, -1)
			_ = chromedp.Run(ctx, fetch.ContinueRequest(fetch.RequestID(reqID)))
		}()
	}
}

func (ec *EventCollector) handleConsoleAPICalled(e *runtime.EventConsoleAPICalled) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	var message string
	for _, arg := range e.Args {
		if arg.Value != nil {
			message += string(arg.Value) + " "
		} else if arg.Description != "" {
			message += arg.Description + " "
		}
	}

	ec.consoleMessages = append(ec.consoleMessages, ConsoleMessageData{
		Level:   e.Type.String(),
		Message: message,
		Time:    time.Now(),
	})
}

func (ec *EventCollector) handleExceptionThrown(e *runtime.EventExceptionThrown) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	details := e.ExceptionDetails
	errData := JSErrorData{
		Message: details.Text,
		Line:    int(details.LineNumber),
		Column:  int(details.ColumnNumber),
		Time:    time.Now(),
	}

	if details.URL != "" {
		errData.Source = details.URL
	}

	if details.StackTrace != nil {
		var stack string
		for _, frame := range details.StackTrace.CallFrames {
			stack += frame.FunctionName + " at " + frame.URL + "\n"
		}
		errData.StackTrace = stack
	}

	if details.Exception != nil && details.Exception.Description != "" {
		errData.Message = details.Exception.Description
	}

	ec.jsErrors = append(ec.jsErrors, errData)
}

func (ec *EventCollector) handleLifecycleEvent(e *page.EventLifecycleEvent) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Skip events until navigation IDs are set (filters out about:blank events)
	if ec.frameID == "" || ec.loaderID == "" {
		return
	}

	// Match frameId AND loaderId to track correct navigation
	if string(e.FrameID) != ec.frameID {
		return // Ignore events from other frames
	}
	if string(e.LoaderID) != ec.loaderID {
		return // Ignore events from other navigations
	}

	ec.lifecycleEvents[e.Name] = time.Now()
}

// GetNetworkResults returns collected network data
func (ec *EventCollector) GetNetworkResults() []types.NetworkRequest {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	requests := make([]types.NetworkRequest, 0, len(ec.networkRequests))

	for _, req := range ec.networkRequests {
		// Skip the main page document request
		if urlsMatchIgnoringFragment(req.URL, ec.pageURL) && req.ResourceType == "Document" {
			continue
		}

		timeMS := int64(0)
		if !req.EndTime.IsZero() {
			timeMS = req.EndTime.Sub(req.StartTime).Milliseconds()
		}

		// Determine if request is internal (same domain as page)
		isInternal := ec.isInternalRequest(req.URL)

		// Use ReceivedBytes as fallback when SizeBytes is 0
		size := req.SizeBytes
		if size == 0 {
			size = req.ReceivedBytes
		}

		requests = append(requests, types.NetworkRequest{
			ID:         req.RequestID,
			URL:        req.URL,
			Method:     req.Method,
			Status:     req.Status,
			Type:       req.ResourceType,
			Size:       int(size),
			Time:       float64(timeMS) / 1000.0,
			IsInternal: isInternal,
			Blocked:    req.Blocked,
			Failed:     req.Failed,
		})
	}

	return requests
}

// isInternalRequest checks if a request URL is internal to the page domain
func (ec *EventCollector) isInternalRequest(requestURL string) bool {
	if ec.pageURL == "" {
		return false
	}
	return parser.IsSubdomainOf(requestURL, ec.pageURL)
}

// GetConsoleResults returns collected console messages
func (ec *EventCollector) GetConsoleResults() []types.ConsoleMessage {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	messages := make([]types.ConsoleMessage, 0, len(ec.consoleMessages))

	for i, msg := range ec.consoleMessages {
		// Convert to seconds (float64) for the Timestamp field
		timeSeconds := float64(msg.Time.Sub(ec.startTime).Milliseconds()) / 1000.0

		// Generate sequential ID: "console-1", "console-2", etc.
		id := fmt.Sprintf("console-%d", i+1)

		messages = append(messages, types.ConsoleMessage{
			ID:      id,
			Level:   msg.Level,
			Message: msg.Message,
			Time:    timeSeconds,
		})
	}

	return messages
}

// GetJSErrors returns collected JavaScript errors
func (ec *EventCollector) GetJSErrors() []types.JSError {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	errors := make([]types.JSError, 0, len(ec.jsErrors))
	for _, err := range ec.jsErrors {
		timeSeconds := float64(err.Time.Sub(ec.startTime).Milliseconds()) / 1000.0
		errors = append(errors, types.JSError{
			Message:    err.Message,
			Source:     err.Source,
			Line:       err.Line,
			Column:     err.Column,
			StackTrace: err.StackTrace,
			Timestamp:  timeSeconds,
		})
	}

	return errors
}

func (ec *EventCollector) GetLifecycleResults() []types.LifecycleEvent {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	var events []types.LifecycleEvent

	if t, ok := ec.lifecycleEvents["DOMContentLoaded"]; ok {
		events = append(events, types.LifecycleEvent{
			Event: "DOMContentLoaded",
			Time:  float64(t.Sub(ec.startTime).Milliseconds()) / 1000.0,
		})
	}
	if t, ok := ec.lifecycleEvents["load"]; ok {
		events = append(events, types.LifecycleEvent{
			Event: "load",
			Time:  float64(t.Sub(ec.startTime).Milliseconds()) / 1000.0,
		})
	}
	if t, ok := ec.lifecycleEvents["networkIdle"]; ok {
		events = append(events, types.LifecycleEvent{
			Event: "networkIdle",
			Time:  float64(t.Sub(ec.startTime).Milliseconds()) / 1000.0,
		})
	}

	return events
}

// ActiveRequestCount returns the number of in-flight requests
func (ec *EventCollector) ActiveRequestCount() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	count := 0
	for _, req := range ec.networkRequests {
		if req.EndTime.IsZero() && !req.Blocked && !req.Failed {
			count++
		}
	}
	return count
}

// GetRedirectInfo returns redirect information if a redirect was detected
func (ec *EventCollector) GetRedirectInfo() (string, int) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.redirectURL, ec.redirectStatus
}

// WaitForFetchHandlers waits for all fetch handler goroutines to complete
func (ec *EventCollector) WaitForFetchHandlers(timeout time.Duration) {
	deadline := time.After(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		if atomic.LoadInt64(&ec.fetchHandlerCount) <= 0 {
			return
		}

		select {
		case <-deadline:
			ec.logger.Warn("Timeout waiting for fetch handlers to complete",
				zap.Int64("remaining", atomic.LoadInt64(&ec.fetchHandlerCount)))
			return
		case <-ticker.C:
			// Continue waiting
		}
	}
}

// urlsMatchIgnoringFragment compares URLs while ignoring fragments and handling encoding differences
func urlsMatchIgnoringFragment(url1, url2 string) bool {
	// Strip fragments from both URLs
	base1 := url1
	if idx := strings.Index(url1, "#"); idx > -1 {
		base1 = url1[:idx]
	}

	base2 := url2
	if idx := strings.Index(url2, "#"); idx > -1 {
		base2 = url2[:idx]
	}

	// Fast path: exact match
	if base1 == base2 {
		return true
	}

	// Decode both URLs to handle encoding differences
	decoded1, err1 := url.QueryUnescape(base1)
	decoded2, err2 := url.QueryUnescape(base2)
	if err1 == nil && err2 == nil && decoded1 == decoded2 {
		return true
	}

	// Try parsing and comparing as proper URLs (handles more complex cases)
	parsed1, err1 := url.Parse(base1)
	parsed2, err2 := url.Parse(base2)
	if err1 != nil || err2 != nil {
		return false
	}

	// Compare scheme, host, and path (case-sensitive for path, case-insensitive for host)
	if !strings.EqualFold(parsed1.Host, parsed2.Host) {
		return false
	}
	if parsed1.Scheme != parsed2.Scheme {
		return false
	}
	// Normalize paths: treat empty path and "/" as equivalent (root)
	path1 := parsed1.Path
	path2 := parsed2.Path
	if path1 == "" {
		path1 = "/"
	}
	if path2 == "" {
		path2 = "/"
	}
	if path1 != path2 {
		return false
	}

	// Compare query parameters (order-independent)
	return parsed1.RawQuery == parsed2.RawQuery
}
