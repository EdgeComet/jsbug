package chrome

// InstanceStatus represents the status of a Chrome instance
type InstanceStatus int

const (
	StatusIdle InstanceStatus = iota
	StatusBusy
	StatusClosed
)

// String returns the string representation of InstanceStatus
func (s InstanceStatus) String() string {
	switch s {
	case StatusIdle:
		return "idle"
	case StatusBusy:
		return "busy"
	case StatusClosed:
		return "closed"
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
}
