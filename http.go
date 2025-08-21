package proxy

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
)

type HttpHandler interface {
	HandleHttp(*Context) error
}

type HandleHttpFn func(*Context) error

func (f HandleHttpFn) HandleHttp(ctx *Context) error { return f(ctx) }

var ErrNilRequest = errors.New("nil request")

var defaultHttpHandler HandleHttpFn = func(ctx *Context) error {
	req, err := http.ReadRequest(bufio.NewReader(ctx.Conn))
	if err != nil {
		return handleEOF(err, ctx)
	}

	req, resp := ctx.filterReq(req, ctx)
	if resp == nil {
		if req == nil {
			return handleEOF(ErrNilRequest, ctx)
		}

		if ctx.DstConn == nil {
			ctx.DstConn, err = ctx.GetDialer().Dial("tcp",
				net.JoinHostPort(ctx.DstHost, ctx.DstPort))
			if err != nil {
				return handleEOF(err, ctx)
			}

			if ctx.Conn.IsTLS() {
				ctx.DstConn = tls.Client(ctx.DstConn, ctx.ClientTLSConfig)
			}
		}

		if err := req.Write(ctx.DstConn); err != nil {
			ctx.DstConn.Close()
			return handleEOF(err, ctx)
		}

		resp, err = http.ReadResponse(bufio.NewReader(ctx.DstConn), req)
		if err != nil {
			ctx.DstConn.Close()
			return handleEOF(err, ctx)
		}
	}

	err = ctx.filterResp(resp, ctx).Write(ctx.Conn)
	if err != nil {
		return handleEOF(err, ctx)
	}

	return nil
}

func handleEOF(err error, cxt *Context) error {
	if !errors.Is(err, io.EOF) &&
		!errors.Is(err, io.ErrUnexpectedEOF) {
		cxt.Error(err)
	}
	return err
}
