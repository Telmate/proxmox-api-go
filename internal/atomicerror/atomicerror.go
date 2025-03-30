package atomicerror

import (
	"sync"
)

// AtomicError is a structure that allows for safe error handling in concurrent environments.
type AtomicError struct {
	once  sync.Once
	mutex sync.RWMutex
	value error
}

// New initializes and returns a new AtomicError.
func New() *AtomicError {
	return &AtomicError{}
}

// Set sets the error value of the AtomicError.
func (ae *AtomicError) Set(err error) {
	ae.once.Do(func() {
		ae.mutex.Lock()
		ae.value = err
		ae.mutex.Unlock()
	})
}

// Get returns the error value of the AtomicError.
func (ae *AtomicError) Get() error {
	ae.mutex.RLock()
	defer ae.mutex.RUnlock()
	return ae.value
}
