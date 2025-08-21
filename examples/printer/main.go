package main

import (
	"github.com/vpxuser/proxy"
	"net"
	"net/http"
	"net/http/httputil"
)

func main() {
	proxy.SetLogLevel(proxy.TraceLevel)

	cfg := proxy.NewConfig()
	cfg.DefaultSAN = Cfg.SAN

	//dialer, err := proxy.FromURL(Cfg.Proxy, mode.Direct)
	//if err != nil {
	//	proxy.Fatal(err)
	//}

	cfg.TLSConfig = proxy.FromCA(Cert, Key)

	cfg.WithReqMatcher().Handle(func(req *http.Request, ctx *proxy.Context) (*http.Request, *http.Response) {
		dump, err := httputil.DumpRequest(req, true)
		if err != nil {
			ctx.Error(err)
			dump, _ = httputil.DumpRequest(req, false)
		}
		ctx.Infof("是否为TLS：%v,\n%s", ctx.Conn.IsTLS(), dump)
		return req, nil
	})

	cfg.WithRespMatcher().Handle(func(resp *http.Response, ctx *proxy.Context) *http.Response {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			ctx.Error(err)
			dump, _ = httputil.DumpResponse(resp, false)
		}
		ctx.Infof("\n%s", dump)
		return resp
	})

	//cfg.WithWsMatcher().Handle(func(frame ws.Frame, ctx *proxy.Context) ws.Frame {
	//	ctx.Infof("\n%s", frame.Payload)
	//	return frame
	//})
	//
	//cfg.WithRawMatcher().Handle(func(raw []byte, ctx *proxy.Context) []byte {
	//	ctx.Infof("\n%s", raw)
	//	return raw
	//})

	if err := proxy.ListenAndServe(net.JoinHostPort(Cfg.Host, Cfg.Port), cfg); err != nil {
		proxy.Fatal(err)
	}
}
