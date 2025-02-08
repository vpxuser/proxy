package handler

import (
	"github.com/vpxuser/proxy"
	yaklog "github.com/yaklang/yaklang/common/log"
	"net/http"
	"net/http/httputil"
)

var DumpRequest proxy.HandleReq = func(req *http.Request, ctx *proxy.Context) (*http.Request, *http.Response) {
	reqStr, err := httputil.DumpRequest(req, true)
	if err != nil {
		yaklog.Errorf("%s dump request failed - %v", ctx.Preffix(false), err)
		return req, nil
	}

	yaklog.Infof("%s\n%s", ctx.Preffix(false), reqStr)

	return req, nil
}

var DumpResponse proxy.HandleResp = func(resp *http.Response, ctx *proxy.Context) *http.Response {
	respStr, err := httputil.DumpResponse(resp, true)
	if err != nil {
		yaklog.Errorf("%s dump response failed - %v", ctx.Preffix(true), err)
		return resp
	}

	yaklog.Infof("%s\n%s", ctx.Preffix(false), respStr)

	return resp
}
