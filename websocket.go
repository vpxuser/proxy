package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"github.com/gobwas/ws"
	"net"
	"net/http"
)

// WsHandler defines the interface for handling WebSocket requests.
// WsHandler 定义了处理 WebSocket 请求的接口。
type WsHandler interface {
	HandleWs(*Context)
}

// HandleWsFn is a function type adapter that implements the WsHandler interface.
// HandleWsFn 是一个函数类型适配器，用于实现 WsHandler 接口。
type HandleWsFn func(*Context)

// HandleWs calls the function itself to handle the WebSocket request.
// HandleWs 会直接调用函数本身来处理 WebSocket 请求。
func (f HandleWsFn) HandleWs(ctx *Context) { f(ctx) }

// defaultWsHandler is the default WebSocket handler which delegates to handleWs.
// defaultWsHandler 是默认的 WebSocket 处理器，会调用 handleWs 来处理请求。
var defaultWsHandler HandleWsFn = func(ctx *Context) { handleWs(ctx) }

// handleWs establishes a proxy connection to the target and forwards the WebSocket traffic.
// It performs WebSocket handshake forwarding and frame tunneling.
// handleWs 会建立到目标地址的代理连接，并转发 WebSocket 流量。
// 它会转发 WebSocket 握手，并在客户端和目标之间进行帧级转发。
func handleWs(ctx *Context) {
	var tlsConfig *tls.Config
	if _, ok := ctx.Conn.Conn.(*tls.Conn); ok {
		// Skip certificate verification for TLS connections.
		// 如果是 TLS 连接，则跳过证书校验。
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Dial to the target WebSocket server.
	// 拨号连接目标 WebSocket 服务端。
	proxyConn, err := dialWithDialer(ctx.dialer, "tcp", ctx.DstHost, tlsConfig)
	if err != nil {
		ctx.Error(err)
		return
	}
	defer proxyConn.Close()

	// Remove unsupported extensions to avoid negotiation issues.
	// 删除扩展字段，避免协商失败。
	ctx.Req.Header.Del("Sec-WebSocket-Extensions")
	if err = ctx.Req.WriteProxy(proxyConn); err != nil {
		ctx.Error(err)
		return
	}

	// Forward the client's handshake request to the target.
	// 转发客户端的握手请求给目标服务器。
	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), ctx.Req)
	if err != nil {
		ctx.Error(err)
		return
	}
	defer resp.Body.Close()

	// Forward the handshake response back to the client.
	// 将握手响应写回给客户端。
	if err = resp.Write(ctx.Conn); err != nil {
		ctx.Error(err)
		return
	}

	// Start bidirectional frame copy with context cancellation support.
	// 启动双向帧复制，并支持上下文取消控制。
	c, cancel := context.WithCancel(context.Background())
	go wsCopy(proxyConn, ctx.Conn, ctx, c, cancel)
	go wsCopy(ctx.Conn, proxyConn, ctx, c, cancel)
	<-c.Done()
}

// wsCopy reads WebSocket frames from src and writes them to dst,
// optionally filtering frames. Cancels the context if an error occurs.
// wsCopy 从 src 中读取 WebSocket 帧并写入到 dst，
// 可选地对帧进行过滤。如发生错误将取消上下文。
func wsCopy(dst, src net.Conn, ctx *Context, c context.Context, cancel context.CancelFunc) {
	for c.Err() == nil {
		frame, err := ws.ReadFrame(src)
		if err != nil {
			if c.Err() == nil {
				ctx.Error(err)
				cancel()
			}
			return
		}

		// Filter and write the frame to destination.
		// 过滤并写入帧到目标连接。
		if err = ws.WriteFrame(dst, ctx.filterWs(frame, ctx)); err != nil {
			if c.Err() == nil {
				ctx.Error(err)
				cancel()
			}
			return
		}
	}
}

// isContextAlive returns true if the context has not been canceled.
// isContextAlive 用于判断上下文是否仍然有效（未被取消）。
func isContextAlive(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}
