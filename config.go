package proxy

import (
	"crypto/tls"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	limiter         Limiter
	negotiator      Negotiator
	Resolver        Resolver
	Dispatcher      Dispatcher
	TLSConfig       TLSConfig
	HttpHandler     HttpHandler
	WsHandler       WsHandler
	TcpHandler      TcpHandler
	client          *http.Client
	forward         proxy.Dialer
	ClientTLSConfig *tls.Config
	DefaultSAN      string
	reqHandlers     []ReqHandlerFn
	respHandlers    []RespHandlerFn
	wsHandlers      []WsHandlerFn
	rawHandlers     []RawHandlerFn
}

func (c *Config) SetLimiter(l Limiter)       { c.limiter = l }
func (c *Config) SetNegotiator(n Negotiator) { c.negotiator = n }
func (c *Config) GetClient() *http.Client    { return c.client }
func (c *Config) GetDialer() proxy.Dialer    { return c.forward }

func (c *Config) SetProxy(rawURL string) error {
	//解析url
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	//为http客户端设置代理
	transport := c.client.Transport.(*http.Transport)
	transport.Proxy = http.ProxyURL(u)

	//为tcp拨号器设置代理
	c.forward, err = proxy.FromURL(u, c.forward)
	return err
}

func NewConfig() *Config {
	httpClient := http.DefaultClient
	if httpClient.Transport == nil {
		httpClient.Transport = &http.Transport{
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			ExpectContinueTimeout: 5 * time.Second,
		}
	}

	return &Config{
		Resolver:        defaultResolver,
		Dispatcher:      defaultDispatcher,
		HttpHandler:     defaultHttpHandler,
		WsHandler:       defaultWsHandler,
		TcpHandler:      defaultTcpHandler,
		forward:         new(net.Dialer),
		client:          httpClient,
		ClientTLSConfig: &tls.Config{InsecureSkipVerify: true},
	}
}
