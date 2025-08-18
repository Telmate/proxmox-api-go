package proxmox

import (
	"errors"
)

// Error serves as a pseudo-namespace for error message generators.
var Error = errorMsg{}

type errorMsg struct {
}

// Generic contextual wrapper
type errorWrapper[T errorContext] struct {
	err error
	id  T
}

func (w *errorWrapper[T]) Error() string {
	return w.err.Error() + ": " + w.id.errorContext()
}

type errorContext interface {
	errorContext() string
}

var errGuestDoesNotExist = errors.New("guest does not exist")

func (msg errorMsg) GuestDoesNotExist() error {
	return errGuestDoesNotExist
}

func (errorMsg) guestDoesNotExist(id GuestID) error {
	return &errorWrapper[GuestID]{
		err: Error.GuestDoesNotExist(),
		id:  id}
}
