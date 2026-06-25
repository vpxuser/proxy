package proxy

import (
	"github.com/google/uuid"
	"net"
	"strings"
)

// ListenAndServe creates a TCP listener on the given address and starts
// the proxy loop. It blocks until the listener is closed.
func ListenAndServe(addr string, cfg *Config) error {
	inner, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return Serve(inner, cfg)
}

// Serve starts the proxy loop on an existing net.Listener.
// It blocks until the listener is closed.
func Serve(ln net.Listener, cfg *Config) error {
	return NewListener(ln, cfg).Serve()
}

// newSessionID generates a short unique session identifier (16 hex chars).
func newSessionID() string {
	id := uuid.New().String()
	return strings.ReplaceAll(id, "-", "")[:16]
}
