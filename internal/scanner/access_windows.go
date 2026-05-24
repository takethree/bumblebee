//go:build windows

package scanner

import (
	"errors"
	"syscall"
)

const (
	errorSharingViolation syscall.Errno = 32
	errorLockViolation    syscall.Errno = 33
)

func isPlatformExpectedAccessError(err error) bool {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return false
	}
	switch errno {
	case syscall.ERROR_ACCESS_DENIED, errorSharingViolation, errorLockViolation:
		return true
	default:
		return false
	}
}
