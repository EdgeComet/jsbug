package chrome

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

const healthCheckTimeout = 5 * time.Second

// Instance represents a Chrome browser instance
type Instance struct {
	id              int
	config          InstanceConfig
	logger          *zap.Logger
	allocatorCtx    context.Context
	allocatorCancel context.CancelFunc
	browserCtx      context.Context
	browserCancel   context.CancelFunc
	status          atomic.Int32
	renderCount     atomic.Int64
	createdAt       atomic.Int64
	mu              sync.RWMutex // protects context fields only
}

// New creates a new Chrome instance with the given ID
func New(id int, cfg InstanceConfig, logger *zap.Logger) (*Instance, error) {
	instance := &Instance{
		id:     id,
		config: cfg,
		logger: logger,
	}
	instance.status.Store(int32(StatusIdle))

	allocCtx, allocCancel, browserCtx, browserCancel, err := instance.createBrowser()
	if err != nil {
		return nil, err
	}

	instance.createdAt.Store(time.Now().UnixNano()) // Set after browser is ready

	instance.allocatorCtx = allocCtx
	instance.allocatorCancel = allocCancel
	instance.browserCtx = browserCtx
	instance.browserCancel = browserCancel

	logger.Info("Chrome instance started",
		zap.Int("id", id),
		zap.Bool("headless", cfg.Headless),
	)

	return instance, nil
}

// buildAllocatorOptions creates Chrome allocator options from config
func buildAllocatorOptions(cfg InstanceConfig) []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-translate", true),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.WindowSize(DesktopWidth, DesktopHeight),
		chromedp.Flag("disk-cache-dir", "/dev/null"),
		chromedp.Flag("disk-cache-size", "1"),
	)

	if cfg.Headless {
		opts = append(opts, chromedp.Headless)
	}

	// Always disable GPU for headless rendering
	opts = append(opts, chromedp.DisableGPU)

	if cfg.NoSandbox {
		opts = append(opts, chromedp.NoSandbox)
	}

	return opts
}

// ID returns the instance identifier
func (i *Instance) ID() int {
	return i.id
}

// Status returns the current instance status (atomic read)
func (i *Instance) Status() InstanceStatus {
	return InstanceStatus(i.status.Load())
}

// SetStatus sets the instance status (atomic write)
func (i *Instance) SetStatus(status InstanceStatus) {
	i.status.Store(int32(status))
}

// RenderCount returns the number of completed renders (atomic read)
func (i *Instance) RenderCount() int64 {
	return i.renderCount.Load()
}

// IncrementRenders increments the render count by 1 (atomic)
func (i *Instance) IncrementRenders() {
	i.renderCount.Add(1)
}

// CreatedAt returns the time when the instance was created
func (i *Instance) CreatedAt() time.Time {
	return time.Unix(0, i.createdAt.Load())
}

// resetCounters resets renderCount to 0 and createdAt to now
func (i *Instance) resetCounters() {
	i.renderCount.Store(0)
	i.createdAt.Store(time.Now().UnixNano())
}

// GetContext creates a new tab context from the browser context
func (i *Instance) GetContext() (context.Context, context.CancelFunc) {
	i.mu.Lock()
	defer i.mu.Unlock()

	return chromedp.NewContext(i.browserCtx)
}

// IsAlive checks if the browser is responsive using a CDP health check.
// It returns false if the instance is dead, closed, or the browser doesn't respond.
func (i *Instance) IsAlive() bool {
	status := i.Status()
	if status == StatusDead || status == StatusClosed {
		return false
	}

	i.mu.RLock()
	browserCtx := i.browserCtx
	i.mu.RUnlock()

	if browserCtx == nil {
		return false
	}

	// Use context.Background() for the timeout, NOT browserCtx.
	// If the browser is dead, browserCtx may already be cancelled.
	ctx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
	defer cancel()

	// Use goroutine + channel pattern for timeout handling
	done := make(chan error, 1)
	go func() {
		done <- chromedp.Run(browserCtx, chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, _, _, _, err := browser.GetVersion().Do(ctx)
			return err
		}))
	}()

	select {
	case err := <-done:
		return err == nil
	case <-ctx.Done():
		// Health check timed out
		return false
	}
}

// Close shuts down the Chrome instance
func (i *Instance) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.Status() == StatusClosed {
		return nil
	}

	i.SetStatus(StatusClosed)

	if i.browserCancel != nil {
		i.browserCancel()
	}

	if i.allocatorCancel != nil {
		i.allocatorCancel()
	}

	i.logger.Info("Chrome instance closed", zap.Int("id", i.id))
	return nil
}

// ShouldRestart returns true if the instance should be restarted based on policy.
// It checks both render count and time-based restart policies.
func (i *Instance) ShouldRestart() bool {
	// Check render count policy
	if i.config.RestartAfterCount > 0 && i.RenderCount() >= int64(i.config.RestartAfterCount) {
		return true
	}

	// Check time-based policy
	if i.config.RestartAfterTime > 0 && time.Since(i.CreatedAt()) >= i.config.RestartAfterTime {
		return true
	}

	return false
}

// Restart restarts the Chrome browser instance.
// Uses "make before break" pattern: creates new browser before destroying old one.
// If restart fails, the old browser remains intact and usable.
func (i *Instance) Restart() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.SetStatus(StatusRestarting)

	// Create new browser FIRST (make before break)
	newAllocCtx, newAllocCancel, newBrowserCtx, newBrowserCancel, err := i.createBrowser()
	if err != nil {
		// New browser failed - keep old browser intact
		i.SetStatus(StatusIdle)
		i.logger.Warn("Restart failed, continuing with existing browser",
			zap.Int("id", i.id),
			zap.Error(err),
		)
		return fmt.Errorf("failed to restart Chrome: %w", err)
	}

	// New browser is ready - NOW cancel old contexts
	if i.browserCancel != nil {
		i.browserCancel()
	}
	if i.allocatorCancel != nil {
		i.allocatorCancel()
	}

	// Swap to new contexts
	i.allocatorCtx = newAllocCtx
	i.allocatorCancel = newAllocCancel
	i.browserCtx = newBrowserCtx
	i.browserCancel = newBrowserCancel

	// Reset counters
	i.resetCounters()

	// Perform warmup (log warning if fails but don't fail restart)
	if err := i.warmup(); err != nil {
		i.logger.Warn("Warmup failed during restart",
			zap.Int("id", i.id),
			zap.Error(err),
		)
	}

	i.SetStatus(StatusIdle)
	i.logger.Info("Chrome instance restarted",
		zap.Int("id", i.id),
	)

	return nil
}

// Terminate forcefully terminates the Chrome instance and marks it as dead.
// This is used for permanent shutdown, not restartable.
func (i *Instance) Terminate() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.SetStatus(StatusDead)

	if i.browserCancel != nil {
		i.browserCancel()
	}
	if i.allocatorCancel != nil {
		i.allocatorCancel()
	}

	i.logger.Info("Chrome instance terminated", zap.Int("id", i.id))
	return nil
}

// createBrowser creates new allocator and browser contexts.
// Returns the new contexts without modifying instance state.
// Caller is responsible for cleanup on error.
func (i *Instance) createBrowser() (
	allocCtx context.Context,
	allocCancel context.CancelFunc,
	browserCtx context.Context,
	browserCancel context.CancelFunc,
	err error,
) {
	opts := buildAllocatorOptions(i.config)

	allocCtx, allocCancel = chromedp.NewExecAllocator(context.Background(), opts...)

	browserCtx, browserCancel = chromedp.NewContext(allocCtx,
		chromedp.WithLogf(func(format string, args ...interface{}) {
			i.logger.Debug(fmt.Sprintf(format, args...))
		}),
	)

	// Test the browser by navigating to about:blank
	if err = chromedp.Run(browserCtx, chromedp.Navigate("about:blank")); err != nil {
		allocCancel()
		return nil, nil, nil, nil, fmt.Errorf("failed to start Chrome: %w", err)
	}

	return allocCtx, allocCancel, browserCtx, browserCancel, nil
}

// warmup navigates to the warmup URL to ensure the browser is ready.
// Must be called with mutex held.
func (i *Instance) warmup() error {
	if i.config.WarmupURL == "" {
		return nil
	}

	timeout := i.config.Timeout
	if timeout == 0 {
		timeout = 25 * time.Second
	}

	ctx, cancel := context.WithTimeout(i.browserCtx, timeout)
	defer cancel()

	tabCtx, tabCancel := chromedp.NewContext(ctx)
	defer tabCancel()

	if err := chromedp.Run(tabCtx, chromedp.Navigate(i.config.WarmupURL)); err != nil {
		return fmt.Errorf("warmup navigation failed: %w", err)
	}

	return nil
}
