# printer — MITM Proxy Example Tool

A complete CLI MITM proxy example demonstrating all features of the `proxy` package. It prints intercepted traffic (HTTP, WebSocket, TCP) to stdout.

## Quick Start

```bash
go run main.go
```

The proxy listens on `0.0.0.0:8080`. Configure your browser or system proxy to point here.

### Cross Compile

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o proxy main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o proxy.exe main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o proxy main.go
```

## What It Demonstrates

This example hooks all four matcher types to log traffic:

| Matcher | Purpose |
|---|---|
| `WithReqMatcher` | Logs HTTP requests |
| `WithRespMatcher` | Logs HTTP responses |
| `WithWsMatcher` | Logs WebSocket frames (handles unmasking) |
| `WithRawMatcher` | Logs raw TCP data |

## Configuration

The tool uses the embedded development CA certificate (`proxy.Certificate` + `proxy.PrivateKey`). For production, replace with your own CA or use `proxy.FromSelfSigned()`.

To chain an upstream proxy, set `conf.Dialer` before `ListenAndServe`:

```go
dialer, _ := proxy.SOCKS5("tcp", "127.0.0.1:10808", nil, nil)
conf.Dialer = dialer
```

## Notes

- Default log level is `TraceLevel` to show all details
- Client-side TLS verification is disabled (`InsecureSkipVerify = true`)
- Default SNI fallback: `www.google.com`
