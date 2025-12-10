package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_ValidConfig(t *testing.T) {
	content := `
server:
  host: "127.0.0.1"
  port: 9000
  cors_origins:
    - "http://localhost:3000"
chrome:
  headless: true
  timeout_default: 20
  timeout_max: 45
  viewport_width: 1280
  viewport_height: 720
logging:
  level: "debug"
  format: "console"
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 9000)
	}
	if cfg.Chrome.TimeoutDefault != 20 {
		t.Errorf("Chrome.TimeoutDefault = %d, want %d", cfg.Chrome.TimeoutDefault, 20)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want %q", cfg.Logging.Level, "debug")
	}
	if cfg.Logging.Format != "console" {
		t.Errorf("Logging.Format = %q, want %q", cfg.Logging.Format, "console")
	}
}

func TestLoad_AppliesDefaults(t *testing.T) {
	content := `
server: {}
chrome: {}
logging: {}
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != defaultHost {
		t.Errorf("Server.Host = %q, want default %q", cfg.Server.Host, defaultHost)
	}
	if cfg.Server.Port != defaultPort {
		t.Errorf("Server.Port = %d, want default %d", cfg.Server.Port, defaultPort)
	}
	if cfg.Chrome.TimeoutDefault != defaultTimeoutDefault {
		t.Errorf("Chrome.TimeoutDefault = %d, want default %d", cfg.Chrome.TimeoutDefault, defaultTimeoutDefault)
	}
	if cfg.Chrome.TimeoutMax != defaultTimeoutMax {
		t.Errorf("Chrome.TimeoutMax = %d, want default %d", cfg.Chrome.TimeoutMax, defaultTimeoutMax)
	}
	if cfg.Chrome.ViewportWidth != defaultViewportWidth {
		t.Errorf("Chrome.ViewportWidth = %d, want default %d", cfg.Chrome.ViewportWidth, defaultViewportWidth)
	}
	if cfg.Chrome.ViewportHeight != defaultViewportHeight {
		t.Errorf("Chrome.ViewportHeight = %d, want default %d", cfg.Chrome.ViewportHeight, defaultViewportHeight)
	}
	if cfg.Logging.Level != defaultLogLevel {
		t.Errorf("Logging.Level = %q, want default %q", cfg.Logging.Level, defaultLogLevel)
	}
	if cfg.Logging.Format != defaultLogFormat {
		t.Errorf("Logging.Format = %q, want default %q", cfg.Logging.Format, defaultLogFormat)
	}
}

func TestLoad_EnvironmentOverrides(t *testing.T) {
	content := `
server:
  port: 8080
chrome:
  executable_path: "/usr/bin/chrome"
logging:
  level: "info"
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	// Set environment variables
	os.Setenv("JSBUG_PORT", "9999")
	os.Setenv("JSBUG_CHROME_PATH", "/custom/chrome")
	os.Setenv("JSBUG_LOG_LEVEL", "debug")
	os.Setenv("JSBUG_CORS_ORIGINS", "http://a.com,http://b.com")
	defer func() {
		os.Unsetenv("JSBUG_PORT")
		os.Unsetenv("JSBUG_CHROME_PATH")
		os.Unsetenv("JSBUG_LOG_LEVEL")
		os.Unsetenv("JSBUG_CORS_ORIGINS")
	}()

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 9999 {
		t.Errorf("Server.Port = %d, want %d (from env)", cfg.Server.Port, 9999)
	}
	if cfg.Chrome.ExecutablePath != "/custom/chrome" {
		t.Errorf("Chrome.ExecutablePath = %q, want %q (from env)", cfg.Chrome.ExecutablePath, "/custom/chrome")
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want %q (from env)", cfg.Logging.Level, "debug")
	}
	if len(cfg.Server.CORSOrigins) != 2 || cfg.Server.CORSOrigins[0] != "http://a.com" {
		t.Errorf("Server.CORSOrigins = %v, want [http://a.com, http://b.com]", cfg.Server.CORSOrigins)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"port negative", -1},
		{"port too high", 70000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := createTempConfigWithPort(t, tt.port)
			defer os.Remove(path)

			_, err := Load(path)
			if err == nil {
				t.Errorf("Load() expected error for port %d, got nil", tt.port)
			}
		})
	}
}

func TestLoad_InvalidTimeout(t *testing.T) {
	content := `
server: {}
chrome:
  timeout_default: 100
logging: {}
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	_, err := Load(path)
	if err == nil {
		t.Error("Load() expected error for invalid timeout_default, got nil")
	}
}

func TestLoad_TimeoutMaxLessThanDefault(t *testing.T) {
	content := `
server: {}
chrome:
  timeout_default: 30
  timeout_max: 20
logging: {}
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	_, err := Load(path)
	if err == nil {
		t.Error("Load() expected error when timeout_max < timeout_default, got nil")
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	content := `
server: {}
chrome: {}
logging:
  level: "invalid"
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	_, err := Load(path)
	if err == nil {
		t.Error("Load() expected error for invalid log level, got nil")
	}
}

func TestLoad_InvalidLogFormat(t *testing.T) {
	content := `
server: {}
chrome: {}
logging:
  format: "xml"
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	_, err := Load(path)
	if err == nil {
		t.Error("Load() expected error for invalid log format, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Load() expected error for non-existent file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	content := `
server:
  port: [invalid yaml
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	_, err := Load(path)
	if err == nil {
		t.Error("Load() expected error for invalid YAML, got nil")
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
		},
		Chrome: ChromeConfig{
			TimeoutDefault:    15,
			TimeoutMax:        60,
			ViewportWidth:     1920,
			ViewportHeight:    1080,
			PoolSize:          4,
			WarmupTimeout:     10 * time.Second,
			ShutdownTimeout:   30 * time.Second,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

// Helper functions

func createTempConfig(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	return path
}

func createTempConfigWithPort(t *testing.T, port int) string {
	t.Helper()
	content := `
server:
  port: ` + itoa(port) + `
chrome: {}
logging: {}
`
	return createTempConfig(t, content)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-1"
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}

// Pool configuration tests

func TestLoad_PoolDefaults(t *testing.T) {
	content := `
server: {}
chrome: {}
logging: {}
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Chrome.PoolSize != defaultPoolSize {
		t.Errorf("Chrome.PoolSize = %d, want default %d", cfg.Chrome.PoolSize, defaultPoolSize)
	}
	if cfg.Chrome.WarmupURL != defaultWarmupURL {
		t.Errorf("Chrome.WarmupURL = %q, want default %q", cfg.Chrome.WarmupURL, defaultWarmupURL)
	}
	if cfg.Chrome.WarmupTimeout != defaultWarmupTimeout {
		t.Errorf("Chrome.WarmupTimeout = %v, want default %v", cfg.Chrome.WarmupTimeout, defaultWarmupTimeout)
	}
	if cfg.Chrome.RestartAfterCount != defaultRestartAfterCount {
		t.Errorf("Chrome.RestartAfterCount = %d, want default %d", cfg.Chrome.RestartAfterCount, defaultRestartAfterCount)
	}
	if cfg.Chrome.RestartAfterTime != defaultRestartAfterTime {
		t.Errorf("Chrome.RestartAfterTime = %v, want default %v", cfg.Chrome.RestartAfterTime, defaultRestartAfterTime)
	}
	if cfg.Chrome.ShutdownTimeout != defaultShutdownTimeout {
		t.Errorf("Chrome.ShutdownTimeout = %v, want default %v", cfg.Chrome.ShutdownTimeout, defaultShutdownTimeout)
	}
}

func TestLoad_InvalidPoolSize(t *testing.T) {
	// Note: pool_size=0 in YAML gets default value applied, so it's not invalid
	// The direct Validate() test covers the pool_size=0 case
	tests := []struct {
		name     string
		poolSize int
	}{
		{"pool_size negative", -1},
		{"pool_size too high", 17},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := `
server: {}
chrome:
  pool_size: ` + itoa(tt.poolSize) + `
logging: {}
`
			path := createTempConfig(t, content)
			defer os.Remove(path)

			_, err := Load(path)
			if err == nil {
				t.Errorf("Load() expected error for pool_size %d, got nil", tt.poolSize)
			}
		})
	}
}

func TestLoad_PoolConfigFromYAML(t *testing.T) {
	content := `
server: {}
chrome:
  pool_size: 8
  warmup_url: "https://test.example.com/"
  warmup_timeout: 15s
  restart_after_count: 100
  restart_after_time: 1h
  shutdown_timeout: 45s
logging: {}
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Chrome.PoolSize != 8 {
		t.Errorf("Chrome.PoolSize = %d, want %d", cfg.Chrome.PoolSize, 8)
	}
	if cfg.Chrome.WarmupURL != "https://test.example.com/" {
		t.Errorf("Chrome.WarmupURL = %q, want %q", cfg.Chrome.WarmupURL, "https://test.example.com/")
	}
	if cfg.Chrome.WarmupTimeout != 15*time.Second {
		t.Errorf("Chrome.WarmupTimeout = %v, want %v", cfg.Chrome.WarmupTimeout, 15*time.Second)
	}
	if cfg.Chrome.RestartAfterCount != 100 {
		t.Errorf("Chrome.RestartAfterCount = %d, want %d", cfg.Chrome.RestartAfterCount, 100)
	}
	if cfg.Chrome.RestartAfterTime != 1*time.Hour {
		t.Errorf("Chrome.RestartAfterTime = %v, want %v", cfg.Chrome.RestartAfterTime, 1*time.Hour)
	}
	if cfg.Chrome.ShutdownTimeout != 45*time.Second {
		t.Errorf("Chrome.ShutdownTimeout = %v, want %v", cfg.Chrome.ShutdownTimeout, 45*time.Second)
	}
}

func TestLoad_PoolSizeEnvOverride(t *testing.T) {
	content := `
server: {}
chrome:
  pool_size: 4
logging: {}
`
	path := createTempConfig(t, content)
	defer os.Remove(path)

	os.Setenv("JSBUG_POOL_SIZE", "8")
	defer os.Unsetenv("JSBUG_POOL_SIZE")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Chrome.PoolSize != 8 {
		t.Errorf("Chrome.PoolSize = %d, want %d (from env)", cfg.Chrome.PoolSize, 8)
	}
}

func TestValidate_ValidPoolConfig(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
		},
		Chrome: ChromeConfig{
			TimeoutDefault:    15,
			TimeoutMax:        60,
			ViewportWidth:     1920,
			ViewportHeight:    1080,
			PoolSize:          4,
			WarmupURL:         "https://example.com/",
			WarmupTimeout:     10 * time.Second,
			RestartAfterCount: 50,
			RestartAfterTime:  30 * time.Minute,
			ShutdownTimeout:   30 * time.Second,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestValidate_InvalidPoolSize(t *testing.T) {
	tests := []struct {
		name     string
		poolSize int
	}{
		{"zero", 0},
		{"negative", -1},
		{"too_high", 17},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{Port: 8080},
				Chrome: ChromeConfig{
					TimeoutDefault:    15,
					TimeoutMax:        60,
					ViewportWidth:     1920,
					ViewportHeight:    1080,
					PoolSize:          tt.poolSize,
					WarmupTimeout:     10 * time.Second,
					ShutdownTimeout:   30 * time.Second,
				},
				Logging: LoggingConfig{Level: "info", Format: "json"},
			}

			err := cfg.Validate()
			if err == nil {
				t.Errorf("Validate() expected error for pool_size %d, got nil", tt.poolSize)
			}
		})
	}
}
