package proxy

import (
	"crypto/tls"
	"golang.org/x/net/proxy"
	"net"
)

type Config struct {
	Limiter         Limiter         // 限速器（可选）
	Negotiator      Negotiator      // 代理协商（HTTP、SOCKS5）
	Resolver        Resolver        // 域名解析器
	Dispatcher      Dispatcher      // 请求分发器
	DefaultSNI      string          // 默认 SNI
	TLSConfig       TLSConfig       // TLS 配置回调函数
	HttpHandler     HttpHandler     // HTTP 请求处理
	WsHandler       WsHandler       // WebSocket 处理
	TcpHandler      TcpHandler      // TCP 处理
	Dialer          proxy.Dialer    // 连接拨号器（可叠加代理）
	ClientTLSConfig *tls.Config     // 客户端 TLS 配置
	reqHandlers     []ReqHandlerFn  // 请求处理链
	respHandlers    []RespHandlerFn // 响应处理链
	wsHandlers      []WsHandlerFn   // WS 处理链
	rawHandlers     []RawHandlerFn  // 原始数据处理链
}

func NewConfig(tlsConfigFn TLSConfig) *Config {
	return &Config{
		Negotiator:      HttpNegotiator,
		Resolver:        defaultResolver,
		Dispatcher:      defaultDispatcher,
		TLSConfig:       tlsConfigFn,
		HttpHandler:     defaultHttpHandler,
		WsHandler:       defaultWsHandler,
		TcpHandler:      defaultTcpHandler,
		Dialer:          new(net.Dialer),
		ClientTLSConfig: new(tls.Config),
	}
}
