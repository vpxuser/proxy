package proxy

import (
	"bufio"
	"crypto/tls"
	"github.com/gorilla/websocket"
	"github.com/inconshreveable/go-vhost"
	"net"
	"net/http"
)

type Dispatcher interface {
	Dispatch(*Context)
}

type DispatchFn func(*Context)

func (f DispatchFn) Dispatch(ctx *Context) { f(ctx) }

var defaultDispatcher DispatchFn = func(ctx *Context) { dispatch(ctx) }

func dispatch(ctx *Context) {
	buf, err := ctx.Conn.Peek(2)
	if err != nil {
		ctx.Error(err)
		return
	}

	reverseLookup := defaultResolver.ReverseLookup

	isTLS := buf[0] == 0x16 && buf[1] == 0x03

	if isTLS {
		san := ctx.DstHost

		if !IsDomain(san) {
			if record, ok := reverseLookup.Load(san); ok {
				san = record.(string)
			} else {
				rawConn, err := vhost.TLS(ctx.Conn)
				if err != nil {
					ctx.Warn("No SNI provided, using fallback cert")

					san = ctx.DefaultSAN
					if san == "" {
						ctx.TcpHandler.HandleTcp(ctx)
						return
					}
				} else {
					san = rawConn.Host()
					reverseLookup.Store(ctx.DstHost, san)
					ctx.Conn = NewConn(rawConn)
				}
			}
		}

		tlsCfg, err := ctx.TLSConfig.From(san)
		if err != nil {
			ctx.Error(err)
			return
		}

		tlsCfg.InsecureSkipVerify = true
		ctx.Conn = NewConn(tls.Server(ctx.Conn, tlsCfg))
	}

	buf, err = ctx.Conn.Peek(3)
	if err != nil {
		ctx.Error(err)
		return
	}

	if _, ok := HttpMethods[string(buf)]; ok {
		ctx.Req, err = http.ReadRequest(bufio.NewReader(ctx.Conn))
		if err != nil {
			ctx.Error(err)
			return
		}

		if isTLS && !IsDomain(ctx.DstHost) {
			hostname, _, err := net.SplitHostPort(ctx.Req.Host)
			if err != nil {
				hostname = ctx.Req.Host
			}
			reverseLookup.Store(ctx.DstHost, hostname)
		}

		if websocket.IsWebSocketUpgrade(ctx.Req) {
			ctx.WsHandler.HandleWs(ctx)
			return
		}

		ctx.HttpHandler.HandleHttp(ctx)
		return
	}

	ctx.TcpHandler.HandleTcp(ctx)
}
