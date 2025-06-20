package proxy

import (
	"net"
)

type Listener struct {
	net.Listener
	cfg *Config
}

func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewConn(conn), nil
}

func NewListener(inner net.Listener, cfg *Config) *Listener {
	if cfg == nil {
		cfg = &Config{}
	}
	return &Listener{Listener: inner, cfg: cfg}
}

func Listen(network, addr string, cfg *Config) (*Listener, error) {
	inner, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	return NewListener(inner, cfg), nil
}
