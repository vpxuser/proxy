package proxy

import (
	"bufio"
	"crypto/tls"
	"errors"
	"github.com/gobwas/ws"
	"io"
	"net"
	"net/http"
	"sync"
)

// WsHandler defines the interface for handling WebSocket requests.
// WsHandler 定义了处理 WebSocket 请求的接口。
type WsHandler interface {
	HandleWs(*Context) error
}

// HandleWsFn is a function type adapter that implements the WsHandler interface.
// HandleWsFn 是一个函数类型适配器，用于实现 WsHandler 接口。
type HandleWsFn func(*Context) error

// HandleWs calls the function itself to handle the WebSocket request.
// HandleWs 会直接调用函数本身来处理 WebSocket 请求。
func (f HandleWsFn) HandleWs(ctx *Context) error { return f(ctx) }

// defaultWsHandler establishes a proxy connection to the target and forwards the WebSocket traffic.
// It performs WebSocket handshake forwarding and frame tunneling.
// defaultWsHandler 会建立到目标地址的代理连接，并转发 WebSocket 流量。
// 它会转发 WebSocket 握手，并在客户端和目标之间进行帧级转发。
var defaultWsHandler HandleWsFn = func(ctx *Context) error {
	// Dial to the target WebSocket server.
	// 拨号连接目标 WebSocket 服务端。
	proxyAddr := net.JoinHostPort(ctx.DstHost, ctx.DstPort)
	proxyConn, err := ctx.GetDialer().Dial("tcp", proxyAddr)
	if err != nil {
		ctx.Error(err)
		return err
	}
	defer proxyConn.Close()

	// Check if the client connection is already TLS
	// 如果客户端连接已经是 TLS（HTTPS 请求）
	if _, ok := ctx.Conn.Conn.(*tls.Conn); ok {
		// Wrap the upstream connection with TLS using ClientTLSConfig
		// 使用 ctx.ClientTLSConfig 在上游连接上建立 TLS 隧道
		proxyConn = tls.Client(proxyConn, ctx.ClientTLSConfig)
	}

	req, err := http.ReadRequest(bufio.NewReader(ctx.Conn))
	if err != nil {
		ctx.Error(err)
		return err
	}

	// Remove unsupported extensions to avoid negotiation issues.
	// 删除扩展字段，避免协商失败。
	req.Header.Del("Sec-WebSocket-Extensions")
	if err = req.WriteProxy(proxyConn); err != nil {
		ctx.Error(err)
		return err
	}

	// Forward the client's handshake request to the target.
	// 转发客户端的握手请求给目标服务器。
	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), req)
	if err != nil {
		ctx.Error(err)
		return err
	}
	defer resp.Body.Close()

	// Forward the handshake response back to the client.
	// 将握手响应写回给客户端。
	if err = resp.Write(ctx.Conn); err != nil {
		ctx.Error(err)
		return err
	}

	// Start bidirectional frame copy with context cancellation support.
	// 启动双向帧复制，并支持上下文取消控制。
	wg := new(sync.WaitGroup)
	go wsCopy(wg, proxyConn, ctx.Conn, ctx)
	go wsCopy(wg, ctx.Conn, proxyConn, ctx)
	wg.Wait()
	return io.EOF
}

// wsCopy reads WebSocket frames from src and writes them to dst,
// optionally filtering frames. Cancels the context if an error occurs.
// wsCopy 从 src 中读取 WebSocket 帧并写入到 dst，
// 可选地对帧进行过滤。如发生错误将取消上下文。
func wsCopy(wg *sync.WaitGroup, dst, src net.Conn, ctx *Context) {
	defer wg.Done()
	for {
		frame, err := ws.ReadFrame(src)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				ctx.Error(err)
			}
			return
		}

		// Filter and write the frame to destination.
		// 过滤并写入帧到目标连接。
		err = ws.WriteFrame(dst, ctx.filterWs(frame, ctx))
		if err != nil {
			if !errors.Is(err, io.EOF) {
				ctx.Error(err)
			}
			return
		}
	}
}
