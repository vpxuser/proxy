package proxy

import (
	"errors"
	"syscall"
)

func IsConnReset(err error) bool {
	return errors.Is(err, syscall.WSAECONNRESET)
}

func IsConnAborted(err error) bool {
	return errors.Is(err, syscall.WSAECONNABORTED)
}
