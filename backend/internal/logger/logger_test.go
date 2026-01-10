package logger

import (
	"os"
	"path/filepath"
	"strings"
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
			logger, cleanup, err := New(level, FormatJSON, "")
			if err != nil {
				t.Errorf("New(%q, %q, \"\") unexpected error: %v", level, FormatJSON, err)
			}
			if logger == nil {
				t.Errorf("New(%q, %q, \"\") returned nil logger", level, FormatJSON)
			}
			if cleanup != nil {
				cleanup()
			}
		})
	}
}

func TestNew_ValidFormats(t *testing.T) {
	formats := []string{FormatJSON, FormatConsole}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			logger, cleanup, err := New(LevelInfo, format, "")
			if err != nil {
				t.Errorf("New(%q, %q, \"\") unexpected error: %v", LevelInfo, format, err)
			}
			if logger == nil {
				t.Errorf("New(%q, %q, \"\") returned nil logger", LevelInfo, format)
			}
			if cleanup != nil {
				cleanup()
			}
		})
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	logger, _, err := New("invalid", FormatJSON, "")
	if err == nil {
		t.Error("New() with invalid level expected error, got nil")
	}
	if logger != nil {
		t.Error("New() with invalid level expected nil logger")
	}
}

func TestNew_InvalidFormat(t *testing.T) {
	logger, _, err := New(LevelInfo, "xml", "")
	if err == nil {
		t.Error("New() with invalid format expected error, got nil")
	}
	if logger != nil {
		t.Error("New() with invalid format expected nil logger")
	}
}

func TestNew_LoggerWorks(t *testing.T) {
	logger, cleanup, err := New(LevelDebug, FormatJSON, "")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer cleanup()

	// Verify logger can be used without panicking
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
}

func TestNew_WithFilePath(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")

	logger, cleanup, err := New(LevelInfo, FormatJSON, logFile)
	if err != nil {
		t.Fatalf("New() with file path unexpected error: %v", err)
	}
	defer cleanup()

	// Log a message
	logger.Info("test message for file")
	logger.Sync()

	// Verify file was created and contains the message
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "test message for file") {
		t.Errorf("Log file does not contain expected message. Content: %s", content)
	}
}

func TestNew_AutoCreateDirectory(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use a nested path that doesn't exist
	logFile := filepath.Join(tmpDir, "subdir", "nested", "test.log")

	logger, cleanup, err := New(LevelInfo, FormatJSON, logFile)
	if err != nil {
		t.Fatalf("New() with nested path unexpected error: %v", err)
	}
	defer cleanup()

	// Log a message
	logger.Info("test message")
	logger.Sync()

	// Verify file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestNew_InvalidFilePath(t *testing.T) {
	// Try to create a log file in a path that requires root permissions
	// On most systems, /root or similar should fail
	_, cleanup, err := New(LevelInfo, FormatJSON, "/root/nonexistent/test.log")

	// Should fail with permission error
	if err == nil {
		cleanup()
		t.Error("New() with invalid path expected error, got nil")
	}
}

func TestNew_ConsoleFormatNoANSIInFile(t *testing.T) {
	// Verify that file output doesn't contain ANSI color codes when using console format
	tmpDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")

	logger, cleanup, err := New(LevelInfo, FormatConsole, logFile)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer cleanup()

	logger.Info("test message")
	logger.Sync()

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// ANSI escape codes start with \x1b[ (ESC[)
	if strings.Contains(string(content), "\x1b[") {
		t.Errorf("File output contains ANSI escape codes: %q", content)
	}

	// Verify the message is still there
	if !strings.Contains(string(content), "test message") {
		t.Errorf("Log file does not contain expected message. Content: %s", content)
	}
}

func TestEnsureLogDirectory(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		expectErr bool
	}{
		{
			name:      "empty path",
			path:      "",
			expectErr: false,
		},
		{
			name:      "current directory",
			path:      "test.log",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ensureLogDirectory(tt.path)
			if tt.expectErr && err == nil {
				t.Error("ensureLogDirectory() expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("ensureLogDirectory() unexpected error: %v", err)
			}
		})
	}
}

func TestEnsureLogDirectory_CreatesDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	newDir := filepath.Join(tmpDir, "newsubdir", "nested")
	logPath := filepath.Join(newDir, "test.log")

	err = ensureLogDirectory(logPath)
	if err != nil {
		t.Fatalf("ensureLogDirectory() unexpected error: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(newDir)
	if os.IsNotExist(err) {
		t.Error("Directory was not created")
	}
	if err == nil && !info.IsDir() {
		t.Error("Created path is not a directory")
	}
}
