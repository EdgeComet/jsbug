package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/user/jsbug/internal/logger"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Chrome  ChromeConfig  `yaml:"chrome"`
	Logging LoggingConfig `yaml:"logging"`
	Captcha CaptchaConfig `yaml:"captcha"`
	API     APIConfig     `yaml:"api"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host        string   `yaml:"host"`
	Port        int      `yaml:"port"`
	Timeout     int      `yaml:"timeout"`
	CORSOrigins []string `yaml:"cors_origins"`
}

// ChromeConfig contains Chrome browser settings
type ChromeConfig struct {
	Headless  bool `yaml:"headless"`
	NoSandbox bool `yaml:"no_sandbox"`

	// Pool settings
	PoolSize          int           `yaml:"pool_size"`
	WarmupURL         string        `yaml:"warmup_url"`
	RestartAfterCount int           `yaml:"restart_after_count"`
	RestartAfterTime  time.Duration `yaml:"restart_after_time"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	FilePath string `yaml:"file_path"`
}

// CaptchaConfig contains Cloudflare Turnstile captcha settings
type CaptchaConfig struct {
	Enabled   bool   `yaml:"enabled"`
	SecretKey string `yaml:"secret_key"`
}

// APIConfig contains API key authentication settings
type APIConfig struct {
	Enabled bool     `yaml:"enabled"`
	Keys    []string `yaml:"keys"`
}

// Default values
const (
	defaultHost      = "0.0.0.0"
	defaultPort      = 9301
	defaultTimeout   = 30
	defaultLogLevel  = logger.LevelInfo
	defaultLogFormat = logger.FormatJSON

	// Pool defaults
	defaultPoolSize          = 4
	defaultWarmupURL         = "https://example.com/"
	defaultRestartAfterCount = 50
	defaultRestartAfterTime  = 30 * time.Minute
)

// Validation constraints
const (
	minPort = 1
	maxPort = 65535

	// Pool validation
	minPoolSize = 1
	maxPoolSize = 16
)

var validLogLevels = map[string]bool{
	logger.LevelDebug: true,
	logger.LevelInfo:  true,
	logger.LevelWarn:  true,
	logger.LevelError: true,
}

var validLogFormats = map[string]bool{
	logger.FormatJSON:    true,
	logger.FormatConsole: true,
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
	if c.Server.Timeout == 0 {
		c.Server.Timeout = defaultTimeout
	}

	// Pool defaults
	if c.Chrome.PoolSize == 0 {
		c.Chrome.PoolSize = defaultPoolSize
	}
	if c.Chrome.WarmupURL == "" {
		c.Chrome.WarmupURL = defaultWarmupURL
	}
	if c.Chrome.RestartAfterCount == 0 {
		c.Chrome.RestartAfterCount = defaultRestartAfterCount
	}
	if c.Chrome.RestartAfterTime == 0 {
		c.Chrome.RestartAfterTime = defaultRestartAfterTime
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

	if poolSize := os.Getenv("JSBUG_POOL_SIZE"); poolSize != "" {
		if p, err := strconv.Atoi(poolSize); err == nil {
			c.Chrome.PoolSize = p
		}
	}

	if logLevel := os.Getenv("JSBUG_LOG_LEVEL"); logLevel != "" {
		c.Logging.Level = logLevel
	}

	if corsOrigins := os.Getenv("JSBUG_CORS_ORIGINS"); corsOrigins != "" {
		c.Server.CORSOrigins = strings.Split(corsOrigins, ",")
	}

	// Captcha overrides
	if captchaEnabled := os.Getenv("JSBUG_CAPTCHA_ENABLED"); captchaEnabled != "" {
		c.Captcha.Enabled = strings.ToLower(captchaEnabled) == "true"
	}
	if captchaSecret := os.Getenv("JSBUG_CAPTCHA_SECRET_KEY"); captchaSecret != "" {
		c.Captcha.SecretKey = captchaSecret
	}

	// API key overrides
	if apiKeys := os.Getenv("JSBUG_API_KEYS"); apiKeys != "" {
		parts := strings.Split(apiKeys, ",")
		var filteredKeys []string
		for _, key := range parts {
			trimmed := strings.TrimSpace(key)
			if trimmed != "" {
				filteredKeys = append(filteredKeys, trimmed)
			}
		}
		if len(filteredKeys) > 0 {
			c.API.Keys = filteredKeys
			c.API.Enabled = true
		}
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate port
	if c.Server.Port < minPort || c.Server.Port > maxPort {
		return fmt.Errorf("invalid port: %d (must be %d-%d)", c.Server.Port, minPort, maxPort)
	}

	// Validate pool settings
	if c.Chrome.PoolSize < minPoolSize || c.Chrome.PoolSize > maxPoolSize {
		return fmt.Errorf("invalid pool_size: %d (must be %d-%d)", c.Chrome.PoolSize, minPoolSize, maxPoolSize)
	}
	// Validate log level
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be one of: debug, info, warn, error)", c.Logging.Level)
	}

	// Validate log format
	if !validLogFormats[c.Logging.Format] {
		return fmt.Errorf("invalid log format: %s (must be one of: json, console)", c.Logging.Format)
	}

	// Validate captcha config
	if c.Captcha.Enabled && c.Captcha.SecretKey == "" {
		return fmt.Errorf("captcha is enabled but secret_key is not set")
	}

	// Validate API config
	if c.API.Enabled && len(c.API.Keys) == 0 {
		return fmt.Errorf("API enabled but no keys configured")
	}
	for _, key := range c.API.Keys {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("API keys must not be empty")
		}
	}

	return nil
}

// ChromeTimeout returns the Chrome render timeout derived from server timeout
func (c *Config) ChromeTimeout() int {
	return c.Server.Timeout - 5
}
