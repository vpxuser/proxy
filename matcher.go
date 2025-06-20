package proxy

import (
	"github.com/gobwas/ws"
	"net/http"
)

type ReqMatcher interface {
	MatchReq(*http.Request, *Context) bool
	RespMatcher
}

type ReqFilter struct {
	cfg     *Config
	matcher []ReqMatcher
}

func (c *Config) WithReqMatcher(matcher ...ReqMatcher) *ReqFilter {
	return &ReqFilter{cfg: c, matcher: matcher}
}

type ReqHandlerFn func(*http.Request, *Context) (*http.Request, *http.Response)

func (r *ReqFilter) Handle(handle ReqHandlerFn) {
	r.cfg.reqHandlers = append(r.cfg.reqHandlers,
		func(req *http.Request, ctx *Context) (*http.Request, *http.Response) {
			for _, matcher := range r.matcher {
				if !matcher.MatchReq(req, ctx) {
					return req, nil
				}
			}
			return handle(req, ctx)
		})
}

func (c *Config) filterReq(req *http.Request, ctx *Context) (*http.Request, *http.Response) {
	var resp *http.Response
	for _, handle := range c.reqHandlers {
		req, resp = handle(req, ctx)
	}
	return req, resp
}

type ReqMatchFn func(*http.Request, *Context) bool

func (f ReqMatchFn) MatchReq(req *http.Request, ctx *Context) bool {
	return f(req, ctx)
}

func (f ReqMatchFn) MatchResp(resp *http.Response, ctx *Context) bool {
	return f(ctx.Req, ctx)
}

func Not(matcher ReqMatcher) ReqMatchFn {
	return func(req *http.Request, ctx *Context) bool {
		return !matcher.MatchReq(req, ctx)
	}
}

func ReqHostIs(hosts ...string) ReqMatchFn {
	match := make(map[string]struct{})
	for _, host := range hosts {
		match[host] = struct{}{}
	}

	return func(req *http.Request, ctx *Context) bool {
		_, ok := match[req.Host]
		return ok
	}
}

type RespMatcher interface {
	MatchResp(*http.Response, *Context) bool
}

type RespFilter struct {
	cfg     *Config
	matcher []RespMatcher
}

func (c *Config) WithRespMatcher(matcher ...RespMatcher) *RespFilter {
	return &RespFilter{cfg: c, matcher: matcher}
}

type RespHandlerFn func(*http.Response, *Context) *http.Response

func (r *RespFilter) Handle(handle RespHandlerFn) {
	r.cfg.respHandlers = append(r.cfg.respHandlers,
		func(resp *http.Response, ctx *Context) *http.Response {
			for _, matcher := range r.matcher {
				if !matcher.MatchResp(resp, ctx) {
					return resp
				}
			}
			return handle(resp, ctx)
		})
}

func (c *Config) filterResp(resp *http.Response, ctx *Context) *http.Response {
	for _, handle := range c.respHandlers {
		resp = handle(resp, ctx)
	}
	return resp
}

type RespMatchFn func(*http.Response, *Context) bool

func (f RespMatchFn) MatchResp(resp *http.Response, ctx *Context) bool {
	return f(resp, ctx)
}

func StatusCodeIs(codes ...int) RespMatchFn {
	match := make(map[int]bool)
	for _, code := range codes {
		match[code] = true
	}
	return func(resp *http.Response, ctx *Context) bool {
		_, ok := match[resp.StatusCode]
		return ok
	}
}

type WsMatcher interface {
	Match(frame ws.Frame, ctx *Context) bool
}

type WsFilter struct {
	cfg     *Config
	matcher []WsMatcher
}

func (c *Config) WithWsMatcher(matcher ...WsMatcher) *WsFilter {
	return &WsFilter{cfg: c, matcher: matcher}
}

type WsHandlerFn func(ws.Frame, *Context) ws.Frame

func (w *WsFilter) Handle(handle WsHandlerFn) {
	w.cfg.wsHandlers = append(w.cfg.wsHandlers,
		func(frame ws.Frame, ctx *Context) ws.Frame {
			for _, matcher := range w.matcher {
				if !matcher.Match(frame, ctx) {
					return frame
				}
			}
			return handle(frame, ctx)
		})
}

func (c *Config) filterWs(frame ws.Frame, ctx *Context) ws.Frame {
	for _, handle := range c.wsHandlers {
		frame = handle(frame, ctx)
	}
	return frame
}

type WsMatchFn func(ws.Frame, *Context) bool

func (f WsMatchFn) Match(frame ws.Frame, ctx *Context) bool {
	return f(frame, ctx)
}

func WsHostIs(hosts ...string) WsMatchFn {
	match := make(map[string]struct{})
	for _, host := range hosts {
		match[host] = struct{}{}
	}

	return func(frame ws.Frame, ctx *Context) bool {
		_, ok := match[ctx.DstHost]
		return ok
	}
}

type RawMatcher interface {
	Match(raw []byte, ctx *Context) bool
}

type RawFilter struct {
	cfg     *Config
	matcher []RawMatcher
}

func (c *Config) WithRawMatcher(matcher ...RawMatcher) *RawFilter {
	return &RawFilter{cfg: c, matcher: matcher}
}

type RawHandlerFn func([]byte, *Context) []byte

func (r *RawFilter) Do(handle RawHandlerFn) {
	r.cfg.rawHandlers = append(r.cfg.rawHandlers,
		func(raw []byte, ctx *Context) []byte {
			for _, matcher := range r.matcher {
				if !matcher.Match(raw, ctx) {
					return raw
				}
			}
			return handle(raw, ctx)
		})
}

func (c *Config) filterRaw(raw []byte, ctx *Context) []byte {
	for _, handle := range c.rawHandlers {
		raw = handle(raw, ctx)
	}
	return raw
}

type RawMatchFn func([]byte, *Context) bool

func (f RawMatchFn) Match(raw []byte, ctx *Context) bool {
	return f(raw, ctx)
}

func RawHostIs(hosts ...string) RawMatchFn {
	match := make(map[string]struct{})
	for _, host := range hosts {
		match[host] = struct{}{}
	}

	return func(raw []byte, ctx *Context) bool {
		_, ok := match[ctx.DstHost]
		return ok
	}
}
