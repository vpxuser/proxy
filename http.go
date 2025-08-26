package proxy

import (
	"bufio"
	"crypto/tls"
	"errors"
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
	reader := bufio.NewReader(ctx.Conn)
	for {
		req, err := http.ReadRequest(reader)
		if err != nil {
			if !IsEOF(err) {
				ctx.Error(err)
			}
			return err
		}

		req, resp := ctx.filterReq(req, ctx)
		if resp == nil {
			if req == nil {
				ctx.Error(ErrNilRequest)
				return ErrNilRequest
			}

			if ctx.DstConn == nil {
				proxyAddr := net.JoinHostPort(ctx.DstHost, ctx.DstPort)
				ctx.DstConn, err = ctx.Dialer.Dial("tcp", proxyAddr)
				if err != nil {
					ctx.Error(err)
					return err
				}

				if ctx.Conn.IsTLS() {
					ctx.DstConn = tls.Client(ctx.DstConn, ctx.ClientTLSConfig)
				}
			}

			err := req.Write(ctx.DstConn)
			if err != nil {
				if !IsEOF(err) {
					ctx.Error(err)
				}
				return err
			}

			resp, err = http.ReadResponse(bufio.NewReader(ctx.DstConn), req)
			if err != nil {
				if !IsEOF(err) {
					ctx.Error(err)
				}
				return err
			}
		}

		err = ctx.filterResp(resp, ctx).Write(ctx.Conn)
		if err != nil {
			ctx.Error(err)
			return err
		}
	}
}
