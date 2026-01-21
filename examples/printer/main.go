package main

import (
	"github.com/gobwas/ws"
	"github.com/vpxuser/proxy"
	"net/http"
	"net/http/httputil"
)

func main() {
	proxy.SetLogLevel(proxy.TraceLevel)
	proxy.Infof("use ca certificate:\n       ├── 证书颁发机构：%s\n       ├── 证书域名：%s\n       └── 证书失效日期：%s",
		proxy.Certificate.Issuer.String(), proxy.Certificate.Subject.String(), proxy.Certificate.NotAfter.String())
	tlsConf := proxy.FromCA(proxy.Certificate, proxy.PrivateKey)
	proxy.Infof("init tls config function")
	conf := proxy.NewConfig(tlsConf)
	conf.DefaultSNI = "www.google.com"
	proxy.Infof("init default sni: www.google.com")
	conf.ClientTLSConfig.InsecureSkipVerify = true
	proxy.Infof("allow untrust certificate")
	conf.WithReqMatcher().Handle(func(req *http.Request, ctx *proxy.Context) (*http.Request, *http.Response) {
		request, err := httputil.DumpRequest(req, true)
		if err != nil {
			ctx.Error(err)
			return req, nil
		}
		ctx.Infof("\n%s", request)
		return req, nil
	})
	conf.WithRespMatcher().Handle(func(resp *http.Response, ctx *proxy.Context) *http.Response {
		response, err := httputil.DumpResponse(resp, true)
		if err != nil {
			ctx.Error(err)
			return resp
		}
		ctx.Infof("\n%s", response)
		return resp
	})
	conf.WithWsMatcher().Handle(func(frame ws.Frame, ctx *proxy.Context) ws.Frame {
		payload := frame.Payload
		if frame.Header.Masked {
			payload = ws.UnmaskFrame(frame).Payload
		}
		ctx.Infof("\n%s", payload)
		return frame
	})
	conf.WithRawMatcher().Handle(func(raw []byte, ctx *proxy.Context) []byte {
		ctx.Infof("\n%s", raw)
		return raw
	})
	proxy.Infof("start mitm server")
	err := proxy.ListenAndServe("0.0.0.0:8080", conf)
	if err != nil {
		proxy.Fatal(err)
	}
}
