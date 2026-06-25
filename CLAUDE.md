# proxy — Go MITM Proxy Library

## 项目概述

一个用于构建中间人（MITM）代理工具的 Go 包，支持 HTTP/HTTPS、TLS、WebSocket、TCP 协议的拦截、修改与转发。

## 模块架构

```
server.go         入口点：ListenAndServe、Serve
listener.go       网络监听器封装
negotiator.go     代理协议握手（HTTP CONNECT、SOCKS5）
dispatcher.go     协议识别与分发（HTTP / TLS / TCP / WebSocket）
http.go           HTTP 请求/响应代理
tcp.go            TCP 透传代理
websocket.go      WebSocket 帧级转发
tls.go            TLS 证书生成与管理（FromCA、FromSelfSigned、From）
conn.go           PeekReader + Conn 包装，支持"偷看"字节不消耗
ctx.go            请求上下文，集成 Logger
config.go         配置聚合
matcher.go        过滤器/匹配器链式 API（Req/Resp/Ws/Raw）
logger.go         日志封装（基于 logrus）
resolver.go       反向 DNS 缓存
limiter.go        并发限速
dialer.go         HTTP CONNECT 上游拨号器
cert.go           内置默认 CA 证书与私钥
errors.go         跨平台错误判定
regexp.go         域名正则
detector.go       HTTP 方法检测
```

## 关键接口

每个核心组件都是单一方法接口 + 函数适配器模式：

```go
type Dispatcher interface { Dispatch(*Context) error }
type Negotiator interface { Handshake(*Context) error }
type HttpHandler interface { HandleHttp(*Context) error }
type WsHandler    interface { HandleWs(*Context) error }
type TcpHandler   interface { HandleTcp(*Context) error }
type TLSConfig    interface { From(string) (*tls.Config, error) }
type Resolver     interface { SetPTR(string, string); GetPTR(string) (string, bool) }
type Limiter      interface { Acquire(); Release() }
```

## 构建 & 测试

```bash
go build ./...
go test ./... -v -count=1
go vet ./...
```

> `TestLimiter` 因 goroutine 中 1s sleep 设计耗时约 135s。`TestServer` 与 `TestConn` 共用端口 23999，顺序执行时后者可能因 TIME_WAIT 卡住——建议单独运行测试或用不同端口。

交叉编译示例（主项目而非库本身）：
```bash
GOOS=linux GOARCH=amd64 go build -o proxy ./examples/printer/
```

## 编程约定

- **语言**：Go 标准库风格，公开类型有中英文双行注释（中文在上方，英文在下方）
- **接口模式**：偏好单一方法接口 + 函数适配器（`type XxxFn func(...) ...`）
- **错误处理**：使用 `errors.Is` 包裹错误并逐层返回，不丢异常
- **配置方式**：函数式 Options 模式（`WithReqMatcher`、`WithRespMatcher` 等）
- **Context**：`*Context` 嵌入 `*Config` 以全局访问配置

## 需要注意的设计细节

- **Conn.PeekRd**：`dispatcher.go` 中的协议检测依赖 `PeekReader` 先读字节再让后续逻辑重新消费，修改 dispatcher 时需确保 Peek 和 Read 的数据一致性
- **Logger 栈跳数**：`logger.go` 中 `formatter(skip, name)` 的 `skip` 参数依赖于调用深度，新增/删除 logger 包装函数时需要同步更新
- **两种 Dispatcher 模式**：`defaultDispatcher`（标准代理模式）和 `tproxyDispatch`（透明代理模式），协议识别逻辑共享但各自独立演化
- **证书**：`cert.go` 内嵌了开发用默认 CA 证书和私钥，生产环境应使用 `FromCA()` 替换为自己的 CA
- **WebSocket 处理**：`gobwas/ws` 用于帧级读写，`isWebSocketUpgrade` 函数在 `dispatcher.go` 中内联实现
- **HTTP CONNECT 拨号器**：`dialer.go` 中 `init()` 注册了 `http` 协议的 Dialer，导入本包会修改 `golang.org/x/net/proxy` 全局注册表

## 相关工作

- `examples/printer/`：一个完整的 CLI MITM 代理示例工具，支持 HTTPS + WS + TCP 拦截和打印
- `config.yml`：示例工具的配置文件（未入库，需参考 `examples/printer/README.md` 自行创建）
