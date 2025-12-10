package chrome

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// Pool errors
var (
	ErrNoInstanceAvailable = errors.New("no chrome instance available")
	ErrPoolShuttingDown    = errors.New("pool is shutting down")
)

// PoolStats contains pool statistics
type PoolStats struct {
	TotalInstances     int
	AvailableInstances int
	ActiveInstances    int32
}

// ChromePool manages a pool of Chrome browser instances
type ChromePool struct {
	config      InstanceConfig
	logger      *zap.Logger
	instances   []*Instance
	available   chan int // buffered channel of instance IDs
	activeCount atomic.Int32
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewChromePool creates a new pool of Chrome instances.
// It initializes all instances sequentially and fails fast if any instance fails.
func NewChromePool(config InstanceConfig, logger *zap.Logger) (*ChromePool, error) {
	if config.PoolSize <= 0 {
		config.PoolSize = 1
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &ChromePool{
		config:    config,
		logger:    logger,
		instances: make([]*Instance, config.PoolSize),
		available: make(chan int, config.PoolSize),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initialize all instances sequentially
	for i := 0; i < config.PoolSize; i++ {
		instance, err := New(i, config, logger)
		if err != nil {
			// Fail-fast: terminate all already-created instances
			logger.Error("Failed to create Chrome instance, terminating pool",
				zap.Int("instance_id", i),
				zap.Error(err),
			)
			for j := 0; j < i; j++ {
				if pool.instances[j] != nil {
					pool.instances[j].Terminate()
				}
			}
			cancel()
			return nil, err
		}

		pool.instances[i] = instance
		pool.available <- i

		logger.Debug("Chrome instance created",
			zap.Int("instance_id", i),
		)
	}

	logger.Info("Chrome pool initialized",
		zap.Int("pool_size", config.PoolSize),
	)

	return pool, nil
}

// Stats returns current pool statistics
func (p *ChromePool) Stats() PoolStats {
	return PoolStats{
		TotalInstances:     len(p.instances),
		AvailableInstances: len(p.available),
		ActiveInstances:    p.activeCount.Load(),
	}
}

// Acquire gets an available Chrome instance from the pool.
// It returns ErrNoInstanceAvailable if all instances are busy,
// or ErrPoolShuttingDown if the pool is shutting down.
func (p *ChromePool) Acquire() (*Instance, error) {
	// Fast-path: check if pool is shutting down
	select {
	case <-p.ctx.Done():
		return nil, ErrPoolShuttingDown
	default:
	}

	// Non-blocking receive from available channel
	select {
	case id := <-p.available:
		// Double-check shutdown to prevent race condition
		select {
		case <-p.ctx.Done():
			// Return instance to queue and report shutdown
			select {
			case p.available <- id:
			default:
			}
			return nil, ErrPoolShuttingDown
		default:
		}

		instance := p.instances[id]

		// Check if instance is alive
		if !instance.IsAlive() {
			// Attempt restart
			if err := instance.Restart(); err != nil {
				// Restart failed, return instance to queue
				p.logger.Error("Failed to restart dead instance",
					zap.Int("instance_id", id),
					zap.Error(err),
				)
				select {
				case p.available <- id:
				default:
				}
				return nil, err
			}
			p.logger.Info("Restarted dead instance",
				zap.Int("instance_id", id),
			)
		}

		// Check if policy-based restart is needed
		if instance.ShouldRestart() {
			if err := instance.Restart(); err != nil {
				p.logger.Warn("Policy restart failed, continuing with existing instance",
					zap.Int("instance_id", id),
					zap.Error(err),
				)
			} else {
				p.logger.Debug("Policy restart completed",
					zap.Int("instance_id", id),
				)
			}
		}

		// Increment active count and set status
		p.activeCount.Add(1)
		instance.SetStatus(StatusRendering)

		p.logger.Debug("Instance acquired",
			zap.Int("instance_id", id),
			zap.Int32("active_count", p.activeCount.Load()),
		)

		return instance, nil

	default:
		// No instance available
		return nil, ErrNoInstanceAvailable
	}
}

// Release returns an instance back to the pool.
// It should be called when done using an acquired instance.
func (p *ChromePool) Release(instance *Instance) {
	if instance == nil {
		return
	}

	// Decrement active count BEFORE returning to queue (important for shutdown)
	p.activeCount.Add(-1)

	// Set status to idle and increment render count
	instance.SetStatus(StatusIdle)
	instance.IncrementRenders()

	// Try to return instance to available queue
	select {
	case p.available <- instance.ID():
		p.logger.Debug("Instance released",
			zap.Int("instance_id", instance.ID()),
			zap.Int32("active_count", p.activeCount.Load()),
		)
	case <-p.ctx.Done():
		p.logger.Debug("Discarding instance during shutdown",
			zap.Int("instance_id", instance.ID()),
		)
	default:
		p.logger.Error("Available queue full - possible double release",
			zap.Int("instance_id", instance.ID()),
		)
	}
}

// NewMockPool creates a mock pool for testing that returns the specified error on Acquire.
// This is useful for testing error handling without needing real Chrome instances.
func NewMockPool(logger *zap.Logger, acquireErr error) *ChromePool {
	ctx, cancel := context.WithCancel(context.Background())

	// If simulating shutdown, cancel immediately
	if acquireErr == ErrPoolShuttingDown {
		cancel()
	}

	return &ChromePool{
		config:    InstanceConfig{PoolSize: 1, ShutdownTimeout: time.Second},
		logger:    logger,
		instances: make([]*Instance, 0),
		available: make(chan int, 1), // Empty channel - no instances available
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Shutdown gracefully shuts down the pool.
// It waits for active renders to complete up to the configured timeout,
// then terminates all instances.
func (p *ChromePool) Shutdown() error {
	// Signal shutdown - prevents new acquires
	p.cancel()

	activeCount := p.activeCount.Load()
	p.logger.Info("Pool shutdown started",
		zap.Int32("active_renders", activeCount),
	)

	// Calculate deadline from config.ShutdownTimeout
	deadline := time.Now().Add(p.config.ShutdownTimeout)

	// Poll loop: wait for active renders to complete
	for {
		if p.activeCount.Load() == 0 {
			p.logger.Info("All renders completed gracefully")
			break
		}

		if time.Now().After(deadline) {
			p.logger.Warn("Shutdown timeout exceeded, forcing termination",
				zap.Int32("active_renders", p.activeCount.Load()),
			)
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	// Terminate all instances
	for i, instance := range p.instances {
		if instance != nil {
			if err := instance.Terminate(); err != nil {
				p.logger.Error("Failed to terminate instance",
					zap.Int("instance_id", i),
					zap.Error(err),
				)
			}
		}
	}

	p.logger.Info("Pool shutdown complete")
	return nil
}
