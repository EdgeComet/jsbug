package chrome

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestPoolErrors(t *testing.T) {
	// Test that error variables are distinct
	if ErrNoInstanceAvailable == ErrPoolShuttingDown {
		t.Error("ErrNoInstanceAvailable and ErrPoolShuttingDown should be distinct")
	}

	if ErrNoInstanceAvailable.Error() != "no chrome instance available" {
		t.Errorf("ErrNoInstanceAvailable.Error() = %q, want %q",
			ErrNoInstanceAvailable.Error(), "no chrome instance available")
	}

	if ErrPoolShuttingDown.Error() != "pool is shutting down" {
		t.Errorf("ErrPoolShuttingDown.Error() = %q, want %q",
			ErrPoolShuttingDown.Error(), "pool is shutting down")
	}
}

func TestPoolStats_Initial(t *testing.T) {
	// Test PoolStats struct
	stats := PoolStats{
		TotalInstances:     4,
		AvailableInstances: 3,
		ActiveInstances:    1,
	}

	if stats.TotalInstances != 4 {
		t.Errorf("TotalInstances = %d, want 4", stats.TotalInstances)
	}
	if stats.AvailableInstances != 3 {
		t.Errorf("AvailableInstances = %d, want 3", stats.AvailableInstances)
	}
	if stats.ActiveInstances != 1 {
		t.Errorf("ActiveInstances = %d, want 1", stats.ActiveInstances)
	}
}

func TestAcquire_PoolExhausted(t *testing.T) {
	// Create a pool structure manually without real Chrome instances
	// to test the exhaustion logic
	logger := zap.NewNop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := &ChromePool{
		config:    InstanceConfig{PoolSize: 2},
		logger:    logger,
		instances: make([]*Instance, 2),
		available: make(chan int, 2),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Don't add any instances to available channel - simulates exhausted pool
	// Acquire should return ErrNoInstanceAvailable immediately
	_, err := pool.Acquire()
	if !errors.Is(err, ErrNoInstanceAvailable) {
		t.Errorf("Acquire() error = %v, want ErrNoInstanceAvailable", err)
	}
}

func TestAcquire_ShuttingDown(t *testing.T) {
	logger := zap.NewNop()
	ctx, cancel := context.WithCancel(context.Background())

	pool := &ChromePool{
		config:    InstanceConfig{PoolSize: 2},
		logger:    logger,
		instances: make([]*Instance, 2),
		available: make(chan int, 2),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Add an instance to available
	pool.available <- 0

	// Cancel context to simulate shutdown
	cancel()

	// Acquire should return ErrPoolShuttingDown
	_, err := pool.Acquire()
	if !errors.Is(err, ErrPoolShuttingDown) {
		t.Errorf("Acquire() error = %v, want ErrPoolShuttingDown", err)
	}
}

func TestActiveCount_IncrementDecrement(t *testing.T) {
	logger := zap.NewNop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := &ChromePool{
		config:    InstanceConfig{PoolSize: 2},
		logger:    logger,
		instances: make([]*Instance, 2),
		available: make(chan int, 2),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initial active count should be 0
	if pool.activeCount.Load() != 0 {
		t.Errorf("initial activeCount = %d, want 0", pool.activeCount.Load())
	}

	// Simulate increment
	pool.activeCount.Add(1)
	if pool.activeCount.Load() != 1 {
		t.Errorf("activeCount after Add(1) = %d, want 1", pool.activeCount.Load())
	}

	// Simulate decrement
	pool.activeCount.Add(-1)
	if pool.activeCount.Load() != 0 {
		t.Errorf("activeCount after Add(-1) = %d, want 0", pool.activeCount.Load())
	}
}

func TestConcurrentActiveCount(t *testing.T) {
	logger := zap.NewNop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := &ChromePool{
		config:    InstanceConfig{PoolSize: 4},
		logger:    logger,
		instances: make([]*Instance, 4),
		available: make(chan int, 4),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Test concurrent increments and decrements
	var wg sync.WaitGroup
	iterations := 1000

	// Spawn goroutines that increment
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pool.activeCount.Add(1)
		}()
	}

	// Spawn goroutines that decrement
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pool.activeCount.Add(-1)
		}()
	}

	wg.Wait()

	// After equal increments and decrements, count should be 0
	if pool.activeCount.Load() != 0 {
		t.Errorf("activeCount after concurrent ops = %d, want 0", pool.activeCount.Load())
	}
}

func TestShutdown_WaitsForActiveRenders(t *testing.T) {
	logger := zap.NewNop()
	ctx, cancel := context.WithCancel(context.Background())

	pool := &ChromePool{
		config:    InstanceConfig{PoolSize: 2},
		logger:    logger,
		instances: make([]*Instance, 2),
		available: make(chan int, 2),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Simulate active render
	pool.activeCount.Add(1)

	// Start shutdown in goroutine
	shutdownDone := make(chan struct{})
	go func() {
		pool.Shutdown()
		close(shutdownDone)
	}()

	// Give shutdown time to start polling
	time.Sleep(100 * time.Millisecond)

	// Simulate render completion
	pool.activeCount.Add(-1)

	// Shutdown should complete quickly after render is done
	select {
	case <-shutdownDone:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Shutdown did not complete after active render finished")
	}
}

func TestShutdown_CompletesWithNoActiveRenders(t *testing.T) {
	logger := zap.NewNop()
	ctx, cancel := context.WithCancel(context.Background())

	pool := &ChromePool{
		config:    InstanceConfig{PoolSize: 2},
		logger:    logger,
		instances: make([]*Instance, 2),
		available: make(chan int, 2),
		ctx:       ctx,
		cancel:    cancel,
	}

	// No active renders - shutdown should complete immediately
	start := time.Now()
	pool.Shutdown()
	elapsed := time.Since(start)

	// Shutdown should complete quickly when no active renders
	if elapsed > 1*time.Second {
		t.Errorf("Shutdown took too long: %v (expected < 1s)", elapsed)
	}
}

func TestShutdown_TerminatesAllInstances(t *testing.T) {
	logger := zap.NewNop()
	ctx, cancel := context.WithCancel(context.Background())

	// Create mock instances that track termination
	instances := make([]*Instance, 2)
	for i := 0; i < 2; i++ {
		instances[i] = &Instance{
			id:     i,
			logger: logger,
		}
		instances[i].status.Store(int32(StatusIdle))
	}

	pool := &ChromePool{
		config:    InstanceConfig{PoolSize: 2},
		logger:    logger,
		instances: instances,
		available: make(chan int, 2),
		ctx:       ctx,
		cancel:    cancel,
	}

	pool.Shutdown()

	// Verify all instances were terminated (status set to Dead)
	for i, instance := range instances {
		if instance.Status() != StatusDead {
			t.Errorf("Instance %d status = %v, want StatusDead", i, instance.Status())
		}
	}
}

func TestAcquire_ReturnsErrorAfterShutdown(t *testing.T) {
	logger := zap.NewNop()
	ctx, cancel := context.WithCancel(context.Background())

	pool := &ChromePool{
		config:    InstanceConfig{PoolSize: 2},
		logger:    logger,
		instances: make([]*Instance, 2),
		available: make(chan int, 2),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Add available instance
	pool.available <- 0

	// Shutdown the pool
	pool.Shutdown()

	// Acquire should return error
	_, err := pool.Acquire()
	if !errors.Is(err, ErrPoolShuttingDown) {
		t.Errorf("Acquire() after shutdown error = %v, want ErrPoolShuttingDown", err)
	}
}
