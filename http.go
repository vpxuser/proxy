package proxy

import (
	"errors"
	yaklog "github.com/yaklang/yaklang/common/log"
	"net"
	"net/http"
)

func (h *HttpProxy) handleHttp(client net.Conn, ctx *Context) (err error) {
	if ctx.IsTLS {
		ctx.Protocol, ctx.Request.URL.Scheme = "HTTPS", "https"
	}

	ctx.Request.URL.Host, ctx.Request.RequestURI = ctx.Request.Host, ""

	ctx.Request, ctx.Response = h.filterReq(ctx.Request, ctx)

	if ctx.Response == nil {
		if ctx.Request != nil {
			ctx.Response, err = h.HTTPClient.Do(ctx.Request)
			if err != nil {
				yaklog.Errorf("%s send request to remote failed - %v", ctx.Preffix(), err)
				return err
			}
		} else {
			errNil := errors.New("request and response are nil")
			yaklog.Errorf("%s fliterReq error - %v", ctx.Preffix(), errNil)
			return errNil
		}
		ctx.Response = h.filterResp(ctx.Response, ctx)
	}

	if err = ctx.Response.Write(client); err != nil {
		yaklog.Errorf("%s send response to client failed - %v", ctx.Preffix(), err)
		return err
	}

	return nil
}

func (h *HttpProxy) handleHTTP(req *http.Request, tls bool, client net.Conn, ctx *Context) {
	req.URL.Scheme = "http"
	if tls {
		req.URL.Scheme = "https"
	}

	req.URL.Host, req.RequestURI = req.Host, ""

	req, resp := h.filterReq(req, ctx)

	if resp != nil {
		if err := resp.Write(client); err != nil {
			yaklog.Error(err)
			return
		}
		return
	}

	if req == nil {
		return
	}

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		yaklog.Error(err)
		return
	}

	ctx.Response = resp

	resp = h.filterResp(resp, ctx)
	if err = resp.Write(client); err != nil {
		yaklog.Error(err)
		return
	}
	return
}
