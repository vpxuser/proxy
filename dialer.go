package proxy

import (
	"bufio"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
)

type httpDialer struct {
	u       *url.URL
	forward proxy.Dialer
}

func (d *httpDialer) Dial(network, addr string) (net.Conn, error) {
	conn, err := d.forward.Dial(network, d.u.Host)
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: make(http.Header),
	}

	if err := req.Write(conn); err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		conn.Close()
		return nil, fmt.Errorf("proxy CONNECT failed: %s", resp.Status)
	}

	return conn, nil
}

func httpDialerFn(u *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	return &httpDialer{u: u, forward: forward}, nil
}

func init() {
	proxy.RegisterDialerType("http", httpDialerFn)
}
