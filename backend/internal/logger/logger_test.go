package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		expected  zapcore.Level
		expectErr bool
	}{
		{
			name:      "debug level",
			level:     LevelDebug,
			expected:  zapcore.DebugLevel,
			expectErr: false,
		},
		{
			name:      "info level",
			level:     LevelInfo,
			expected:  zapcore.InfoLevel,
			expectErr: false,
		},
		{
			name:      "warn level",
			level:     LevelWarn,
			expected:  zapcore.WarnLevel,
			expectErr: false,
		},
		{
			name:      "error level",
			level:     LevelError,
			expected:  zapcore.ErrorLevel,
			expectErr: false,
		},
		{
			name:      "invalid level",
			level:     "invalid",
			expected:  zapcore.InfoLevel,
			expectErr: true,
		},
		{
			name:      "empty level",
			level:     "",
			expected:  zapcore.InfoLevel,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := ParseLevel(tt.level)
			if tt.expectErr {
				if err == nil {
					t.Errorf("ParseLevel(%q) expected error, got nil", tt.level)
				}
			} else {
				if err != nil {
					t.Errorf("ParseLevel(%q) unexpected error: %v", tt.level, err)
				}
				if level != tt.expected {
					t.Errorf("ParseLevel(%q) = %v, want %v", tt.level, level, tt.expected)
				}
			}
		})
	}
}

func TestNew_ValidLevels(t *testing.T) {
	levels := []string{LevelDebug, LevelInfo, LevelWarn, LevelError}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			logger, err := New(level, FormatJSON)
			if err != nil {
				t.Errorf("New(%q, %q) unexpected error: %v", level, FormatJSON, err)
			}
			if logger == nil {
				t.Errorf("New(%q, %q) returned nil logger", level, FormatJSON)
			}
			if logger != nil {
				logger.Sync()
			}
		})
	}
}

func TestNew_ValidFormats(t *testing.T) {
	formats := []string{FormatJSON, FormatConsole}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			logger, err := New(LevelInfo, format)
			if err != nil {
				t.Errorf("New(%q, %q) unexpected error: %v", LevelInfo, format, err)
			}
			if logger == nil {
				t.Errorf("New(%q, %q) returned nil logger", LevelInfo, format)
			}
			if logger != nil {
				logger.Sync()
			}
		})
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	logger, err := New("invalid", FormatJSON)
	if err == nil {
		t.Error("New() with invalid level expected error, got nil")
	}
	if logger != nil {
		t.Error("New() with invalid level expected nil logger")
	}
}

func TestNew_InvalidFormat(t *testing.T) {
	logger, err := New(LevelInfo, "xml")
	if err == nil {
		t.Error("New() with invalid format expected error, got nil")
	}
	if logger != nil {
		t.Error("New() with invalid format expected nil logger")
	}
}

func TestNew_LoggerWorks(t *testing.T) {
	logger, err := New(LevelDebug, FormatJSON)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer logger.Sync()

	// Verify logger can be used without panicking
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
}
