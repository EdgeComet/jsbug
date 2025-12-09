package chrome

import (
	"context"
	"fmt"
	"sync"

	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

// Instance represents a Chrome browser instance
type Instance struct {
	config          InstanceConfig
	logger          *zap.Logger
	allocatorCtx    context.Context
	allocatorCancel context.CancelFunc
	browserCtx      context.Context
	browserCancel   context.CancelFunc
	status          InstanceStatus
	mu              sync.RWMutex
}

// New creates a new Chrome instance
func New(cfg InstanceConfig, logger *zap.Logger) (*Instance, error) {
	opts := buildAllocatorOptions(cfg)

	allocatorCtx, allocatorCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	browserCtx, browserCancel := chromedp.NewContext(allocatorCtx,
		chromedp.WithLogf(func(format string, args ...interface{}) {
			logger.Debug(fmt.Sprintf(format, args...))
		}),
	)

	instance := &Instance{
		config:          cfg,
		logger:          logger,
		allocatorCtx:    allocatorCtx,
		allocatorCancel: allocatorCancel,
		browserCtx:      browserCtx,
		browserCancel:   browserCancel,
		status:          StatusIdle,
	}

	if err := chromedp.Run(browserCtx, chromedp.Navigate("about:blank")); err != nil {
		allocatorCancel()
		return nil, fmt.Errorf("failed to start Chrome: %w", err)
	}

	logger.Info("Chrome instance started",
		zap.Bool("headless", cfg.Headless),
		zap.Int("viewport_width", cfg.ViewportWidth),
		zap.Int("viewport_height", cfg.ViewportHeight),
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
		chromedp.WindowSize(cfg.ViewportWidth, cfg.ViewportHeight),
		chromedp.Flag("disk-cache-dir", "/dev/null"),
		chromedp.Flag("disk-cache-size", "1"),
	)

	if cfg.Headless {
		opts = append(opts, chromedp.Headless)
	}

	if cfg.DisableGPU {
		opts = append(opts, chromedp.DisableGPU)
	}

	if cfg.NoSandbox {
		opts = append(opts, chromedp.NoSandbox)
	}

	if cfg.ExecutablePath != "" {
		opts = append(opts, chromedp.ExecPath(cfg.ExecutablePath))
	}

	return opts
}

// GetContext creates a new tab context from the browser context
func (i *Instance) GetContext() (context.Context, context.CancelFunc) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.status = StatusBusy
	return chromedp.NewContext(i.browserCtx)
}

// SetStatus sets the instance status
func (i *Instance) SetStatus(status InstanceStatus) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.status = status
}

// Status returns the current instance status
func (i *Instance) Status() InstanceStatus {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.status
}

// IsAlive checks if the browser context is still valid
func (i *Instance) IsAlive() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if i.status == StatusClosed {
		return false
	}

	select {
	case <-i.browserCtx.Done():
		return false
	default:
		return true
	}
}

// Close shuts down the Chrome instance
func (i *Instance) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.status == StatusClosed {
		return nil
	}

	i.status = StatusClosed

	if i.browserCancel != nil {
		i.browserCancel()
	}

	if i.allocatorCancel != nil {
		i.allocatorCancel()
	}

	i.logger.Info("Chrome instance closed")
	return nil
}
