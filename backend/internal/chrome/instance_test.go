//go:build chrome

package chrome

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestConfig() InstanceConfig {
	return InstanceConfig{
		Headless:  true,
		NoSandbox: false,
	}
}

func TestNew_Success(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	if instance == nil {
		t.Fatal("New() returned nil instance")
	}

	if instance.Status() != StatusIdle {
		t.Errorf("initial status = %v, want %v", instance.Status(), StatusIdle)
	}

	if instance.ID() != 0 {
		t.Errorf("ID() = %d, want 0", instance.ID())
	}
}

func TestInstance_GetContext(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	tabCtx, cancel := instance.GetContext()
	defer cancel()

	if tabCtx == nil {
		t.Error("GetContext() returned nil context")
	}

	if instance.Status() != StatusIdle {
		t.Errorf("status after GetContext = %v, want %v", instance.Status(), StatusIdle)
	}
}

func TestInstance_IsAlive(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if !instance.IsAlive() {
		t.Error("IsAlive() = false, want true for new instance")
	}

	instance.Close()

	if instance.IsAlive() {
		t.Error("IsAlive() = true, want false after Close")
	}
}

func TestInstance_IsAlive_StatusDead(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	// Set status to Dead
	instance.SetStatus(StatusDead)

	if instance.IsAlive() {
		t.Error("IsAlive() = true, want false when status is StatusDead")
	}
}

func TestInstance_IsAlive_NilBrowserCtx(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Manually set browserCtx to nil (simulating a crashed browser)
	instance.mu.Lock()
	instance.browserCtx = nil
	instance.mu.Unlock()

	if instance.IsAlive() {
		t.Error("IsAlive() = true, want false when browserCtx is nil")
	}

	// Clean up - set browserCtx back so Close() doesn't panic
	instance.SetStatus(StatusClosed)
}

func TestInstance_Close(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = instance.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if instance.Status() != StatusClosed {
		t.Errorf("status after Close = %v, want %v", instance.Status(), StatusClosed)
	}

	// Second close should be idempotent
	err = instance.Close()
	if err != nil {
		t.Errorf("second Close() error = %v", err)
	}
}

func TestInstance_SetStatus(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	instance.SetStatus(StatusRendering)
	if instance.Status() != StatusRendering {
		t.Errorf("Status() = %v, want %v", instance.Status(), StatusRendering)
	}

	instance.SetStatus(StatusIdle)
	if instance.Status() != StatusIdle {
		t.Errorf("Status() = %v, want %v", instance.Status(), StatusIdle)
	}
}

func TestInstance_RenderCount(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	// Initial render count should be 0
	if instance.RenderCount() != 0 {
		t.Errorf("initial RenderCount() = %d, want 0", instance.RenderCount())
	}

	// Increment render count
	instance.IncrementRenders()
	if instance.RenderCount() != 1 {
		t.Errorf("RenderCount() after increment = %d, want 1", instance.RenderCount())
	}

	// Increment again
	instance.IncrementRenders()
	instance.IncrementRenders()
	if instance.RenderCount() != 3 {
		t.Errorf("RenderCount() after 3 increments = %d, want 3", instance.RenderCount())
	}
}

func TestInstance_CreatedAt(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	before := time.Now()
	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()
	after := time.Now()

	createdAt := instance.CreatedAt()
	if createdAt.Before(before) || createdAt.After(after) {
		t.Errorf("CreatedAt() = %v, want between %v and %v", createdAt, before, after)
	}
}

func TestInstance_ResetCounters(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	// Increment some renders
	instance.IncrementRenders()
	instance.IncrementRenders()
	instance.IncrementRenders()

	if instance.RenderCount() != 3 {
		t.Fatalf("RenderCount() = %d, want 3", instance.RenderCount())
	}

	originalCreatedAt := instance.CreatedAt()
	time.Sleep(10 * time.Millisecond) // Small delay to ensure time difference

	// Reset counters
	instance.resetCounters()

	if instance.RenderCount() != 0 {
		t.Errorf("RenderCount() after reset = %d, want 0", instance.RenderCount())
	}

	newCreatedAt := instance.CreatedAt()
	if !newCreatedAt.After(originalCreatedAt) {
		t.Errorf("CreatedAt() after reset = %v, should be after %v", newCreatedAt, originalCreatedAt)
	}
}

func TestInstance_ID(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(42, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	if instance.ID() != 42 {
		t.Errorf("ID() = %d, want 42", instance.ID())
	}
}

func TestInstance_ShouldRestart_CountExceeded(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()
	cfg.RestartAfterCount = 3
	cfg.RestartAfterTime = 0 // Disable time-based restart

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	// Initially should not need restart
	if instance.ShouldRestart() {
		t.Error("ShouldRestart() = true, want false when count is 0")
	}

	// Increment to threshold
	instance.IncrementRenders()
	instance.IncrementRenders()
	instance.IncrementRenders()

	if !instance.ShouldRestart() {
		t.Error("ShouldRestart() = false, want true when count >= RestartAfterCount")
	}
}

func TestInstance_ShouldRestart_TimeExceeded(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()
	cfg.RestartAfterCount = 0 // Disable count-based restart
	cfg.RestartAfterTime = 10 * time.Millisecond

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	// Initially should not need restart
	if instance.ShouldRestart() {
		t.Error("ShouldRestart() = true, want false initially")
	}

	// Wait for time to exceed
	time.Sleep(15 * time.Millisecond)

	if !instance.ShouldRestart() {
		t.Error("ShouldRestart() = false, want true when time >= RestartAfterTime")
	}
}

func TestInstance_ShouldRestart_NeitherExceeded(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()
	cfg.RestartAfterCount = 100
	cfg.RestartAfterTime = 1 * time.Hour

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	// Increment a bit but not to threshold
	instance.IncrementRenders()
	instance.IncrementRenders()

	if instance.ShouldRestart() {
		t.Error("ShouldRestart() = true, want false when neither threshold exceeded")
	}
}

func TestInstance_Terminate(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = instance.Terminate()
	if err != nil {
		t.Errorf("Terminate() error = %v", err)
	}

	if instance.Status() != StatusDead {
		t.Errorf("status after Terminate = %v, want %v", instance.Status(), StatusDead)
	}

	// IsAlive should return false
	if instance.IsAlive() {
		t.Error("IsAlive() = true, want false after Terminate")
	}
}

func TestInstance_Restart_Success(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(0, cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	// Increment render count to verify it gets reset
	instance.IncrementRenders()
	instance.IncrementRenders()
	instance.IncrementRenders()

	if instance.RenderCount() != 3 {
		t.Fatalf("RenderCount() before restart = %d, want 3", instance.RenderCount())
	}

	originalCreatedAt := instance.CreatedAt()
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	// Restart the instance
	err = instance.Restart()
	if err != nil {
		t.Fatalf("Restart() error = %v", err)
	}

	// Verify status is idle after restart
	if instance.Status() != StatusIdle {
		t.Errorf("status after Restart = %v, want %v", instance.Status(), StatusIdle)
	}

	// Verify instance is still alive
	if !instance.IsAlive() {
		t.Error("IsAlive() = false, want true after successful Restart")
	}

	// Verify render count was reset
	if instance.RenderCount() != 0 {
		t.Errorf("RenderCount() after restart = %d, want 0", instance.RenderCount())
	}

	// Verify createdAt was reset
	newCreatedAt := instance.CreatedAt()
	if !newCreatedAt.After(originalCreatedAt) {
		t.Errorf("CreatedAt() after restart = %v, should be after %v", newCreatedAt, originalCreatedAt)
	}
}
