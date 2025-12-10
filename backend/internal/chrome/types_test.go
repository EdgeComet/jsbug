package chrome

import "testing"

func TestInstanceStatus_String(t *testing.T) {
	tests := []struct {
		status   InstanceStatus
		expected string
	}{
		{StatusIdle, "idle"},
		{StatusRendering, "rendering"},
		{StatusRestarting, "restarting"},
		{StatusClosed, "closed"},
		{StatusDead, "dead"},
		{InstanceStatus(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestInstanceStatus_Values(t *testing.T) {
	// Verify status constants have expected integer values (iota order)
	if StatusIdle != 0 {
		t.Errorf("StatusIdle = %d, want 0", StatusIdle)
	}
	if StatusRendering != 1 {
		t.Errorf("StatusRendering = %d, want 1", StatusRendering)
	}
	if StatusRestarting != 2 {
		t.Errorf("StatusRestarting = %d, want 2", StatusRestarting)
	}
	if StatusClosed != 3 {
		t.Errorf("StatusClosed = %d, want 3", StatusClosed)
	}
	if StatusDead != 4 {
		t.Errorf("StatusDead = %d, want 4", StatusDead)
	}
}
