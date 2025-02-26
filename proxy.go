package proxy

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"github.com/gobwas/ws"
	yaklog "github.com/yaklang/yaklang/common/log"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"time"
)

type HandleConn func(conn net.Conn, ctx *Context) net.Conn

type HandleReq func(req *http.Request, ctx *Context) (*http.Request, *http.Response)

type HandleResp func(resp *http.Response, ctx *Context) *http.Response

type HandleWebSocket func(frame ws.Frame, reverse bool, ctx *Context) ws.Frame

type HandleRaw func(raw []byte, reverse bool, ctx *Context) []byte

type HttpProxy struct {
	Host              string
	Port              string
	Threads           int
	Cert              *x509.Certificate
	Key               *rsa.PrivateKey
	DefaultSNI        string
	HTTPClient        *http.Client
	Dialer            proxy.Dialer
	connHandlers      []HandleConn
	reqHandlers       []HandleReq
	respHandlers      []HandleResp
	webSocketHandlers []HandleWebSocket
	rawHandlers       []HandleRaw
	WhiteList         map[string]struct{}
}

func NewHttpProxy() *HttpProxy {
	return &HttpProxy{
		Host:    "0.0.0.0",
		Port:    "1080",
		Threads: 100,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   15 * time.Second,
					KeepAlive: 15 * time.Second,
				}).DialContext,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				ForceAttemptHTTP2: false,
				//DisableCompression: false,
			},
		},
		WhiteList: make(map[string]struct{}),
	}
}

const (
	MODE_ALL = iota
	MODE_WITHOUT_HANDLER
)

func (h *HttpProxy) Copy(mode int) *HttpProxy {
	httpProxy := &HttpProxy{
		Host:       h.Host,
		Port:       h.Port,
		Threads:    h.Threads,
		Cert:       h.Cert,
		Key:        h.Key,
		DefaultSNI: h.DefaultSNI,
		HTTPClient: h.HTTPClient,
		Dialer:     h.Dialer,
	}
	switch mode {
	case MODE_ALL:
		httpProxy.reqHandlers = h.reqHandlers
		httpProxy.respHandlers = h.respHandlers
		httpProxy.webSocketHandlers = h.webSocketHandlers
		httpProxy.rawHandlers = h.rawHandlers
	case MODE_WITHOUT_HANDLER:
	}
	return httpProxy
}

func (h *HttpProxy) Serve(mode Mode) {
	httpProxy, err := net.Listen("tcp", h.Host+":"+h.Port)
	if err != nil {
		yaklog.Fatalf("listen %s failed", h.Host+":"+h.Port)
	}
	yaklog.Infof("listen %s success", h.Host+":"+h.Port)

	threads := make(chan struct{}, h.Threads)

	for {
		ctx := NewContext()

		client, err := httpProxy.Accept()
		if err != nil {
			yaklog.Errorf("%s accept client connection failed - %v", ctx.Preffix(), err)
			continue
		}

		ctx.ClientAddr = client.RemoteAddr().String()

		yaklog.Infof("%s accept %s connection success", ctx.Preffix(), ctx.ClientAddr)

		threads <- struct{}{}
		go func(client net.Conn) {
			defer func() {
				_ = client.Close()
				<-threads
			}()

			_ = mode.HandleConnect(client, h, ctx)
		}(client)
	}
}

func (h *HttpProxy) SetWhiteList(hosts ...string) {
	for _, host := range hosts {
		h.WhiteList[host] = struct{}{}
	}
}
