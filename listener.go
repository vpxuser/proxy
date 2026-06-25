package proxy

import (
	"errors"
	"net"
)

// Listener wraps a net.Listener and associates it with a proxy Config.
// Use Listen or NewListener to create one, then call Serve to start
// the proxy loop, or Shutdown to stop gracefully.
type Listener struct {
	net.Listener
	cfg *Config
}

// NewListener creates a Listener from an existing net.Listener.
func NewListener(ln net.Listener, cfg *Config) *Listener {
	return &Listener{Listener: ln, cfg: cfg}
}

// Listen creates a Listener by binding to the given network and address.
// network must be "tcp"; addr is a TCP address string (e.g. "0.0.0.0:8080").
func Listen(network, addr string, cfg *Config) (*Listener, error) {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	return NewListener(ln, cfg), nil
}

// Serve accepts incoming connections and dispatches them according to
// the configured protocol negotiator and dispatcher. It blocks until
// the underlying listener is closed (via Shutdown or Close).
func (ln *Listener) Serve() error {
	defer ln.Close()
	for {
		id := newSessionID()
		ctx := NewContext(ctxLogger, id[:16], ln.cfg)
		inner, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			ctx.Error(err)
			continue
		}
		ctx.Conn = NewConn(inner)
		go func() {
			defer ctx.Conn.Close()
			if ctx.Negotiator != nil {
				err = ctx.Negotiator.Handshake(ctx)
				if err != nil {
					ctx.Error(err)
					return
				}
			}
			_ = ctx.Dispatcher.Dispatch(ctx)
		}()
	}
}

// Shutdown closes the underlying listener, causing Serve to return
// (after any in-flight connections complete). It is safe to call
// concurrently with Serve.
func (ln *Listener) Shutdown() error {
	return ln.Listener.Close()
}
