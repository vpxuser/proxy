package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
)

type Dialer interface {
	proxy.Dialer
	SetTLS(isTLS bool)
}

type HttpDialer struct {
	proxyURL *url.URL
	isTLS    bool
}

func (h *HttpDialer) Dial(network, addr string) (c net.Conn, err error) {
	c, err = net.Dial(network, h.proxyURL.Host)
	if err != nil {
		return nil, err
	}

	if h.isTLS {
		_, _ = httpsHandshake(c, addr)

		return DailTLS(c)
	}

	return c, nil
}

func (h *HttpDialer) SetTLS(isTLS bool) {
	h.isTLS = isTLS
}

type DefaultDialer struct {
	dialer proxy.Dialer
	isTLS  bool
}

func (d *DefaultDialer) Dial(network, addr string) (c net.Conn, err error) {
	c, err = d.dialer.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	if d.isTLS {
		return DailTLS(c)
	}

	return c, nil
}

func (d *DefaultDialer) SetTLS(isTLS bool) {
	d.isTLS = isTLS
}

func FromURL(httpClient *http.Client, rawURL any) (dialer proxy.Dialer, err error) {
	// input URL support type are string , *url.URL
	var proxyURL *url.URL
	switch rawURL.(type) {
	case string:
		proxyURL, err = url.Parse(rawURL.(string))
		if err != nil {
			return nil, err
		}
	case *url.URL:
		proxyURL = rawURL.(*url.URL)
	default:
		return nil, fmt.Errorf("unsupported type: %T", rawURL)
	}

	transport := httpClient.Transport.(*http.Transport)

	switch proxyURL.Scheme {
	case "http":
		transport.Proxy = http.ProxyURL(proxyURL)

		return &HttpDialer{proxyURL: proxyURL}, nil
	default:
		dialer, err = proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			return nil, err
		}

		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}

		return &DefaultDialer{dialer: dialer}, nil
	}
}

func httpsHandshake(c net.Conn, addr string) (code int, err error) {
	connReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\n", addr)
	connReq += fmt.Sprintf("Host: %s\r\n", addr)
	connReq += "Connection: close\r\n\r\n"

	if _, err = c.Write([]byte(connReq)); err != nil {
		return 0, err
	}

	connResp, err := http.ReadResponse(bufio.NewReader(c), nil)
	if err != nil {
		return 0, err
	}

	return connResp.StatusCode, nil
}

func DailTLS(c net.Conn) (tlsClient *tls.Conn, err error) {
	tlsClient = tls.Client(c, &tls.Config{InsecureSkipVerify: true})
	if err = tlsClient.Handshake(); err != nil {
		return nil, err
	}
	return tlsClient, nil
}
