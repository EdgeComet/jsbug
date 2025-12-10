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
	ExecutablePath string
	Headless       bool
	DisableGPU     bool
	NoSandbox      bool
	ViewportWidth  int
	ViewportHeight int

	// Pool-related settings
	PoolSize          int
	WarmupURL         string
	WarmupTimeout     time.Duration
	RestartAfterCount int
	RestartAfterTime  time.Duration
	ShutdownTimeout   time.Duration
}
