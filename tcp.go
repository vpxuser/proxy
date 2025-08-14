package proxy

import (
	"context"
	"crypto/tls"
	"errors"
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
	var tlsCfg *tls.Config
	if _, ok := ctx.Conn.Conn.(*tls.Conn); ok {
		tlsCfg = &tls.Config{InsecureSkipVerify: true}
	}

	dstAddr := net.JoinHostPort(ctx.DstHost, ctx.DstPort)
	proxyConn, err := dialWithDialer(ctx.dialer, "tcp", dstAddr, tlsCfg)
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
		if c.Err() == nil && !errors.Is(err, io.EOF) {
			ctx.Error(err)
			cancel()
		}
		return
	}
}
