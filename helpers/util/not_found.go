package util //nolint:revive

import (
	"errors"
	"strings"
)

// ErrResourceNotFound is returned when a remote object no longer exists (e.g. HTTP 404).
// Resource Read methods may treat this as a signal to remove the instance from state.
var ErrResourceNotFound = errors.New("resource not found")

// IsResourceNotFound reports whether err indicates the remote resource is absent.
// It matches wrapped ErrResourceNotFound and legacy error strings from older call paths.
func IsResourceNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrResourceNotFound) {
		return true
	}
	s := err.Error()
	if strings.Contains(s, "unexpected status code 404") {
		return true
	}
	if strings.Contains(s, "template group not found") {
		return true
	}
	if strings.Contains(s, "template not found for the") {
		return true
	}
	// notification-srv envelope: JSON status 404 while HTTP may still be 200
	if strings.Contains(s, `"status":404`) {
		return true
	}
	return false
}
