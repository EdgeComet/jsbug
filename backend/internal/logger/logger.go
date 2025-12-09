package logger

import (
	"fmt"

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

// New creates a new zap logger with the specified level and format
func New(level string, format string) (*zap.Logger, error) {
	zapLevel, err := ParseLevel(level)
	if err != nil {
		return nil, err
	}

	var config zap.Config

	switch format {
	case FormatJSON:
		config = zap.NewProductionConfig()
	case FormatConsole:
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		return nil, fmt.Errorf("invalid log format: %s", format)
	}

	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.DisableCaller = true

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}
