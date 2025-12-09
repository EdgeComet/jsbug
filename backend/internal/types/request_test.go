package types

import (
	"testing"
)

func TestResolveUserAgent(t *testing.T) {
	tests := []struct {
		name     string
		preset   string
		expected string
	}{
		{
			name:     "chrome preset",
			preset:   UserAgentChrome,
			expected: UserAgentPresets[UserAgentChrome],
		},
		{
			name:     "firefox preset",
			preset:   UserAgentFirefox,
			expected: UserAgentPresets[UserAgentFirefox],
		},
		{
			name:     "safari preset",
			preset:   UserAgentSafari,
			expected: UserAgentPresets[UserAgentSafari],
		},
		{
			name:     "mobile preset",
			preset:   UserAgentMobile,
			expected: UserAgentPresets[UserAgentMobile],
		},
		{
			name:     "bot preset",
			preset:   UserAgentBot,
			expected: UserAgentPresets[UserAgentBot],
		},
		{
			name:     "empty preset returns chrome default",
			preset:   "",
			expected: UserAgentPresets[DefaultUserAgent],
		},
		{
			name:     "custom user agent returned as-is",
			preset:   "Custom/1.0",
			expected: "Custom/1.0",
		},
		{
			name:     "unknown preset returned as-is",
			preset:   "unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveUserAgent(tt.preset)
			if result != tt.expected {
				t.Errorf("ResolveUserAgent(%q) = %q, want %q", tt.preset, result, tt.expected)
			}
		})
	}
}

func TestIsValidWaitEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    string
		expected bool
	}{
		{
			name:     "DOMContentLoaded is valid",
			event:    WaitDOMContentLoaded,
			expected: true,
		},
		{
			name:     "load is valid",
			event:    WaitLoad,
			expected: true,
		},
		{
			name:     "networkIdle is valid",
			event:    WaitNetworkIdle,
			expected: true,
		},
		{
			name:     "networkAlmostIdle is valid",
			event:    WaitNetworkAlmostIdle,
			expected: true,
		},
		{
			name:     "empty string is valid (uses default)",
			event:    "",
			expected: true,
		},
		{
			name:     "invalid event",
			event:    "invalid",
			expected: false,
		},
		{
			name:     "case sensitive - lowercase fails",
			event:    "domcontentloaded",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidWaitEvent(tt.event)
			if result != tt.expected {
				t.Errorf("IsValidWaitEvent(%q) = %v, want %v", tt.event, result, tt.expected)
			}
		})
	}
}

func TestRenderRequest_ApplyDefaults(t *testing.T) {
	t.Run("applies defaults to empty request", func(t *testing.T) {
		req := &RenderRequest{}
		req.ApplyDefaults()

		if req.UserAgent != DefaultUserAgent {
			t.Errorf("UserAgent = %q, want %q", req.UserAgent, DefaultUserAgent)
		}
		if req.Timeout != DefaultTimeout {
			t.Errorf("Timeout = %d, want %d", req.Timeout, DefaultTimeout)
		}
		if req.WaitEvent != DefaultWaitEvent {
			t.Errorf("WaitEvent = %q, want %q", req.WaitEvent, DefaultWaitEvent)
		}
	})

	t.Run("preserves non-empty values", func(t *testing.T) {
		req := &RenderRequest{
			UserAgent: "custom",
			Timeout:   30,
			WaitEvent: WaitNetworkIdle,
		}
		req.ApplyDefaults()

		if req.UserAgent != "custom" {
			t.Errorf("UserAgent = %q, want %q", req.UserAgent, "custom")
		}
		if req.Timeout != 30 {
			t.Errorf("Timeout = %d, want %d", req.Timeout, 30)
		}
		if req.WaitEvent != WaitNetworkIdle {
			t.Errorf("WaitEvent = %q, want %q", req.WaitEvent, WaitNetworkIdle)
		}
	})
}

func TestRenderRequest_ValidateTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int
		expected bool
	}{
		{
			name:     "minimum timeout is valid",
			timeout:  MinTimeout,
			expected: true,
		},
		{
			name:     "maximum timeout is valid",
			timeout:  MaxTimeout,
			expected: true,
		},
		{
			name:     "timeout within range is valid",
			timeout:  30,
			expected: true,
		},
		{
			name:     "zero timeout is invalid",
			timeout:  0,
			expected: false,
		},
		{
			name:     "negative timeout is invalid",
			timeout:  -1,
			expected: false,
		},
		{
			name:     "timeout above max is invalid",
			timeout:  MaxTimeout + 1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &RenderRequest{Timeout: tt.timeout}
			result := req.ValidateTimeout()
			if result != tt.expected {
				t.Errorf("ValidateTimeout() with timeout=%d returned %v, want %v", tt.timeout, result, tt.expected)
			}
		})
	}
}
