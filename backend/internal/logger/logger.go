package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log levels
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// Log formats
const (
	FormatJSON    = "json"
	FormatConsole = "console"
)

// ParseLevel converts a string level to zapcore.Level
func ParseLevel(level string) (zapcore.Level, error) {
	switch level {
	case LevelDebug:
		return zapcore.DebugLevel, nil
	case LevelInfo:
		return zapcore.InfoLevel, nil
	case LevelWarn:
		return zapcore.WarnLevel, nil
	case LevelError:
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

// ensureLogDirectory creates the parent directory for the log file if it doesn't exist
func ensureLogDirectory(filePath string) error {
	dir := filepath.Dir(filePath)
	if dir == "" || dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", dir, err)
	}
	return nil
}

// New creates a new zap logger with the specified level, format, and optional file path.
// Returns the logger and a cleanup function that should be called on shutdown.
// The cleanup function syncs the logger and closes any open file handles.
func New(level string, format string, filePath string) (*zap.Logger, func(), error) {
	zapLevel, err := ParseLevel(level)
	if err != nil {
		return nil, nil, err
	}

	// Create console encoder based on format
	var consoleEncoder zapcore.Encoder
	switch format {
	case FormatJSON:
		consoleEncoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	case FormatConsole:
		consoleConfig := zap.NewDevelopmentEncoderConfig()
		consoleConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		consoleEncoder = zapcore.NewConsoleEncoder(consoleConfig)
	default:
		return nil, nil, fmt.Errorf("invalid log format: %s", format)
	}

	// Console core (always enabled)
	consoleWriter := zapcore.Lock(os.Stdout)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, zapLevel)

	// Track file for cleanup
	var logFile *os.File

	var core zapcore.Core
	if filePath != "" {
		if err := ensureLogDirectory(filePath); err != nil {
			return nil, nil, err
		}

		logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file %s: %w", filePath, err)
		}

		// File encoder: same format but without ANSI colors
		var fileEncoder zapcore.Encoder
		switch format {
		case FormatJSON:
			fileEncoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		case FormatConsole:
			fileConfig := zap.NewDevelopmentEncoderConfig()
			fileConfig.EncodeLevel = zapcore.CapitalLevelEncoder // No colors for file
			fileEncoder = zapcore.NewConsoleEncoder(fileConfig)
		}

		fileWriter := zapcore.AddSync(logFile)
		fileCore := zapcore.NewCore(fileEncoder, fileWriter, zapLevel)

		// Combine console and file cores
		core = zapcore.NewTee(consoleCore, fileCore)
	} else {
		core = consoleCore
	}

	logger := zap.New(core)

	// Cleanup function: sync logger and close file
	cleanup := func() {
		logger.Sync()
		if logFile != nil {
			logFile.Close()
		}
	}

	return logger, cleanup, nil
}
