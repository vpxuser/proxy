package proxy

import (
	"bufio"
	"crypto/tls"
	"net"
	"net/http"
)

type HttpHandler interface {
	HandleHttp(*Context)
}

type HandleHttpFn func(*Context)

func (f HandleHttpFn) HandleHttp(ctx *Context) { f(ctx) }

var defaultHttpHandler HandleHttpFn = func(ctx *Context) { handleHttp(ctx) }

func handleHttp(ctx *Context) {
	req, resp := ctx.filterReq(ctx.Req, ctx)
	if resp == nil {
		if req == nil {
			return
		}

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

		if err = req.WriteProxy(proxyConn); err != nil {
			ctx.Error(err)
			return
		}

		resp, err = http.ReadResponse(bufio.NewReader(proxyConn), req)
		if err != nil {
			ctx.Error(err)
			return
		}
		defer resp.Body.Close()
	}

	if err := ctx.filterResp(resp, ctx).Write(ctx.Conn); err != nil {
		ctx.Error(err)
	}
}

func dialWithDialer(dialer Dialer, network, addr string, tlsConfig *tls.Config) (net.Conn, error) {
	if dialer != nil {
		var opts []DialerOption
		if tlsConfig != nil {
			opts = append(opts, WithTLSConfig(tlsConfig))
		}

		return dialer.Dial(network, addr, opts...)
	}

	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		return tls.Client(conn, tlsConfig), nil
	}

	return conn, nil
}
