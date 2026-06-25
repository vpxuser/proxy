# proxy — Go MITM Proxy Library

[![Go Reference](https://pkg.go.dev/badge/github.com/vpxuser/proxy)](https://pkg.go.dev/github.com/vpxuser/proxy)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vpxuser/proxy)](https://github.com/vpxuser/proxy)

一个用于构建中间人（MITM）代理工具的 Go 包，支持 HTTP/HTTPS/TLS/WebSocket/TCP 协议的拦截、修改与转发，适用于安全测试、流量分析和协议研究。

## 功能特性

- **HTTP/HTTPS 中间人** — 拦截修改 HTTP 请求/响应，自动生成 TLS 证书
- **WebSocket 中间人** — 帧级 WebSocket 消息拦截与修改
- **TCP 中间人** — 原始 TCP 流量转发与修改
- **SOCKS5** — 符合 RFC 1928 的 SOCKS5 CONNECT 握手
- **透明代理** — 无需客户端配置，协议感知分发（配合 Proxifier / iptables）
- **匹配器链** — 链式 API，按条件过滤请求/响应/WS/TCP
- **并发限速** — 可选 goroutine 配额控制
- **上游代理** — 支持 HTTP CONNECT / SOCKS5 上游链式代理

## 安装

```bash
go get github.com/vpxuser/proxy
```

## 快速开始

### HTTP/HTTPS 中间人代理

使用内置开发 CA 证书进行 TLS 中间人攻击：

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

	// 打印所有 HTTP 请求
	cfg.WithReqMatcher().Handle(func(req *http.Request, ctx *proxy.Context) (*http.Request, *http.Response) {
		dump, _ := httputil.DumpRequest(req, true)
		ctx.Infof("\n%s", dump)
		return req, nil
	})

	// 打印所有 HTTP 响应
	cfg.WithRespMatcher().Handle(func(resp *http.Response, ctx *proxy.Context) *http.Response {
		dump, _ := httputil.DumpResponse(resp, true)
		ctx.Infof("\n%s", dump)
		return resp
	})

	proxy.ListenAndServe("0.0.0.0:8080", cfg)
}
```

### SOCKS5 代理

```go
cfg := proxy.NewConfig(tlsConf)
cfg.Negotiator = proxy.Socks5Negotiator
proxy.ListenAndServe("0.0.0.0:1080", cfg)
```

### WebSocket 中间人

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

### TCP 中间人

```go
cfg.WithRawMatcher().Handle(func(raw []byte, ctx *proxy.Context) []byte {
	ctx.Infof("\n%s", raw)
	return raw
})
```

### 自定义 CA 证书

替换内置的开发用 CA 证书：

```go
cert, _ := tls.X509KeyPair(certPEM, keyPEM)
x509Cert, _ := x509.ParseCertificate(cert.Certificate[0])
tlsConf := proxy.FromCA(x509Cert, cert.PrivateKey)
// 或即时生成：
tlsConf := proxy.FromSelfSigned()
```

### 配置上游代理

```go
import "golang.org/x/net/proxy"

dialer, _ := proxy.SOCKS5("tcp", "127.0.0.1:10808", nil, nil)
cfg.Dialer = dialer
```

### 透明代理模式

配合 Proxifier 或 iptables 使用：

```go
cfg.Dispatcher = proxy.TransparentDispatch
```

## 工作原理

代理充当客户端与目标服务器之间的中间人：

1. **握手** — 通过 HTTP CONNECT 或 SOCKS5 接受连接（由 `Negotiator` 配置）
2. **分发** — 通过 peek 首字节识别协议（HTTP / TLS / TCP）
3. **TLS 拦截** — 为每个域名动态签发 CA 证书，终止客户端 TLS 并向目标发起新连接
4. **转发** — 将解析后的请求/响应传入 matcher 链供修改

> **注意**：部分站点使用证书钉扎或 HSTS 防御中间人。透明代理模式可配合 Proxifier 等工具绕过客户端代理限制。

## 示例项目

完整的 CLI MITM 代理示例，展示所有功能：[examples/printer](./examples/printer/)

## 许可证

MIT
