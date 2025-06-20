package proxy

import (
	"context"
	"crypto/tls"
	"io"
	"net"
)

type TcpHandler interface {
	HandleTcp(*Context)
}

type HandleTcpFn func(*Context)

func (f HandleTcpFn) HandleTcp(ctx *Context) { f(ctx) }

var defaultTcpHandler HandleTcpFn = func(ctx *Context) { handleTcp(ctx) }

func handleTcp(ctx *Context) {
	var tlsConfig *tls.Config
	if _, ok := ctx.Conn.Conn.(*tls.Conn); ok {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}

	proxyConn, err := dialWithDialer(ctx.dialer, "tcp", net.JoinHostPort(ctx.DstHost, ctx.DstPort), tlsConfig)
	if err != nil {
		ctx.Error(err)
		return
	}
	defer proxyConn.Close()

	c, cancel := context.WithCancel(context.Background())
	go tcpCopy(proxyConn, ctx.Conn, ctx, c, cancel)
	go tcpCopy(ctx.Conn, proxyConn, ctx, c, cancel)
	<-c.Done()
}

type ctxWriter struct {
	net.Conn
	ctx *Context
}

func (w *ctxWriter) Write(p []byte) (int, error) {
	_, err := w.Conn.Write(w.ctx.filterRaw(p, w.ctx))
	return len(p), err
}

func tcpCopy(dst, src net.Conn, ctx *Context, c context.Context, cancel context.CancelFunc) {
	if _, err := io.Copy(&ctxWriter{dst, ctx}, src); err != nil {
		if isContextAlive(c) && err != io.EOF {
			ctx.Error(err)
			cancel()
		}
		return
	}
}
