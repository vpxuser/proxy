package proxy

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/inconshreveable/go-vhost"
	"io"
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
	req, err := http.ReadRequest(bufio.NewReader(ctx.Conn.GetTeeReader()))
	if err != nil {
		if !errors.Is(err, io.EOF) &&
			!errors.Is(err, io.ErrUnexpectedEOF) {
			return err //非解析错误，直接退出
		}
		return extraDispatcher(ctx)
	}

	//代理协议协商器为空，说明连接使用默认的http代理
	if ctx.negotiator == nil {
		if ctx.DstHost == "" {
			ctx.DstHost = req.URL.Hostname()
		}

		if ctx.DstPort == "" {
			ctx.DstPort = req.URL.Port()
			if ctx.DstPort == "" {
				switch req.URL.Scheme {
				case "http":
					ctx.DstPort = "80"
				case "https":
					ctx.DstPort = "443"
				}
			}
		}
	}

	switch {
	case req.Method == http.MethodConnect: //识别连接进入HTTP代理传输模式
		//代理协议握手
		_, err := http.ReadRequest(bufio.NewReader(ctx.Conn))
		if err != nil {
			ctx.Error(err)
			return err
		}

		status := "Connection established"
		resp := fmt.Sprintf("%s %d %s\r\n\r\n",
			req.Proto, http.StatusOK, status)
		_, err = ctx.Conn.Write([]byte(resp))
		if err != nil {
			ctx.Error(err)
			return err
		}

		//交给扩展调度器处理
		return extraDispatcher(ctx)
	case websocket.IsWebSocketUpgrade(req): //识别连接升级为WebSocket
		return ctx.WsHandler.HandleWs(ctx)
	default:
		return ctx.HttpHandler.HandleHttp(ctx)
	}
}

// 扩展调度器，基于其他封装协议的解析
var extraDispatcher DispatchFn = func(ctx *Context) error {
	//预读取数据，默认预读取1024字节
	raw, err := ctx.Conn.Peek(1024)
	if err != nil && len(raw) <= 0 {
		ctx.Error(err)
		return err
	}

	//tls协议解析,基于前两个字节识别tls协议
	if raw[0] == 0x16 && raw[1] == 0x03 {
		//获取SNI
		serverName := ctx.DstHost
		if !IsDomain(serverName) { //如果目标地址不是域名，手动提取SNI
			if record, ok := ctx.Resolver.GetPTR(serverName); ok { //检查反向解析缓存是否存储了IP对于的域名
				serverName = record
			} else { //通过ClientHello提取SNI
				rawConn, err := vhost.TLS(ctx.Conn)
				if err == nil {
					serverName = rawConn.Host()
					ctx.Resolver.SetPTR(ctx.DstHost, serverName)
					ctx.Conn = NewConn(rawConn)
				} else { //解析出错，使用默认设定的SNI
					serverName = ctx.DefaultSAN
					if serverName == "" { //默认SNI为空，无法进行中间人攻击，直接使用TCP直连，放弃中间人攻击
						return ctx.TcpHandler.HandleTcp(ctx)
					}

					ctx.Warn("No SNI provided, using fallback cert")
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
	}

	//todo后续追加其他封装协议中间人攻击

	//当前中间人只支持http、websocket和tcp
	for {
		reader := bufio.NewReader(ctx.Conn.GetTeeReader())
		req, err := http.ReadRequest(reader)
		switch {
		case err != nil:
			if !errors.Is(err, io.EOF) &&
				!errors.Is(err, io.ErrUnexpectedEOF) {
				//其他错误，判断连接为非http、websocket协议，直接转发到tcp处理器
				return ctx.TcpHandler.HandleTcp(ctx)
			}
			return err //对端连接关闭，直接结束进程
		case websocket.IsWebSocketUpgrade(req):
			return ctx.WsHandler.HandleWs(ctx)
		default:
			err = ctx.HttpHandler.HandleHttp(ctx)
			if err != nil { //错误不为空，结束循环
				return err
			}
		}
	}
}
