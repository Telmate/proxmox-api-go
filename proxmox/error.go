package proxmox

import (
	"errors"
	"fmt"
	"strings"
)

// Error serves as a pseudo-namespace for error message generators.
var Error = errorMsg{}

type errorMsg struct{}

// Generic contextual wrapper
type errorWrapper[T errorContext] struct {
	err error
	id  T
}

func (w *errorWrapper[T]) Error() string { return w.err.Error() + ": " + w.id.errorContext() }

func (w *errorWrapper[T]) Unwrap() error { return w.err }

type errorWrap struct {
	err     error
	message string
}

func (w *errorWrap) Error() string { return w.message + ": " + w.err.Error() }

func (w *errorWrap) Unwrap() error { return w.err }

type errorContext interface{ errorContext() string }

var errGuestDoesNotExist = errors.New("guest does not exist")

func (msg errorMsg) GuestDoesNotExist() error { return errGuestDoesNotExist }

func (errorMsg) guestDoesNotExist(id GuestID) error {
	return &errorWrapper[GuestID]{
		err: Error.GuestDoesNotExist(),
		id:  id}
}

var errGuestIsProtectedCantDelete = errors.New("cannot delete guest because it is protected")

func (msg errorMsg) GuestIsProtectedCantDelete() error { return errGuestIsProtectedCantDelete }

func (errorMsg) guestIsProtectedCantDelete(id GuestID) error {
	return &errorWrapper[GuestID]{
		err: Error.GuestIsProtectedCantDelete(),
		id:  id}
}

type ApiError struct {
	Errors  map[string]any
	Message string
	Code    string
}

func (e ApiError) Error() string {
	builder := strings.Builder{}

	builder.WriteString("api error: code: ")
	builder.WriteString(e.Code)
	builder.WriteString(" message: ")
	builder.WriteString(e.Message)

	if len(e.Errors) != 0 {
		builder.WriteString(" details: ( ")

		for k, v := range e.Errors {
			builder.WriteString(k)
			builder.WriteString(":")
			fmt.Fprintf(&builder, "%v", v)
			builder.WriteString(" | ")
		}
		builder.WriteString(" )")
	}

	return builder.String()
}

type TaskError struct {
	TaskID  string
	Message string
}

func (e TaskError) Error() string {
	return "task error: task id: " + e.TaskID + " message: " + e.Message
}

type ValidateError struct {
	Message string
}

func newValidateError(message string) error { return &ValidateError{Message: message} }

func (e *ValidateError) Error() string {
	return "validation error: " + e.Message
}

type VersionError struct {
	err           error
	Functionality string
	Version       Version
}

var errNotSupportedInVersion = errors.New("functionality not supported in version")

func (msg errorMsg) VersionError() error { return errNotSupportedInVersion }

func newVersionError(functionality string, version Version) error {
	return &VersionError{
		err:           errNotSupportedInVersion,
		Functionality: functionality,
		Version:       version}
}

func (e *VersionError) Error() string {
	return "version error: functionality (" + e.Functionality + ") is only supported on Proxmox VE (" + e.Version.String() + ") and higher"
}

func (e *VersionError) Unwrap() error { return e.err }
