package main

import (
	"crypto/tls"
	"github.com/gobwas/ws"
	"github.com/vpxuser/proxy"
	mode "golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/http/httputil"
)

func main() {
	proxy.SetLogLevel(proxy.TraceLevel)

	cfg := proxy.NewConfig()
	cfg.DefaultSAN = Cfg.SAN

	dialer, err := proxy.FromURL(Cfg.Proxy, mode.Direct)
	if err != nil {
		proxy.Fatal(err)
	}

	cfg.WithOptions(
		//proxy.WithNegotiator(proxy.Socks5Negotiator),
		proxy.WithTLSConfigFn(proxy.FromCA(Cert, Key)),
		proxy.WithDialer(dialer),
	)

	cfg.WithReqMatcher().Handle(func(req *http.Request, ctx *proxy.Context) (*http.Request, *http.Response) {
		dump, err := httputil.DumpRequest(req, true)
		if err != nil {
			ctx.Error(err)
		}
		_, ok := ctx.Conn.Conn.(*tls.Conn)
		ctx.Infof("是否为TLS：%v,\n%s", ok, dump)
		return req, nil
	})

	cfg.WithRespMatcher().Handle(func(resp *http.Response, ctx *proxy.Context) *http.Response {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			ctx.Error(err)
		}
		ctx.Infof("\n%s", dump)
		return resp
	})

	cfg.WithWsMatcher().Handle(func(frame ws.Frame, ctx *proxy.Context) ws.Frame {
		ctx.Infof("\n%s", frame.Payload)
		return frame
	})

	cfg.WithRawMatcher().Handle(func(raw []byte, ctx *proxy.Context) []byte {
		ctx.Infof("\n%s", raw)
		return raw
	})

	if err := proxy.ListenAndServe("tcp", net.JoinHostPort(Cfg.Host, Cfg.Port), cfg); err != nil {
		proxy.Fatal(err)
	}
}
