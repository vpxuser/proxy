package proxy

import (
	"github.com/gobwas/ws"
	"net/http"
	"strings"
)

type ReqCond interface {
	MatchReq(req *http.Request, ctx *Context) bool
}

type ReqConds struct {
	httpProxy *HttpProxy
	conds     []ReqCond
}

func (h *HttpProxy) OnRequest(conds ...ReqCond) *ReqConds {
	return &ReqConds{httpProxy: h, conds: conds}
}

func (r *ReqConds) Do(handle HandleReq) {
	r.httpProxy.reqHandlers = append(r.httpProxy.reqHandlers,
		func(req *http.Request, ctx *Context) (*http.Request, *http.Response) {
			for _, cond := range r.conds {
				if !cond.MatchReq(req, ctx) {
					return req, nil
				}
			}
			return handle(req, ctx)
		})
}

func (h *HttpProxy) filterReq(req *http.Request, ctx *Context) (*http.Request, *http.Response) {
	for _, handle := range h.reqHandlers {
		req, ctx.Response = handle(req, ctx)
		if req == nil {
			break
		}
	}
	return req, ctx.Response
}

type ReqMatch func(req *http.Request, ctx *Context) bool

func (r ReqMatch) MatchResp(resp *http.Response, ctx *Context) bool {
	return r(ctx.Request, ctx)
}

func (r ReqMatch) MatchReq(req *http.Request, ctx *Context) bool {
	return r(req, ctx)
}

func Not(cond ReqCond) ReqMatch {
	return func(req *http.Request, ctx *Context) bool {
		return !cond.MatchReq(req, ctx)
	}
}

func ReqHostIs(hosts ...string) ReqMatch {
	hostSet := make(map[string]struct{})
	for _, host := range hosts {
		hostSet[host] = struct{}{}
	}

	return func(req *http.Request, ctx *Context) bool {
		_, ok := hostSet[req.Host]
		return ok
	}
}

func ReqWildcardIs(hosts ...string) ReqMatch {
	wildcardSet := make(map[string]struct{})
	for _, host := range hosts {
		if strings.Contains(host, "*") {
			wildcardSet[host] = struct{}{}
		}
	}

	return func(req *http.Request, ctx *Context) bool {
		_, ok := wildcardSet[WildcardFQDN(req.Host)]
		return ok
	}
}

func ReqUrlIs(urls ...string) ReqMatch {
	urlSet := make(map[string]struct{})
	for _, url := range urls {
		urlSet[url] = struct{}{}
	}

	return func(req *http.Request, ctx *Context) bool {
		_, ok := urlSet[req.Host+req.URL.Path]
		return ok
	}
}

func ReqContentTypeIs(contentTypes ...string) ReqMatch {
	contentTypeSet := make(map[string]struct{})
	for _, contentType := range contentTypes {
		contentTypeSet[contentType] = struct{}{}
	}

	return func(req *http.Request, ctx *Context) bool {
		_, ok := contentTypeSet[req.Header.Get("Content-Type")]
		return ok
	}
}

func ReqUrlLike(urls ...string) ReqMatch {
	urlSet := make(map[string]struct{})
	for _, url := range urls {
		urlSet[url] = struct{}{}
	}

	return func(req *http.Request, ctx *Context) bool {
		paths := strings.Split(req.URL.Path, "/")[1:]
		var path strings.Builder
		for _, p := range paths {
			path.Write([]byte("/" + p))
			if _, ok := urlSet[req.Host+path.String()]; ok {
				return true
			}
		}
		return false
	}
}

func ReqMethodIs(methods ...string) ReqMatch {
	methodSet := make(map[string]struct{})
	for _, method := range methods {
		methodSet[method] = struct{}{}
	}

	return func(req *http.Request, ctx *Context) bool {
		_, ok := methodSet[req.Method]
		return ok
	}
}

type RespCond interface {
	ReqCond
	MatchResp(resp *http.Response, ctx *Context) bool
}

type RespConds struct {
	httpProxy *HttpProxy
	conds     []RespCond
}

func (h *HttpProxy) OnResponse(conds ...RespCond) *RespConds {
	return &RespConds{httpProxy: h, conds: conds}
}

func (r *RespConds) Do(handle HandleResp) {
	r.httpProxy.respHandlers = append(r.httpProxy.respHandlers,
		func(resp *http.Response, ctx *Context) *http.Response {
			for _, cond := range r.conds {
				if !cond.MatchReq(ctx.Request, ctx) {
					return resp
				}
				if !cond.MatchResp(resp, ctx) {
					return resp
				}
			}
			return handle(resp, ctx)
		})
}

func (h *HttpProxy) filterResp(resp *http.Response, ctx *Context) *http.Response {
	for _, handle := range h.respHandlers {
		resp = handle(resp, ctx)
	}
	return resp
}

type RespMatch func(resp *http.Response, ctx *Context) bool

func (r RespMatch) MatchReq(req *http.Request, ctx *Context) bool {
	return r(ctx.Response, ctx)
}

func (r RespMatch) MatchResp(resp *http.Response, ctx *Context) bool {
	return r(resp, ctx)
}

type WebSocketCond interface {
	Match(frame ws.Frame, reverse bool, ctx *Context) bool
}

type WebSocketConds struct {
	httpProxy *HttpProxy
	conds     []WebSocketCond
}

func (h *HttpProxy) OnWebSocket(conds ...WebSocketCond) *WebSocketConds {
	return &WebSocketConds{httpProxy: h, conds: conds}
}

func (w *WebSocketConds) Do(handle HandleWebSocket) {
	w.httpProxy.webSocketHandlers = append(w.httpProxy.webSocketHandlers,
		func(frame ws.Frame, reverse bool, ctx *Context) ws.Frame {
			for _, cond := range w.conds {
				if !cond.Match(frame, reverse, ctx) {
					return frame
				}
			}
			return handle(frame, reverse, ctx)
		})
}

func (h *HttpProxy) filterWebSocket(frame ws.Frame, reverse bool, ctx *Context) ws.Frame {
	for _, handle := range h.webSocketHandlers {
		frame = handle(frame, reverse, ctx)
	}
	return frame
}

type WebSocketMatch func(frame ws.Frame, reverse bool, ctx *Context) bool

func (w WebSocketMatch) Match(frame ws.Frame, reverse bool, ctx *Context) bool {
	return w(frame, reverse, ctx)
}

func WebSocketHostIs(hosts ...string) WebSocketMatch {
	hostSet := make(map[string]struct{})
	for _, host := range hosts {
		hostSet[host] = struct{}{}
	}

	return func(frame ws.Frame, reverse bool, ctx *Context) bool {
		_, ok := hostSet[ctx.RemoteHost]
		return ok
	}
}

type RawCond interface {
	Match(raw []byte, reverse bool, ctx *Context) bool
}

type RawConds struct {
	httpProxy *HttpProxy
	conds     []RawCond
}

func (h *HttpProxy) OnRaw(conds ...RawCond) *RawConds {
	return &RawConds{httpProxy: h, conds: conds}
}

func (r *RawConds) Do(handle HandleRaw) {
	r.httpProxy.rawHandlers = append(r.httpProxy.rawHandlers,
		func(raw []byte, reverse bool, ctx *Context) []byte {
			for _, cond := range r.conds {
				if !cond.Match(raw, reverse, ctx) {
					return raw
				}
			}
			return handle(raw, reverse, ctx)
		})
}

func (h *HttpProxy) filterRaw(raw []byte, reverse bool, ctx *Context) []byte {
	for _, handle := range h.rawHandlers {
		raw = handle(raw, reverse, ctx)
	}
	return raw
}

type RawMatch func(raw []byte, reverse bool, ctx *Context) bool

func (r RawMatch) Match(raw []byte, reverse bool, ctx *Context) bool {
	return r(raw, reverse, ctx)
}

func RemoteIs(hosts ...string) RawMatch {
	hostSet := make(map[string]struct{})
	for _, host := range hosts {
		hostSet[host] = struct{}{}
	}

	return func(raw []byte, reverse bool, ctx *Context) bool {
		_, ok := hostSet[ctx.RemoteHost+":"+ctx.RemotePort]
		return ok
	}
}
