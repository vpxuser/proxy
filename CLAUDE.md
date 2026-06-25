# proxy — Go MITM Proxy Library

Go 中间人代理库，支持 HTTP/HTTPS/TLS/WebSocket/TCP 协议的拦截、修改与转发。

## Build & Test

```bash
go build ./...
go test ./... -v -count=1
go vet ./...
```

> 注意测试陷阱（见下方"陷阱"章节）。

Cross-compile（示例工具）:
```bash
GOOS=linux GOARCH=amd64 go build -o proxy ./examples/printer/
```

## Project Structure

| File | Purpose |
|---|---|
| `server.go` / `listener.go` | 入口与连接循环 |
| `negotiator.go` | 代理协议握手（HTTP CONNECT + SOCKS5） |
| `dispatcher.go` | 协议识别与分发（HTTP / TLS / TCP / WS） |
| `http.go` / `tcp.go` / `websocket.go` | 各协议处理 |
| `tls.go` | 证书生成与管理 |
| `conn.go` / `ctx.go` / `config.go` | 基础设施 |
| `matcher.go` | 请求/响应/WS/TCP 过滤器链式 API |

## Key Interfaces

All core components use a single-method interface + function adapter pattern:

```go
type Negotiator interface { Handshake(*Context) error }
type Dispatcher interface { Dispatch(*Context) error }
type HttpHandler interface { HandleHttp(*Context) error }
type WsHandler    interface { HandleWs(*Context) error }
type TcpHandler   interface { HandleTcp(*Context) error }
type TLSConfig    interface { From(string) (*tls.Config, error) }
type Resolver     interface { SetPTR(string, string); GetPTR(string) (string, bool) }
type Limiter      interface { Acquire(); Release() }
```

Adapter convention: `type XxxFn func(*Context) error`

## Conventions

- **Comments**: exported types/functions get bilingual comments (CN above, EN below)
- **Errors**: wrap with `errors.Is`/`fmt.Errorf`, return up the chain
- **Config**: functional Options pattern (`WithReqMatcher`, `WithRespMatcher`, etc.)
- **Context**: `*Context` embeds `*Config`, all components access config through it
- **Two dispatchers**: `defaultDispatcher` (standard proxy) and `tproxyDispatch` (transparent proxy, unexported) — protocol detection logic is shared but evolves independently

## Traps

- **Conn.PeekRd**: dispatcher reads bytes via `PeekReader` then lets downstream re-consume. Modifying dispatcher MUST maintain Peek → Read data consistency.
- **Logger skip depth**: `formatter(skip, name)`'s `skip` param depends on call depth. Adding/removing logger wrappers requires updating the skip count.
- **Embedded CA cert**: `cert.go` contains dev-only default CA. Production MUST replace via `FromCA()`.
- **WebSocket**: uses `gobwas/ws` for frame-level I/O; `isWebSocketUpgrade` inlined in `dispatcher.go`
- **dialer.go**: `init()` registers `http` protocol dialer in `golang.org/x/net/proxy` global registry — importing this package has a side effect
- **TestLimiter**: ~135s due to 1s sleep in goroutine. **TestServer + TestConn** share port 23999 — sequential execution may block on TIME_WAIT. Run independently or use different ports.

## Related

- `examples/printer/`: full CLI MITM proxy tool (reference usage)
- `config.yml`: example config (not checked in, see `examples/printer/README.md`)
