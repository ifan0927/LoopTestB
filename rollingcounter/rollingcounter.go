// Package rollingcounter provides a concurrency-safe fixed-width counter.
package rollingcounter

import (
	"errors"
	"sync"
)

var (
	// ErrInvalidWindowCount is returned when a counter is created without windows.
	ErrInvalidWindowCount = errors.New("rollingcounter: window count must be positive")
	// ErrInvalidIndex is returned when an operation targets a window outside the counter.
	ErrInvalidIndex = errors.New("rollingcounter: window index out of range")
)

// Counter stores a fixed number of independently incremented windows.
type Counter struct {
	mu      sync.RWMutex
	windows []int64
}

// NewCounter creates a counter with windowCount windows.
func NewCounter(windowCount int) (*Counter, error) {
	if windowCount <= 0 {
		return nil, ErrInvalidWindowCount
	}

	return &Counter{windows: make([]int64, windowCount)}, nil
}

// Add adds delta to the window at index. Invalid indexes leave the counter unchanged.
func (c *Counter) Add(index int, delta int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if index < 0 || index >= len(c.windows) {
		return ErrInvalidIndex
	}

	c.windows[index] += delta
	return nil
}

// Total returns the sum of all windows.
func (c *Counter) Total() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var total int64
	for _, value := range c.windows {
		total += value
	}
	return total
}

// Snapshot returns a copy of the counter's current windows.
func (c *Counter) Snapshot() []int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	snapshot := make([]int64, len(c.windows))
	copy(snapshot, c.windows)
	return snapshot
}
