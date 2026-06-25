# printer — MITM 代理示例工具

完整的 CLI MITM 代理示例，展示 `proxy` 包的所有功能，将拦截到的流量（HTTP/WebSocket/TCP）打印到标准输出。

## 快速开始

```bash
go run main.go
```

代理监听 `0.0.0.0:8080`，将浏览器或系统代理指向此地址即可。

### 交叉编译

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o proxy main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o proxy.exe main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o proxy main.go
```

## 功能演示

注册了全部 4 种 matcher，用于打印流量：

| Matcher | 作用 |
|---|---|
| `WithReqMatcher` | 打印 HTTP 请求 |
| `WithRespMatcher` | 打印 HTTP 响应 |
| `WithWsMatcher` | 打印 WebSocket 帧（处理掩码） |
| `WithRawMatcher` | 打印原始 TCP 数据 |

## 配置说明

工具使用内置的开发 CA 证书（`proxy.Certificate` + `proxy.PrivateKey`）。生产环境应替换为自定义 CA，或使用 `proxy.FromSelfSigned()`。

如需链式上游代理，在 `ListenAndServe` 前设置 `conf.Dialer`：

```go
dialer, _ := proxy.SOCKS5("tcp", "127.0.0.1:10808", nil, nil)
conf.Dialer = dialer
```

## 注意事项

- 默认日志级别为 `TraceLevel`，展示所有细节
- 客户端 TLS 验证已关闭（`InsecureSkipVerify = true`）
- 默认 SNI 回退地址：`www.google.com`
