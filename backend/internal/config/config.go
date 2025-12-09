package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Chrome  ChromeConfig  `yaml:"chrome"`
	Logging LoggingConfig `yaml:"logging"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host         string   `yaml:"host"`
	Port         int      `yaml:"port"`
	ReadTimeout  int      `yaml:"read_timeout"`
	WriteTimeout int      `yaml:"write_timeout"`
	CORSOrigins  []string `yaml:"cors_origins"`
}

// ChromeConfig contains Chrome browser settings
type ChromeConfig struct {
	ExecutablePath string `yaml:"executable_path"`
	Headless       bool   `yaml:"headless"`
	DisableGPU     bool   `yaml:"disable_gpu"`
	NoSandbox      bool   `yaml:"no_sandbox"`
	TimeoutDefault int    `yaml:"timeout_default"`
	TimeoutMax     int    `yaml:"timeout_max"`
	ViewportWidth  int    `yaml:"viewport_width"`
	ViewportHeight int    `yaml:"viewport_height"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Valid log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// Valid log formats
const (
	LogFormatJSON    = "json"
	LogFormatConsole = "console"
)

// Default values
const (
	defaultHost           = "0.0.0.0"
	defaultPort           = 9301
	defaultReadTimeout    = 30
	defaultWriteTimeout   = 30
	defaultHeadless       = true
	defaultTimeoutDefault = 15
	defaultTimeoutMax     = 60
	defaultViewportWidth  = 1920
	defaultViewportHeight = 1080
	defaultLogLevel       = LogLevelInfo
	defaultLogFormat      = LogFormatJSON
)

// Validation constraints
const (
	minPort           = 1
	maxPort           = 65535
	minTimeout        = 1
	maxTimeout        = 60
	minViewportWidth  = 1
	minViewportHeight = 1
)

var validLogLevels = map[string]bool{
	LogLevelDebug: true,
	LogLevelInfo:  true,
	LogLevelWarn:  true,
	LogLevelError: true,
}

var validLogFormats = map[string]bool{
	LogFormatJSON:    true,
	LogFormatConsole: true,
}

// Load reads configuration from a YAML file and applies environment overrides
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.applyDefaults()
	cfg.applyEnvOverrides()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// applyDefaults sets default values for unset fields
func (c *Config) applyDefaults() {
	// Server defaults
	if c.Server.Host == "" {
		c.Server.Host = defaultHost
	}
	if c.Server.Port == 0 {
		c.Server.Port = defaultPort
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = defaultReadTimeout
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = defaultWriteTimeout
	}

	// Chrome defaults
	if c.Chrome.TimeoutDefault == 0 {
		c.Chrome.TimeoutDefault = defaultTimeoutDefault
	}
	if c.Chrome.TimeoutMax == 0 {
		c.Chrome.TimeoutMax = defaultTimeoutMax
	}
	if c.Chrome.ViewportWidth == 0 {
		c.Chrome.ViewportWidth = defaultViewportWidth
	}
	if c.Chrome.ViewportHeight == 0 {
		c.Chrome.ViewportHeight = defaultViewportHeight
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = defaultLogLevel
	}
	if c.Logging.Format == "" {
		c.Logging.Format = defaultLogFormat
	}
}

// applyEnvOverrides applies environment variable overrides
func (c *Config) applyEnvOverrides() {
	if port := os.Getenv("JSBUG_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Server.Port = p
		}
	}

	if chromePath := os.Getenv("JSBUG_CHROME_PATH"); chromePath != "" {
		c.Chrome.ExecutablePath = chromePath
	}

	if logLevel := os.Getenv("JSBUG_LOG_LEVEL"); logLevel != "" {
		c.Logging.Level = logLevel
	}

	if corsOrigins := os.Getenv("JSBUG_CORS_ORIGINS"); corsOrigins != "" {
		c.Server.CORSOrigins = strings.Split(corsOrigins, ",")
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate port
	if c.Server.Port < minPort || c.Server.Port > maxPort {
		return fmt.Errorf("invalid port: %d (must be %d-%d)", c.Server.Port, minPort, maxPort)
	}

	// Validate timeout default
	if c.Chrome.TimeoutDefault < minTimeout || c.Chrome.TimeoutDefault > maxTimeout {
		return fmt.Errorf("invalid timeout_default: %d (must be %d-%d)", c.Chrome.TimeoutDefault, minTimeout, maxTimeout)
	}

	// Validate timeout max
	if c.Chrome.TimeoutMax < c.Chrome.TimeoutDefault {
		return fmt.Errorf("timeout_max (%d) must be >= timeout_default (%d)", c.Chrome.TimeoutMax, c.Chrome.TimeoutDefault)
	}

	// Validate viewport dimensions
	if c.Chrome.ViewportWidth < minViewportWidth {
		return fmt.Errorf("invalid viewport_width: %d (must be > 0)", c.Chrome.ViewportWidth)
	}
	if c.Chrome.ViewportHeight < minViewportHeight {
		return fmt.Errorf("invalid viewport_height: %d (must be > 0)", c.Chrome.ViewportHeight)
	}

	// Validate log level
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be one of: debug, info, warn, error)", c.Logging.Level)
	}

	// Validate log format
	if !validLogFormats[c.Logging.Format] {
		return fmt.Errorf("invalid log format: %s (must be one of: json, console)", c.Logging.Format)
	}

	return nil
}
