package rollingcounter

import (
	"errors"
	"sync"
	"testing"
)

func TestNewCounter(t *testing.T) {
	tests := []struct {
		name        string
		windowCount int
		wantErr     error
	}{
		{name: "zero windows", windowCount: 0, wantErr: ErrInvalidWindowCount},
		{name: "negative windows", windowCount: -1, wantErr: ErrInvalidWindowCount},
		{name: "one window", windowCount: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter, err := NewCounter(tt.windowCount)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewCounter(%d) error = %v, want %v", tt.windowCount, err, tt.wantErr)
			}
			if err == nil && len(counter.Snapshot()) != tt.windowCount {
				t.Fatalf("counter has %d windows, want %d", len(counter.Snapshot()), tt.windowCount)
			}
		})
	}
}

func TestCounterAddTotalAndSnapshot(t *testing.T) {
	counter, err := NewCounter(3)
	if err != nil {
		t.Fatalf("NewCounter() error = %v", err)
	}

	tests := []struct {
		name      string
		index     int
		delta     int64
		wantErr   error
		wantTotal int64
		want      []int64
	}{
		{name: "first window", index: 0, delta: 2, wantTotal: 2, want: []int64{2, 0, 0}},
		{name: "last window", index: 2, delta: 5, wantTotal: 7, want: []int64{2, 0, 5}},
		{name: "negative delta", index: 0, delta: -1, wantTotal: 6, want: []int64{1, 0, 5}},
		{name: "negative index", index: -1, delta: 10, wantErr: ErrInvalidIndex, wantTotal: 6, want: []int64{1, 0, 5}},
		{name: "index after range", index: 3, delta: 10, wantErr: ErrInvalidIndex, wantTotal: 6, want: []int64{1, 0, 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := counter.Add(tt.index, tt.delta); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Add(%d, %d) error = %v, want %v", tt.index, tt.delta, err, tt.wantErr)
			}
			if got := counter.Total(); got != tt.wantTotal {
				t.Fatalf("Total() = %d, want %d", got, tt.wantTotal)
			}
			if got := counter.Snapshot(); !equal(got, tt.want) {
				t.Fatalf("Snapshot() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("snapshot is defensive", func(t *testing.T) {
		snapshot := counter.Snapshot()
		snapshot[0] = 99
		if got := counter.Snapshot()[0]; got != 1 {
			t.Fatalf("counter changed through snapshot: got %d, want 1", got)
		}
	})
}

func TestCounterConcurrentUse(t *testing.T) {
	const (
		windows    = 4
		goroutines = 32
		increments = 1_000
	)

	counter, err := NewCounter(windows)
	if err != nil {
		t.Fatalf("NewCounter() error = %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for j := 0; j < increments; j++ {
				if err := counter.Add(index%windows, 1); err != nil {
					t.Errorf("Add() error = %v", err)
					return
				}
			}
		}(i)
	}
	wg.Wait()

	if got, want := counter.Total(), int64(goroutines*increments); got != want {
		t.Fatalf("Total() = %d, want %d", got, want)
	}
}

func equal(left, right []int64) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
