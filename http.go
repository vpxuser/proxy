package proxy

import (
	"errors"
	yaklog "github.com/yaklang/yaklang/common/log"
	"net"
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
			errNil := errors.New("request and response is nil")
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
