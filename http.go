package proxy

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strings"
)

type HttpHandler interface {
	HandleHttp(*Context)
}

type HandleHttpFn func(*Context)

func (f HandleHttpFn) HandleHttp(ctx *Context) { f(ctx) }

var defaultHttpHandler HandleHttpFn = func(ctx *Context) { handleHttp(ctx) }

func handleHttp(ctx *Context) {
	var tlsConfig *tls.Config
	if _, ok := ctx.Conn.Conn.(*tls.Conn); ok {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}

	dstAddr := net.JoinHostPort(ctx.DstHost, ctx.DstPort)
	proxyConn, err := dialWithDialer(ctx.dialer, "tcp", dstAddr, tlsConfig)
	if err != nil {
		ctx.Error(err)
		return
	}
	defer proxyConn.Close()

	src := bufio.NewReader(ctx.Conn)
	dst := bufio.NewReader(proxyConn)
	for {
		req, err := http.ReadRequest(src)
		if err != nil {
			if err != io.EOF {
				ctx.Error(err)
			}
			return
		}

		req, resp := ctx.filterReq(req, ctx)
		if resp == nil {
			if req == nil {
				return
			}

			err := req.WriteProxy(proxyConn)
			if err != nil {
				if err != io.EOF {
					ctx.Error(err)
				}
				return
			}

			resp, err = http.ReadResponse(dst, req)
			if err != nil {
				if err != io.EOF {
					ctx.Error(err)
				}
				return
			}
		}

		err = ctx.filterResp(resp, ctx).Write(ctx.Conn)
		if err != nil {
			if err != io.EOF {
				ctx.Error(err)
			}
			return
		}
		resp.Body.Close()

		if !isKeepAlive(req) {
			return
		}
	}
}

func isKeepAlive(req *http.Request) bool {
	if req.ProtoMajor == 1 {
		connection := req.Header.Get("Connection")
		connection = strings.ToLower(connection)

		switch req.ProtoMinor {
		case 0:
			return connection == "keep-alive"
		case 1:
			return connection != "close"
		}
	}
	return false
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
