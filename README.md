# proxy — Go MITM Proxy Library

[![Go Reference](https://pkg.go.dev/badge/github.com/vpxuser/proxy)](https://pkg.go.dev/github.com/vpxuser/proxy)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vpxuser/proxy)](https://github.com/vpxuser/proxy)

A Go package for building man-in-the-middle (MITM) proxy tools. Intercept, modify, and forward traffic over HTTP, HTTPS, TLS, WebSocket, and TCP. Suitable for security testing, traffic analysis, and protocol research.

## Features

- **HTTP/HTTPS MITM** — Intercept and modify HTTP requests/responses with automatic TLS certificate generation
- **WebSocket MITM** — Frame-level interception and modification
- **TCP MITM** — Raw TCP traffic forwarding with optional modification
- **SOCKS5** — RFC 1928 compliant SOCKS5 CONNECT handshake
- **Transparent proxy** — Protocol-aware dispatch without explicit client configuration (internal, see `tproxyDispatch` in dispatcher.go)
- **Matcher chain** — Fluent API for conditional request/response/WS/TCP handling
- **Concurrency limiter** — Optional goroutine budget per proxy session
- **Upstream proxy** — Chainable via HTTP CONNECT or SOCKS5 upstream

## Installation

```bash
go get github.com/vpxuser/proxy
```

## Quick Start

### HTTP/HTTPS MITM Proxy

Uses the embedded development CA certificate for TLS interception.

```go
package main

import (
	"net/http"
	"net/http/httputil"

	"github.com/vpxuser/proxy"
)

func main() {
	tlsConf := proxy.FromCA(proxy.Certificate, proxy.PrivateKey)
	cfg := proxy.NewConfig(tlsConf)
	cfg.DefaultSNI = "www.google.com"

	// Log all HTTP requests
	cfg.WithReqMatcher().Handle(func(req *http.Request, ctx *proxy.Context) (*http.Request, *http.Response) {
		dump, _ := httputil.DumpRequest(req, true)
		ctx.Infof("\n%s", dump)
		return req, nil
	})

	// Log all HTTP responses
	cfg.WithRespMatcher().Handle(func(resp *http.Response, ctx *proxy.Context) *http.Response {
		dump, _ := httputil.DumpResponse(resp, true)
		ctx.Infof("\n%s", dump)
		return resp
	})

	proxy.ListenAndServe("0.0.0.0:8080", cfg)
}
```

### SOCKS5 Proxy

```go
cfg := proxy.NewConfig(tlsConf)
cfg.Negotiator = proxy.Socks5Negotiator
proxy.ListenAndServe("0.0.0.0:1080", cfg)
```

### WebSocket MITM

```go
cfg.WithWsMatcher().Handle(func(frame ws.Frame, ctx *proxy.Context) ws.Frame {
	payload := frame.Payload
	if frame.Header.Masked {
		payload = ws.UnmaskFrame(frame).Payload
	}
	ctx.Infof("\n%s", payload)
	return frame
})
```

### TCP MITM

```go
cfg.WithRawMatcher().Handle(func(raw []byte, ctx *proxy.Context) []byte {
	ctx.Infof("\n%s", raw)
	return raw
})
```

### Custom CA Certificate

Replace the embedded development CA with your own.

```go
cert, _ := tls.X509KeyPair(certPEM, keyPEM)
x509Cert, _ := x509.ParseCertificate(cert.Certificate[0])
tlsConf := proxy.FromCA(x509Cert, cert.PrivateKey)
// Or generate on the fly:
tlsConf := proxy.FromSelfSigned()
```

### Upstream Proxy

```go
import "golang.org/x/net/proxy"

dialer, _ := proxy.SOCKS5("tcp", "127.0.0.1:10808", nil, nil)
cfg.Dialer = dialer
```

## How It Works

The proxy acts as an intermediary between client and target server:

1. **Handshake** — Accepts connections via HTTP CONNECT or SOCKS5 (configured via `Negotiator`)
2. **Dispatch** — Identifies protocol (HTTP, TLS, TCP) by peeking at initial bytes
3. **TLS Interception** — Generates per-host certificates signed by the CA, terminates TLS from the client, then initiates a new TLS connection to the target
4. **Forwarding** — Passes parsed requests/responses through the matcher chain for modification

> **Note**: Some sites use certificate pinning or HSTS to prevent MITM. The transparent proxy mode can bypass client proxy configuration by working with tools like Proxifier.

## Example Project

See [examples/printer](./examples/printer/) for a complete CLI MITM proxy tool that demonstrates all features.

## License

MIT
