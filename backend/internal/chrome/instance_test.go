//go:build chrome

package chrome

import (
	"testing"

	"go.uber.org/zap"
)

func newTestConfig() InstanceConfig {
	return InstanceConfig{
		Headless:       true,
		DisableGPU:     true,
		NoSandbox:      false,
		ViewportWidth:  1920,
		ViewportHeight: 1080,
	}
}

func TestNew_Success(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(cfg, logger)
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
}

func TestInstance_GetContext(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	tabCtx, cancel := instance.GetContext()
	defer cancel()

	if tabCtx == nil {
		t.Error("GetContext() returned nil context")
	}

	if instance.Status() != StatusBusy {
		t.Errorf("status after GetContext = %v, want %v", instance.Status(), StatusBusy)
	}
}

func TestInstance_IsAlive(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(cfg, logger)
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

func TestInstance_Close(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestConfig()

	instance, err := New(cfg, logger)
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

	instance, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer instance.Close()

	instance.SetStatus(StatusBusy)
	if instance.Status() != StatusBusy {
		t.Errorf("Status() = %v, want %v", instance.Status(), StatusBusy)
	}

	instance.SetStatus(StatusIdle)
	if instance.Status() != StatusIdle {
		t.Errorf("Status() = %v, want %v", instance.Status(), StatusIdle)
	}
}
