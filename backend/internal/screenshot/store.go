package screenshot

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DefaultTTL is the default time-to-live for screenshots
const DefaultTTL = 5 * time.Minute

// entry holds screenshot data with creation timestamp
type entry struct {
	data      []byte
	createdAt time.Time
}

// ScreenshotStore provides thread-safe in-memory storage for screenshots with TTL
type ScreenshotStore struct {
	mu      sync.RWMutex
	entries map[string]*entry
	ttl     time.Duration
}

// NewStore creates a new ScreenshotStore with the specified TTL
// If ttl is 0, DefaultTTL is used
func NewStore(ttl time.Duration) *ScreenshotStore {
	if ttl == 0 {
		ttl = DefaultTTL
	}
	return &ScreenshotStore{
		entries: make(map[string]*entry),
		ttl:     ttl,
	}
}

// Store saves screenshot data and returns a unique ID
func (s *ScreenshotStore) Store(data []byte) string {
	id := uuid.New().String()
	e := &entry{
		data:      data,
		createdAt: time.Now(),
	}

	s.mu.Lock()
	s.entries[id] = e
	s.mu.Unlock()

	return id
}

// Get retrieves screenshot data by ID
// Returns nil, false if not found or expired
func (s *ScreenshotStore) Get(id string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, exists := s.entries[id]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Since(e.createdAt) > s.ttl {
		delete(s.entries, id)
		return nil, false
	}

	return e.data, true
}

// cleanup removes all expired entries
func (s *ScreenshotStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, e := range s.entries {
		if time.Since(e.createdAt) > s.ttl {
			delete(s.entries, id)
		}
	}
}

// StartCleanup starts a background goroutine that periodically removes expired entries
// The cleanup stops when ctx is cancelled
func (s *ScreenshotStore) StartCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.cleanup()
			}
		}
	}()
}
