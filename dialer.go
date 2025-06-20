package proxy

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
)

func FromURL(proxyURL string, dialer proxy.Dialer) (Dialer, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http":
		return &httpDialer{u: u}, nil
	default:
		return newDefaultDialer(u, dialer)
	}
}

type Dialer interface {
	Dial(string, string, ...DialerOption) (net.Conn, error)
	WithOptions(...DialerOption) Dialer
}

type DialerOption func(*dialerConfig)

type dialerConfig struct {
	tlsConfig *tls.Config
}

func newDialerConfig() *dialerConfig {
	return &dialerConfig{}
}

func (d *dialerConfig) clone() *dialerConfig {
	var tlsConfig *tls.Config
	if d.tlsConfig != nil {
		tlsConfig = d.tlsConfig.Clone()
	}

	return &dialerConfig{
		tlsConfig: tlsConfig,
	}
}

func (d *dialerConfig) cloneWithOptions(opts ...DialerOption) *dialerConfig {
	var cfg *dialerConfig
	if d != nil {
		cfg = d.clone()
	} else {
		cfg = newDialerConfig()
	}

	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func WithTLSConfig(cfg *tls.Config) DialerOption {
	return func(c *dialerConfig) {
		c.tlsConfig = cfg
	}
}

type httpDialer struct {
	u   *url.URL
	cfg *dialerConfig
}

func (h *httpDialer) WithOptions(opts ...DialerOption) Dialer {
	if h.cfg == nil {
		h.cfg = newDialerConfig()
	}

	for _, opt := range opts {
		opt(h.cfg)
	}

	return h
}

func (h *httpDialer) Dial(network, addr string, opts ...DialerOption) (net.Conn, error) {
	conn, err := net.Dial(network, h.u.Host)
	if err != nil {
		return nil, err
	}

	if cfg := h.cfg.cloneWithOptions(opts...); cfg != nil {
		if cfg.tlsConfig != nil {
			if err = httpConnect(conn, addr); err != nil {
				conn.Close()
				return nil, err
			}
			return newConnWithTLS(conn, cfg.tlsConfig)
		}
	}
	return conn, nil
}

func httpConnect(c net.Conn, target string) error {
	req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\nProxy-Connection: Keep-Alive\r\n\r\n", target, target)
	if _, err := c.Write([]byte(req)); err != nil {
		return err
	}

	resp, err := http.ReadResponse(bufio.NewReader(c), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("proxy CONNECT response status: %s", resp.Status)
	}
	return nil
}

func newConnWithTLS(conn net.Conn, config *tls.Config) (net.Conn, error) {
	tlsClient := tls.Client(conn, config)
	if err := tlsClient.Handshake(); err != nil {
		conn.Close()
		return nil, err
	}
	return tlsClient, nil
}

type defaultDialer struct {
	dialer proxy.Dialer
	cfg    *dialerConfig
}

func newDefaultDialer(u *url.URL, d proxy.Dialer) (*defaultDialer, error) {
	dialer, err := proxy.FromURL(u, d)
	if err != nil {
		return nil, err
	}

	return &defaultDialer{
		dialer: dialer,
	}, nil
}

func (d *defaultDialer) WithOptions(opts ...DialerOption) Dialer {
	if d.cfg == nil {
		d.cfg = newDialerConfig()
	}

	for _, opt := range opts {
		opt(d.cfg)
	}

	return d
}

func (d *defaultDialer) Dial(network, addr string, opts ...DialerOption) (c net.Conn, err error) {
	conn, err := d.dialer.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	if cfg := d.cfg.cloneWithOptions(opts...); cfg != nil {
		if cfg.tlsConfig != nil {
			return newConnWithTLS(conn, cfg.tlsConfig)
		}
	}
	return conn, nil
}
