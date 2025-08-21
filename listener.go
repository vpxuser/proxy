package proxy

import (
	"net"
)

type Listener struct {
	net.Listener
	cfg *Config
}

func NewListener(ln net.Listener, cfg *Config) *Listener {
	return &Listener{Listener: ln, cfg: cfg}
}

func Listen(network, addr string, cfg *Config) (*Listener, error) {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	return NewListener(ln, cfg), nil
}
