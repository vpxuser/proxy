package proxy

import (
	"errors"
	"io"
)

func IsEOF(err error) bool {
	return errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF)
}
