package proxy

import (
	"errors"
	"io"
)

// IsEOF reports whether the error indicates an end-of-file condition,
// matching io.EOF or io.ErrUnexpectedEOF.
func IsEOF(err error) bool {
	return errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF)
}
