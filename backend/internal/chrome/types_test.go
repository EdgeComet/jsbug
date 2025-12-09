package chrome

import "testing"

func TestInstanceStatus_String(t *testing.T) {
	tests := []struct {
		status   InstanceStatus
		expected string
	}{
		{StatusIdle, "idle"},
		{StatusBusy, "busy"},
		{StatusClosed, "closed"},
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
