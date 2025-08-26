package proxy

import (
	"errors"
	"io"
	"syscall"
)

func IsConnReset(err error) bool {
	return errors.Is(err, syscall.WSAECONNRESET) ||
		errors.Is(err, syscall.ECONNRESET)
}

func IsConnAborted(err error) bool {
	return errors.Is(err, syscall.WSAECONNABORTED) ||
		errors.Is(err, syscall.ECONNABORTED)
}

func IsEOF(err error) bool {
	return errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF)
}
