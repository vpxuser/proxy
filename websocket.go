package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"github.com/gobwas/ws"
	"net"
	"net/http"
)

type WsHandler interface {
	HandleWs(*Context)
}

type HandleWsFn func(*Context)

func (f HandleWsFn) HandleWs(ctx *Context) { f(ctx) }

var defaultWsHandler HandleWsFn = func(ctx *Context) { handleWs(ctx) }

func handleWs(ctx *Context) {
	var tlsConfig *tls.Config
	if _, ok := ctx.Conn.Conn.(*tls.Conn); ok {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}

	proxyConn, err := dialWithDialer(ctx.dialer, "tcp", ctx.DstHost, tlsConfig)
	if err != nil {
		ctx.Error(err)
		return
	}
	defer proxyConn.Close()

	ctx.Req.Header.Del("Sec-WebSocket-Extensions")
	if err = ctx.Req.WriteProxy(proxyConn); err != nil {
		ctx.Error(err)
		return
	}

	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), ctx.Req)
	if err != nil {
		ctx.Error(err)
		return
	}
	defer resp.Body.Close()

	if err = resp.Write(ctx.Conn); err != nil {
		ctx.Error(err)
		return
	}

	c, cancel := context.WithCancel(context.Background())
	go wsCopy(proxyConn, ctx.Conn, ctx, c, cancel)
	go wsCopy(ctx.Conn, proxyConn, ctx, c, cancel)
	<-c.Done()
}

func wsCopy(dst, src net.Conn, ctx *Context, c context.Context, cancel context.CancelFunc) {
	for isContextAlive(c) {
		frame, err := ws.ReadFrame(src)
		if err != nil {
			if isContextAlive(c) {
				ctx.Error(err)
				cancel()
			}
			return
		}

		if err = ws.WriteFrame(dst, ctx.filterWs(frame, ctx)); err != nil {
			if isContextAlive(c) {
				ctx.Error(err)
				cancel()
			}
			return
		}
	}
}

func isContextAlive(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}
