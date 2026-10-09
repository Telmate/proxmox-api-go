package atomicerror

import (
	"sync"
)

// AtomicError is a structure that allows for safe error handling in concurrent environments.
type AtomicError struct {
	value error
	mutex sync.RWMutex
	once  sync.Once
}

// New initializes and returns a new AtomicError.
func New() *AtomicError { return &AtomicError{} }

// Store stores the error value of the AtomicError.
func (ae *AtomicError) Store(err error) {
	ae.once.Do(func() {
		ae.mutex.Lock()
		ae.value = err
		ae.mutex.Unlock()
	})
}

// Load returns the error value of the AtomicError.
func (ae *AtomicError) Load() error {
	ae.mutex.RLock()
	err := ae.value
	ae.mutex.RUnlock()
	return err
}
