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

var defaultDispatcher DispatchFn = func(ctx *Context) {
	buf := make([]byte, 3)
	_, err := ctx.Conn.TeeReader().Read(buf)
	if err == nil {
		isTLS := buf[0] == 0x16 && buf[1] == 0x03
		if isTLS {
			serverName := ctx.DstHost
			if !IsDomain(serverName) {
				if record, ok := ctx.Resolver.PTRGet(serverName); ok {
					serverName = record
				} else {
					rawConn, err := vhost.TLS(ctx.Conn)
					if err != nil {
						ctx.Warn("No SNI provided, using fallback cert")

						serverName = ctx.DefaultSAN
						if serverName == "" {
							ctx.TcpHandler.HandleTcp(ctx)
							return
						}
					} else {
						serverName = rawConn.Host()
						ctx.Resolver.PTRSet(ctx.DstHost, serverName)
						ctx.Conn = NewConn(rawConn)
					}
				}
			}

			tlsCfg, err := ctx.TLSConfig.From(serverName)
			if err != nil {
				ctx.Error(err)
				return
			}
			ctx.Conn = NewConn(tls.Server(ctx.Conn, tlsCfg))
		}

		ctx.Req, err = http.ReadRequest(bufio.NewReader(ctx.Conn.TeeReader()))
		if err != nil {
			ctx.TcpHandler.HandleTcp(ctx)
			return
		}

		//todo 确认是http情况下否还需要存储反向DNS解析记录PTR
		if !isTLS && !IsDomain(ctx.DstHost) {
			hostname, _, err := net.SplitHostPort(ctx.Req.Host)
			if err != nil {
				hostname = ctx.Req.Host
			}
			ctx.Resolver.PTRSet(ctx.DstHost, hostname)
		}
	}

	if websocket.IsWebSocketUpgrade(ctx.Req) {
		ctx.WsHandler.HandleWs(ctx)
		return
	}

	ctx.HttpHandler.HandleHttp(ctx)
	return
}
