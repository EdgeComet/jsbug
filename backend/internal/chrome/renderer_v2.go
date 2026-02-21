package chrome

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	cdpruntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"

	"github.com/user/jsbug/internal/types"
)

const (
	maxConsoleErrorsSize = 5120 // Maximum total size of console error messages in bytes (5KB)
)

// RenderOptions contains options for rendering a page
type RenderOptions struct {
	URL               string
	UserAgent         string
	Timeout           time.Duration
	WaitEvent         string
	Blocklist         *Blocklist
	IsMobile          bool
	CaptureScreenshot bool
}

// RenderResult contains the results of rendering a page
type RenderResult struct {
	HTML          string
	FinalURL      string
	RedirectURL   string // Set when redirect was detected (original URL differs from FinalURL)
	StatusCode    int
	PageSizeBytes int
	RenderTime    float64 // seconds
	Network       []types.NetworkRequest
	Console       []types.ConsoleMessage
	JSErrors      []types.JSError
	Lifecycle     []types.LifecycleEvent
	Screenshot    []byte `json:"-"` // PNG screenshot data, excluded from JSON serialization
}

// RendererV2 handles page rendering using Chrome with improved task-based architecture
type RendererV2 struct {
	instance  *Instance
	logger    *zap.Logger
	serviceID string
}

// NewRendererV2 creates a new RendererV2
func NewRendererV2(instance *Instance, logger *zap.Logger) *RendererV2 {
	return &RendererV2{
		instance:  instance,
		logger:    logger,
		serviceID: "jsbug-renderer",
	}
}

// renderState holds mutable state during rendering
type renderState struct {
	html          string
	finalURL      string
	statusCode    int
	headers       map[string]string
	errorMessages []string
	lifecycle     []types.LifecycleEvent
	timedOut      bool
	screenshot    []byte
	mu            sync.Mutex
}

// Render navigates to a URL and captures page data using the task-based pattern
func (r *RendererV2) Render(ctx context.Context, opts RenderOptions) (*RenderResult, error) {
	startTime := time.Now()

	// Create new tab context from browser context
	tabCtx, tabCancel := r.instance.GetContext()
	defer tabCancel()

	// Cancel tab when request context times out or is cancelled
	// This allows both soft timeout (in navigation) and hard timeout (via context) to work
	stop := context.AfterFunc(ctx, tabCancel)
	defer stop()

	// Initialize render state
	state := &renderState{
		headers: make(map[string]string),
	}

	// Create event collector for network/console data
	collector := NewEventCollector(r.logger)
	collector.SetPageURL(opts.URL)

	// Track active fetch handler goroutines
	var fetchHandlerCount int64

	// Execute rendering tasks
	err := chromedp.Run(tabCtx, r.buildTasks(opts, state, collector, &fetchHandlerCount))

	renderTime := time.Since(startTime)

	// Check hard timeout FIRST (highest priority - prevents shadowing by redirect cancellation)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(ctx.Err(), context.Canceled) {
		return &RenderResult{
			HTML:       state.html,
			FinalURL:   state.finalURL,
			StatusCode: state.statusCode,
			RenderTime: renderTime.Seconds(),
			Network:    collector.GetNetworkResults(),
			Console:    collector.GetConsoleResults(),
			JSErrors:   collector.GetJSErrors(),
			Lifecycle:  state.lifecycle,
		}, fmt.Errorf("hard timeout exceeded: %w", ctx.Err())
	}

	// Check redirect cancellation (intentional cancel with 3xx status code captured)
	state.mu.Lock()
	statusCode := state.statusCode
	state.mu.Unlock()

	if errors.Is(err, context.Canceled) && statusCode >= 300 && statusCode < 400 {
		// Redirect detected - return success with captured data
		return r.buildResult(state, collector, renderTime), nil
	}

	// Handle other errors
	if err != nil {
		result := r.buildResult(state, collector, renderTime)
		return result, err
	}

	// Validate status code was captured
	state.mu.Lock()
	finalStatusCode := state.statusCode
	state.mu.Unlock()

	if finalStatusCode == 0 {
		r.logger.Error("Status code capture failed completely (event + fallback)",
			zap.String("url", opts.URL),
			zap.Float64("render_time", renderTime.Seconds()))
		result := r.buildResult(state, collector, renderTime)
		return result, fmt.Errorf("failed to capture status code")
	}

	return r.buildResult(state, collector, renderTime), nil
}

// buildResult constructs the RenderResult from collected state
func (r *RendererV2) buildResult(state *renderState, collector *EventCollector, renderTime time.Duration) *RenderResult {
	state.mu.Lock()
	defer state.mu.Unlock()

	result := &RenderResult{
		HTML:          state.html,
		FinalURL:      state.finalURL,
		StatusCode:    state.statusCode,
		PageSizeBytes: len(state.html),
		RenderTime:    renderTime.Seconds(),
		Network:       collector.GetNetworkResults(),
		Console:       collector.GetConsoleResults(),
		JSErrors:      collector.GetJSErrors(),
		Lifecycle:     state.lifecycle,
		Screenshot:    state.screenshot,
	}

	// Get redirect info if a redirect was detected
	redirectURL, _ := collector.GetRedirectInfo()
	if redirectURL != "" {
		result.RedirectURL = redirectURL
	}

	return result
}

// buildTasks creates the chromedp task sequence for rendering
func (r *RendererV2) buildTasks(opts RenderOptions, state *renderState, collector *EventCollector, fetchHandlerCount *int64) chromedp.Tasks {
	timeOrigin := time.Now().UnixMilli()

	return chromedp.Tasks{
		// Set up event listeners FIRST - before any CDP commands
		chromedp.ActionFunc(func(ctx context.Context) error {
			chromedp.ListenTarget(ctx, func(event interface{}) {
				switch ev := event.(type) {
				case *fetch.EventRequestPaused:
					// Handle each fetch event in a goroutine to avoid blocking
					atomic.AddInt64(fetchHandlerCount, 1)
					go func(event *fetch.EventRequestPaused) {
						defer atomic.AddInt64(fetchHandlerCount, -1)

						// Create timeout context for CDP commands to prevent goroutine leaks
						cmdCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
						defer cancel()

						// Get executor context for CDP commands
						c := chromedp.FromContext(cmdCtx)
						ctxExecutor := cdp.WithExecutor(cmdCtx, c.Target)

						// Check if request should be blocked
						shouldBlock := opts.Blocklist != nil && opts.Blocklist.ShouldBlock(event.Request.URL, string(event.ResourceType))

						if shouldBlock {
							// Block the request
							err := fetch.FailRequest(event.RequestID, network.ErrorReasonAborted).Do(ctxExecutor)
							if err != nil {
								r.logger.Warn("Failed to block request",
									zap.String("url", event.Request.URL),
									zap.Error(err))
							}

							// Track blocked request in collector
							collector.mu.Lock()
							if req, ok := collector.networkRequests[string(event.NetworkID)]; ok {
								req.Blocked = true
							} else {
								collector.networkRequests[string(event.NetworkID)] = &NetworkRequestData{
									RequestID:    string(event.NetworkID),
									URL:          event.Request.URL,
									ResourceType: string(event.ResourceType),
									Blocked:      true,
									StartTime:    time.Now(),
									EndTime:      time.Now(),
								}
							}
							collector.mu.Unlock()
						} else {
							// Allow the request to continue
							err := fetch.ContinueRequest(event.RequestID).Do(ctxExecutor)
							if err != nil {
								r.logger.Warn("Failed to continue request, failing instead to prevent hang",
									zap.String("url", event.Request.URL),
									zap.Error(err))
								// Fallback: fail the request to prevent it from hanging in paused state
								fetch.FailRequest(event.RequestID, network.ErrorReasonAborted).Do(ctxExecutor)
							}
						}
					}(ev)

				case *network.EventRequestWillBeSent:
					collector.handleRequestWillBeSent(ev)

					// Check for redirect
					if ev.RedirectResponse != nil &&
						urlsMatchIgnoringFragment(ev.RedirectResponse.URL, opts.URL) &&
						ev.DocumentURL == ev.Request.URL &&
						ev.RedirectResponse.Status != 0 {
						state.mu.Lock()
						state.statusCode = int(ev.RedirectResponse.Status)
						state.finalURL = ev.Request.URL
						state.mu.Unlock()

						// Cancel to abort navigation on redirect
						err := chromedp.Cancel(ctx)
						if err != nil {
							r.logger.Warn("Unable to cancel chrome instance on redirect",
								zap.String("url", opts.URL),
								zap.Int("status_code", int(ev.RedirectResponse.Status)))
						}
						return
					}

				case *network.EventResponseReceived:
					collector.handleResponseReceived(ev)

					// Capture initial response status code and headers
					state.mu.Lock()
					if urlsMatchIgnoringFragment(ev.Response.URL, opts.URL) && state.statusCode == 0 {
						state.statusCode = int(ev.Response.Status)

						// Capture response headers
						if ev.Response.Headers != nil {
							for key, value := range ev.Response.Headers {
								switch v := value.(type) {
								case string:
									state.headers[key] = v
								case []interface{}:
									strValues := make([]string, 0, len(v))
									for _, item := range v {
										if str, ok := item.(string); ok {
											strValues = append(strValues, str)
										}
									}
									if len(strValues) > 0 {
										state.headers[key] = strings.Join(strValues, ", ")
									}
								}
							}
						}
					}
					state.mu.Unlock()

				case *network.EventLoadingFinished:
					collector.handleLoadingFinished(ev)

				case *network.EventLoadingFailed:
					collector.handleLoadingFailed(ev)

				case *cdpruntime.EventConsoleAPICalled:
					// Capture console errors with size limiting
					if ev.Type == cdpruntime.APITypeError {
						for _, arg := range ev.Args {
							if msg, err := strconv.Unquote(string(arg.Value)); err == nil {
								state.mu.Lock()
								currentSize := 0
								for _, existingMsg := range state.errorMessages {
									currentSize += len(existingMsg)
								}
								if currentSize+len(msg) <= maxConsoleErrorsSize {
									state.errorMessages = append(state.errorMessages, msg)
								}
								state.mu.Unlock()
							}
						}
					}
					collector.handleConsoleAPICalled(ev)

				case *cdpruntime.EventExceptionThrown:
					collector.handleExceptionThrown(ev)

				case *page.EventLifecycleEvent:
					collector.handleLifecycleEvent(ev)

					// Track lifecycle events with timestamps - only for main frame
					collector.mu.RLock()
					frameID := collector.frameID
					loaderID := collector.loaderID
					collector.mu.RUnlock()

					// Skip events until navigation IDs are set (filters out about:blank events)
					if frameID == "" || loaderID == "" {
						break
					}

					// Only track events from the main frame/navigation
					if string(ev.FrameID) != frameID || string(ev.LoaderID) != loaderID {
						break // Skip events from other frames
					}

					state.mu.Lock()
					now := time.Now().UnixMilli()
					delta := now - timeOrigin
					state.lifecycle = append(state.lifecycle, types.LifecycleEvent{
						Event: string(ev.Name),
						Time:  float64(delta) / 1000.0,
					})
					state.mu.Unlock()
				}
			})
			return nil
		}),

		network.Enable(),

		// Enable fetch interception for request blocking
		chromedp.ActionFunc(func(ctx context.Context) error {
			if opts.Blocklist != nil && !opts.Blocklist.IsEmpty() {
				patterns := []*fetch.RequestPattern{
					{RequestStage: fetch.RequestStageRequest},
				}
				return fetch.Enable().WithPatterns(patterns).Do(ctx)
			}
			return nil
		}),

		network.ClearBrowserCookies(),
		page.Enable(),
		css.Disable(),

		r.enableLifeCycle(),

		// Set user agent
		chromedp.ActionFunc(func(ctx context.Context) error {
			if opts.UserAgent != "" {
				return emulation.SetUserAgentOverride(opts.UserAgent).Do(ctx)
			}
			return nil
		}),

		// Set viewport
		chromedp.ActionFunc(func(ctx context.Context) error {
			width, height := DesktopWidth, DesktopHeight
			if opts.IsMobile {
				width, height = MobileWidth, MobileHeight
			}
			return emulation.SetDeviceMetricsOverride(
				int64(width),
				int64(height),
				1.0,
				opts.IsMobile,
			).Do(ctx)
		}),

		// Navigate and wait for page ready (with soft timeout)
		r.navigateAndWait(opts, state, collector),

		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.WaitVisible("body", chromedp.ByQuery),

		r.extractHTML(&state.html),

		chromedp.Location(&state.finalURL),

		// Fallback status code retrieval (if event listener missed it)
		chromedp.ActionFunc(func(ctx context.Context) error {
			state.mu.Lock()
			currentStatusCode := state.statusCode
			state.mu.Unlock()

			if currentStatusCode == 0 {
				var statusCodeResult int64
				err := chromedp.Evaluate(`
					(function() {
						try {
							var navEntry = performance.getEntriesByType('navigation')[0];
							if (navEntry && navEntry.responseStatus) {
								return navEntry.responseStatus;
							}
							return 0;
						} catch (e) {
							return 0;
						}
					})()
				`, &statusCodeResult).Do(ctx)

				if err != nil {
					r.logger.Warn("Failed to retrieve status code via Performance API fallback",
						zap.String("url", opts.URL),
						zap.Error(err))
					return nil
				}

				if statusCodeResult > 0 {
					state.mu.Lock()
					state.statusCode = int(statusCodeResult)
					state.mu.Unlock()

					r.logger.Info("Status code retrieved via Performance API fallback",
						zap.String("url", opts.URL),
						zap.Int("status_code", int(statusCodeResult)))
				}
			}
			return nil
		}),

		// Capture screenshot (viewport only, PNG format) - only when requested
		chromedp.ActionFunc(func(ctx context.Context) error {
			if !opts.CaptureScreenshot {
				return nil
			}
			var screenshotBuf []byte
			err := chromedp.CaptureScreenshot(&screenshotBuf).Do(ctx)
			if err != nil {
				r.logger.Warn("Failed to capture screenshot",
					zap.String("url", opts.URL),
					zap.Error(err))
				// Don't fail the render if screenshot fails
				return nil
			}

			state.mu.Lock()
			state.screenshot = screenshotBuf
			state.mu.Unlock()

			return nil
		}),

		// Wait for all fetch handlers to complete BEFORE closing page
		chromedp.ActionFunc(func(ctx context.Context) error {
			timeout := time.After(5 * time.Second)
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for {
				if atomic.LoadInt64(fetchHandlerCount) <= 0 {
					return nil
				}

				select {
				case <-timeout:
					r.logger.Warn("Timeout waiting for fetch handlers to complete",
						zap.String("url", opts.URL),
						zap.Int64("remaining", atomic.LoadInt64(fetchHandlerCount)))
					return nil
				case <-ticker.C:
					// Continue waiting
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}),

		page.Close(),
	}
}

// navigateAndWait navigates to URL and waits for the specified event
func (r *RendererV2) navigateAndWait(opts RenderOptions, state *renderState, collector *EventCollector) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		// Navigate and capture frame/loader IDs
		frameID, loaderID, _, _, err := page.Navigate(opts.URL).Do(ctx)
		if err != nil {
			return fmt.Errorf("navigate failed: %w", err)
		}

		// Set navigation IDs for lifecycle event matching
		collector.SetNavigationIDs(string(frameID), string(loaderID))

		// Wait for lifecycle event with timeout (soft - we continue on timeout)
		// Default to "load" if no wait event specified
		waitEvent := opts.WaitEvent
		if waitEvent == "" {
			waitEvent = "load"
		}
		err = r.waitForLifecycleEvent(ctx, waitEvent, collector, opts.Timeout)

		// If timeout occurred, mark it but don't fail
		if err != nil && err.Error() == "wait timeout exceeded" {
			state.mu.Lock()
			state.timedOut = true
			state.mu.Unlock()

			r.logger.Debug("Navigation wait timed out, continuing with HTML extraction",
				zap.String("url", opts.URL),
				zap.Duration("timeout", opts.Timeout),
				zap.Bool("timed_out", true))

			return nil
		} else if err != nil {
			return err
		}

		return nil
	}
}

// waitForLifecycleEvent waits for a specific lifecycle event with frame/loader matching
func (r *RendererV2) waitForLifecycleEvent(ctx context.Context, eventName string, collector *EventCollector, timeout time.Duration) error {
	ch := make(chan struct{})

	listenerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	collector.mu.RLock()
	frameID := collector.frameID
	loaderID := collector.loaderID
	collector.mu.RUnlock()

	chromedp.ListenTarget(listenerCtx, func(ev interface{}) {
		if e, ok := ev.(*page.EventLifecycleEvent); ok {
			// Match both frameId AND loaderId to track correct navigation
			if string(e.FrameID) == frameID && string(e.LoaderID) == loaderID {
				if string(e.Name) == eventName {
					cancel()
					close(ch)
				}
			}
		}
	})

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(timeout):
		return fmt.Errorf("wait timeout exceeded")
	}
}

// extractHTML extracts the page HTML with retry logic
func (r *RendererV2) extractHTML(output *string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		var lastErr error

		for attempt := 0; attempt < 3; attempt++ {
			// Get document root node
			rootNode, err := dom.GetDocument().Do(ctx)
			if err != nil {
				lastErr = err
				time.Sleep(300 * time.Millisecond)
				continue
			}

			// Extract HTML
			html, err := dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(ctx)
			if err != nil {
				lastErr = err
				time.Sleep(300 * time.Millisecond)
				continue
			}

			*output = html
			return nil
		}

		return fmt.Errorf("extract HTML failed after 3 attempts: %w", lastErr)
	}
}

// enableLifeCycle enables page lifecycle events
func (r *RendererV2) enableLifeCycle() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		if err := page.Enable().Do(ctx); err != nil {
			return err
		}
		return page.SetLifecycleEventsEnabled(true).Do(ctx)
	}
}

// IsAvailable returns true if the renderer is ready to use
func (r *RendererV2) IsAvailable() bool {
	return r.instance != nil && r.instance.IsAlive()
}
