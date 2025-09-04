//go:build linux
// +build linux

package proxy

import (
	"errors"
	"syscall"
)

func IsConnReset(err error) bool {
	return errors.Is(err, syscall.ECONNRESET)
}

func IsConnAborted(err error) bool {
	return errors.Is(err, syscall.ECONNABORTED)
}
