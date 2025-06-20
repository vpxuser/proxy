package proxy

import (
	"bufio"
	"crypto/tls"
	"golang.org/x/net/proxy"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
)

func doRequest(t *testing.T, dialer Dialer, dstURL string, tlsConfig *tls.Config) {
	u, err := url.Parse(dstURL)
	if err != nil {
		t.Fatalf("parse url %s failed: %v", dstURL, err)
	}

	addr := u.Host
	if !strings.Contains(addr, ":") {
		if u.Scheme == "https" {
			addr += ":443"
		} else {
			addr += ":80"
		}
	}

	dialerWithOpts := dialer
	var opts []DialerOption
	if tlsConfig != nil {
		dialerWithOpts = dialer.WithOptions(WithTLSConfig(tlsConfig))
		opts = append(opts, WithTLSConfig(tlsConfig))
	}

	conn, err := dialerWithOpts.Dial("tcp", addr, opts...)
	if err != nil {
		t.Fatalf("dial %s failed: %v", addr, err)
	}
	defer conn.Close()

	req, err := http.NewRequest(http.MethodGet, dstURL, nil)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}

	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		t.Fatalf("dump request failed: %v", err)
	}
	t.Log(string(dump))

	// 写请求到连接
	if err = req.WriteProxy(conn); err != nil {
		t.Fatalf("write request failed: %v", err)
	}

	// 读响应
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatalf("read response failed: %v", err)
	}
	defer resp.Body.Close()

	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		t.Fatalf("dump response failed: %v", err)
	}
	t.Log(string(dump))
}

func TestHttpProxy(t *testing.T) {
	dialer, err := FromURL("http://127.0.0.1:10808", proxy.Direct)
	if err != nil {
		t.Fatal(err)
	}

	doRequest(t, dialer, "http://www.baidu.com", nil)
	doRequest(t, dialer, "https://www.baidu.com", &tls.Config{InsecureSkipVerify: true})
}

func TestSocks5Proxy(t *testing.T) {
	dialer, err := FromURL("socks5://127.0.0.1:10808", proxy.Direct)
	if err != nil {
		t.Fatal(err)
	}

	doRequest(t, dialer, "http://www.baidu.com", nil)
	doRequest(t, dialer, "https://www.baidu.com", &tls.Config{InsecureSkipVerify: true})
}

func TestSocks5hProxy(t *testing.T) {
	dialer, err := FromURL("socks5h://127.0.0.1:10808", proxy.Direct)
	if err != nil {
		t.Fatal(err)
	}

	doRequest(t, dialer, "http://www.baidu.com", nil)
	doRequest(t, dialer, "https://www.baidu.com", &tls.Config{InsecureSkipVerify: true})
}
