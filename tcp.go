package proxy

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"sync"
)

type TcpHandler interface {
	HandleTcp(*Context) error
}

type HandleTcpFn func(*Context) error

func (f HandleTcpFn) HandleTcp(ctx *Context) error { return f(ctx) }

var defaultTcpHandler HandleTcpFn = func(ctx *Context) error {
	proxyAddr := net.JoinHostPort(ctx.DstHost, ctx.DstPort)
	proxyConn, err := ctx.Dialer.Dial("tcp", proxyAddr)
	if err != nil {
		ctx.Error(err)
		return err
	}
	defer proxyConn.Close()

	if ctx.Conn.IsTLS() {
		proxyConn = tls.Client(proxyConn, ctx.ClientTLSConfig)
	}

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go tcpCopy(wg, proxyConn, ctx.Conn, ctx)
	go tcpCopy(wg, ctx.Conn, proxyConn, ctx)
	wg.Wait()
	return nil
}

type ctxWriter struct {
	net.Conn
	ctx *Context
}

func (w *ctxWriter) Write(p []byte) (int, error) {
	_, err := w.Conn.Write(w.ctx.filterRaw(p, w.ctx))
	return len(p), err
}

func tcpCopy(wg *sync.WaitGroup, dst, src net.Conn, ctx *Context) {
	defer wg.Done()
	cw := &ctxWriter{dst, ctx}
	_, err := io.Copy(cw, src)
	if err != nil &&
		!errors.Is(err, io.EOF) {
		ctx.Error(err)
	}
	return
}
