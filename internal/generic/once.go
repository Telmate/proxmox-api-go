package generic

import (
	"sync"

	"github.com/Telmate/proxmox-api-go/internal/noCopy"
)

type Once[T any] struct {
	_     noCopy.True
	value T
	mutex sync.RWMutex
	once  sync.Once
}

// NewOnce initializes and returns a new *Once[T].
func NewOnce[T any]() *Once[T] { return new(Once[T]) }

// Set atomically sets the value in x, it cannot be changed afterwards.
func (x *Once[T]) Set(value T) {
	x.once.Do(func() {
		x.mutex.Lock()
		x.value = value
		x.mutex.Unlock()
	})
}

// Load atomically loads and returns the value stored in x.
func (x *Once[T]) Load() T {
	x.mutex.RLock()
	value := x.value
	x.mutex.RUnlock()
	return value
}
