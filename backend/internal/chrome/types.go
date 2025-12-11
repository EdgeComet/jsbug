package chrome

import "time"

// InstanceStatus represents the status of a Chrome instance
type InstanceStatus int

const (
	StatusIdle       InstanceStatus = iota // Instance is available for rendering
	StatusRendering                        // Instance is currently rendering a page
	StatusRestarting                       // Instance is being restarted
	StatusClosed                           // Instance has been closed normally
	StatusDead                             // Instance failed and cannot recover
)

// Viewport dimensions
const (
	DesktopWidth  = 1920
	DesktopHeight = 1080
	MobileWidth   = 375
	MobileHeight  = 812
)

// ShutdownTimeout is the maximum time to wait for graceful shutdown
const ShutdownTimeout = 30 * time.Second

// String returns the string representation of InstanceStatus
func (s InstanceStatus) String() string {
	switch s {
	case StatusIdle:
		return "idle"
	case StatusRendering:
		return "rendering"
	case StatusRestarting:
		return "restarting"
	case StatusClosed:
		return "closed"
	case StatusDead:
		return "dead"
	default:
		return "unknown"
	}
}

// InstanceConfig contains Chrome browser configuration
type InstanceConfig struct {
	Headless  bool
	NoSandbox bool

	// Pool-related settings
	PoolSize          int
	WarmupURL         string
	Timeout           time.Duration // General timeout for operations (warmup, render)
	RestartAfterCount int
	RestartAfterTime  time.Duration
}
