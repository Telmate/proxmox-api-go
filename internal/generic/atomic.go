package generic

import (
	"sync"

	"github.com/Telmate/proxmox-api-go/internal/noCopy"
)

type Atomic[T any] struct {
	_     noCopy.True
	value T
	mutex sync.RWMutex
}

// NewAtomic initializes and returns a new *Atomic[T].
func NewAtomic[T any]() *Atomic[T] { return new(Atomic[T]) }

// Store atomically stores value in x.
func (x *Atomic[T]) Store(value T) {
	x.mutex.Lock()
	x.value = value
	x.mutex.Unlock()
}

// Load atomically loads and returns the value stored in x.
func (x *Atomic[T]) Load() T {
	x.mutex.RLock()
	value := x.value
	x.mutex.RUnlock()
	return value
}
