package proxy

import (
	"bufio"
	"crypto/tls"
	"github.com/gorilla/websocket"
	"github.com/inconshreveable/go-vhost"
	"net/http"
)

type Dispatcher interface {
	Dispatch(*Context) error
}

type DispatchFn func(*Context) error

func (f DispatchFn) Dispatch(ctx *Context) error { return f(ctx) }

// 调度器，基于明文TCP进行调度
var defaultDispatcher DispatchFn = func(ctx *Context) error {
	//识别TCP流数据是否为http
	req, parseErr := http.ReadRequest(bufio.NewReader(ctx.Conn.PeekRd))
	if parseErr != nil {
		//预读取数据，默认预读取1024字节
		raw, err := ctx.Conn.Peek(2)
		if err != nil && len(raw) <= 0 {
			ctx.Error(err)
			return err
		}

		//tls协议解析,基于前两个字节识别tls协议
		switch {
		case raw[0] == 0x16 && raw[1] == 0x03:
			//获取SNI
			serverName := ctx.DstHost
			if !IsDomain(serverName) { //如果目标地址不是域名，手动提取SNI
				if record, ok := ctx.Resolver.GetPTR(serverName); ok { //检查反向解析缓存是否存储了IP对于的域名
					serverName = record
				} else { //通过ClientHello提取SNI
					rawConn, err := vhost.TLS(ctx.Conn)
					if err != nil { //解析出错，使用默认设定的SNI
						serverName = ctx.DefaultSNI
						if serverName == "" { //默认SNI为空，无法进行中间人攻击，直接使用TCP直连，放弃中间人攻击
							return ctx.TcpHandler.HandleTcp(ctx)
						}
						ctx.Warn("No SNI provided, using fallback cert")
					} else {
						serverName = rawConn.Host()
						ctx.Resolver.SetPTR(ctx.DstHost, serverName)
						ctx.Conn = NewConn(rawConn)
					}
				}
			}

			//将连接审计为TLS
			tlsCfg, err := ctx.TLSConfig.From(serverName)
			if err != nil {
				ctx.Error(err)
				return err
			}
			ctx.Conn = NewConn(tls.Server(ctx.Conn, tlsCfg))

			req, err = http.ReadRequest(bufio.NewReader(ctx.Conn.PeekRd))
			if err != nil {
				return ctx.TcpHandler.HandleTcp(ctx)
			}
		default:
			return ctx.TcpHandler.HandleTcp(ctx)
		}
	}

	//当前中间人只支持http、websocket和tcp
	switch {
	case websocket.IsWebSocketUpgrade(req):
		return ctx.WsHandler.HandleWs(ctx)
	default:
		return ctx.HttpHandler.HandleHttp(ctx)
	}
}

var isHttp = map[string]struct{}{
	http.MethodGet[:3]:     {},
	http.MethodHead[:3]:    {},
	http.MethodPost[:3]:    {},
	http.MethodPut[:3]:     {},
	http.MethodPatch[:3]:   {},
	http.MethodDelete[:3]:  {},
	http.MethodConnect[:3]: {},
	http.MethodOptions[:3]: {},
	http.MethodTrace[:3]:   {},
}

// 基于透传模式下进行调度
var tproxyDispatch DispatchFn = func(ctx *Context) error {
	ctx.Debugf("进入 Dispatcher 处理流程")

	//识别TCP流数据是否为http
	raw, err := ctx.Conn.Peek(3)
	if err != nil && len(raw) == 0 {
		ctx.Error(err)
		return err
	}
	ctx.Debugf("预读取数据：%v", raw)

	ctx.Debugf("开始进行协议解析")
	var req *http.Request
	if _, ok := isHttp[string(raw)]; ok {
		ctx.Debugf("可能是 HTTP 连接")
		req, err = http.ReadRequest(bufio.NewReader(ctx.Conn.PeekRd))
		if err != nil {
			ctx.Error(err)
			return err
		}
		ctx.Debugf("提取 HTTP 请求报文")
	} else if raw[0] == 0x16 && raw[1] == 0x03 && raw[2] <= 0x04 { //tls协议解析,基于前两个字节识别tls协议
		ctx.Debugf("可能是 TLS 连接")
		//获取SNI
		serverName := ctx.DstHost
		if !IsDomain(serverName) { //如果目标地址不是域名，手动提取SNI
			if record, ok := ctx.Resolver.GetPTR(serverName); ok { //检查反向解析缓存是否存储了IP对于的域名
				ctx.Debugf("从缓存中提取 SNI")
				serverName = record
			} else { //通过ClientHello提取SNI
				ctx.Debugf("开始解析 ClientHello 并提取 SNI")
				rawConn, err := vhost.TLS(ctx.Conn)
				if err != nil { //解析出错，使用默认设定的SNI
					ctx.Debugf("提取 SNI 失败：%v", err)
					serverName = ctx.DefaultSNI
					if serverName == "" { //默认SNI为空，无法进行中间人攻击，直接使用TCP直连，放弃中间人攻击
						ctx.Debugf("默认 SNI 为空")
						return ctx.TcpHandler.HandleTcp(ctx)
					}
					ctx.Warn("No SNI provided, using fallback cert")
					ctx.Debug("使用默认 SNI")
				} else {
					ctx.Debugf("提取 SNI 成功")
					serverName = rawConn.Host()
					ctx.Resolver.SetPTR(ctx.DstHost, serverName)
					ctx.Conn = NewConn(rawConn)
				}
			}
		}

		ctx.Debugf("SNI 域名：%s", serverName)

		//将连接审计为TLS
		tlsCfg, err := ctx.TLSConfig.From(serverName)
		if err != nil {
			ctx.Error(err)
			return err
		}
		ctx.Conn = NewConn(tls.Server(ctx.Conn, tlsCfg))

		req, err = http.ReadRequest(bufio.NewReader(ctx.Conn.PeekRd))
		if err != nil {
			ctx.Errorf("预读取请求失败：%v", err)
			return ctx.TcpHandler.HandleTcp(ctx)
		}
		ctx.Debugf("提取 HTTPS 请求报文")
	} else {
		ctx.Debugf("直接进行 TCP 透传")
		return ctx.TcpHandler.HandleTcp(ctx)
	}

	//当前中间人只支持http、websocket和tcp
	switch {
	case websocket.IsWebSocketUpgrade(req):
		ctx.Debugf("WS(S) 握手请求消息摘要：%s", req.URL.String())
		return ctx.WsHandler.HandleWs(ctx)
	default:
		ctx.Debugf("HTTP(S) 请求消息摘要：%s", req.URL.String())
		return ctx.HttpHandler.HandleHttp(ctx)
	}
}
