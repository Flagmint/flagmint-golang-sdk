// Package syncutil provides small synchronisation helpers used internally.
package syncutil

import "sync"

// RWValue is a generic value protected by a [sync.RWMutex].
// It is safe for concurrent use.
type RWValue[T any] struct {
	mu  sync.RWMutex
	val T
}

// Load returns the current value.
func (r *RWValue[T]) Load() T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.val
}

// Store replaces the current value.
func (r *RWValue[T]) Store(v T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.val = v
}

// Swap atomically replaces the value and returns the previous one.
func (r *RWValue[T]) Swap(v T) T {
	r.mu.Lock()
	defer r.mu.Unlock()
	prev := r.val
	r.val = v
	return prev
}
