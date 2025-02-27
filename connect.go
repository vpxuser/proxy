package proxy

import (
	"bufio"
	"crypto/tls"
	"github.com/elazarl/goproxy"
	"github.com/inconshreveable/go-vhost"
	yaklog "github.com/yaklang/yaklang/common/log"
	"net"
	"net/http"
	"strings"
)

type ConnCond interface {
	Match(conn net.Conn, ctx *Context) bool
}

type ConnConds struct {
	httpProxy *HttpProxy
	conds     []ConnCond
}

func (h *HttpProxy) OnConnect(conds ...ConnCond) *ConnConds {
	return &ConnConds{httpProxy: h, conds: conds}
}

type HandleConn func(conn net.Conn, h *HttpProxy, ctx *Context) //todo

func (c *ConnConds) Handle(handle HandleConn) {
	c.httpProxy.hijack = func(conn net.Conn, h *HttpProxy, ctx *Context) {
		for _, cond := range c.conds {
			if !cond.Match(conn, ctx) {
				c.httpProxy.direct(conn, h, ctx)
				return
			}
		}
		handle(conn, h, ctx)
	}
}

type ConnMatch func(conn net.Conn, ctx *Context) bool

func (c ConnMatch) Match(conn net.Conn, ctx *Context) bool {
	return c(conn, ctx)
}

func RemoteHostIs(hosts ...string) ConnMatch {
	hostSet := make(map[string]struct{})
	for _, host := range hosts {
		hostSet[host] = struct{}{}
	}

	return func(conn net.Conn, ctx *Context) bool {
		_, ok := hostSet[ctx.RemoteHost]
		return ok
	}
}

type ConnectMode func(conn net.Conn, h *HttpProxy, ctx *Context) (err error)

func (h *HttpProxy) Service(handle ConnectMode) {
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

			_ = handle(client, h, ctx)
		}(client)
	}
}

var HttpManual ConnectMode = func(client net.Conn, h *HttpProxy, ctx *Context) (err error) {
	ctx.ThirdCtx = true

	ctx.Request, err = http.ReadRequest(bufio.NewReader(client))
	if err != nil {
		yaklog.Errorf("%s read http request failed - %v", ctx.Preffix(), err)
		return err
	}

	ctx.Request.URL.Scheme, ctx.Protocol = "http", "HTTP"

	ctx.RemoteHost, ctx.RemotePort, err = net.SplitHostPort(ctx.Request.Host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			ctx.RemoteHost, ctx.RemotePort = ctx.Request.Host, "80"
		} else {
			yaklog.Errorf("%s split remote host failed - %v", ctx.Preffix(), err)
			return err
		}
	}

	h.hijack(client, h, ctx)

	return nil
}

var HttpMitm HandleConn = func(client net.Conn, h *HttpProxy, ctx *Context) {
	if ctx.Request.Method == http.MethodConnect {
		ctx.IsTLS, ctx.Request.URL.Scheme, ctx.Protocol = true, "https", "HTTPS"

		if _, err := client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
			yaklog.Errorf("%s write http response failed - %v", ctx.Preffix(), err)
			return
		}

		subConf := goproxy.NewProxyHttpServer()
		//proxyConf.Tr.DisableCompression = true

		tlsConfig, err := goproxy.TLSConfigFromCA(&tls.Certificate{
			Certificate: [][]byte{h.Cert.Raw},
			PrivateKey:  h.Key,
		})(ctx.RemoteHost, &goproxy.ProxyCtx{
			Proxy: subConf,
		})

		client = tls.Server(client, tlsConfig)

		ctx.Request, err = http.ReadRequest(bufio.NewReader(client))
		if err != nil {
			yaklog.Errorf("%s read https request failed - %v", ctx.Preffix(), err)
			return
		}
	}

	if ctx.Request.Header.Get("Upgrade") == "websocket" {
		_ = h.handleWebSocket(client, ctx)
	} else {
		_ = h.handleHttp(client, ctx)
	}
}

var Direct HandleConn = func(client net.Conn, h *HttpProxy, ctx *Context) {
	switch ctx.ThirdCtx.(type) {
	case bool:
		if ctx.ThirdCtx.(bool) {
			var err error

			ctx.Response, err = h.HTTPClient.Do(ctx.Request)
			if err != nil {
				yaklog.Errorf("%s send http request to remote failed - %v", ctx.Preffix(), err)
				return
			}

			if err = ctx.Response.Write(client); err != nil {
				yaklog.Errorf("%s send response to client failed - %v", ctx.Preffix(), err)
				return
			}
		}
	}

	_ = h.handleTCP(client, ctx)
}

var HttpTransparent ConnectMode = func(client net.Conn, h *HttpProxy, ctx *Context) (err error) {
	ctx.Request, err = http.ReadRequest(bufio.NewReader(client))
	if err != nil {
		yaklog.Errorf("%s read http connect request failed - %v", ctx.Preffix(), err)
		return err
	}

	if ctx.Request.Method == http.MethodConnect {
		if _, err = client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
			yaklog.Errorf("%s write http connect response failed - %v", ctx.Preffix(), err)
			return err
		}
	}

	ctx.RemoteHost, ctx.RemotePort, err = net.SplitHostPort(ctx.Request.Host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			ctx.RemoteHost, ctx.RemotePort = ctx.Request.Host, "80"
		} else {
			yaklog.Errorf("%s split remote host failed - %v", ctx.Preffix(), err)
			return err
		}
	}

	h.hijack(client, h, ctx)

	return nil
}

var HttpTpMitm HandleConn = func(client net.Conn, h *HttpProxy, ctx *Context) {
	remoteIp := ctx.RemoteHost

	ioClient, buf := NewConn(client), make([]byte, 3)
	if _, err := ioClient.Reader.Read(buf); err != nil {
		yaklog.Errorf("%s peek buf failed - %v", ctx.Preffix(), err)
		return
	}

	var err error

	if buf[0] == 0x16 {
		ctx.IsTLS, ctx.Protocol = true, "TLS"

		var (
			vhostClient *vhost.TLSConn
			fqdn        string
			servName    string
		)

		if IsDomain(remoteIp) {
			fqdn = remoteIp
		} else {
			if host, ok := NsLookup.Load(remoteIp); ok {
				fqdn = host.(string)
			} else {
				vhostClient, err = vhost.TLS(ioClient)
				if err != nil {
					yaklog.Errorf("%s parse client hello failed - %v", ctx.Preffix(), err)
					return
				}

				fqdn = vhostClient.Host()
			}
		}

		if fqdn != "" {
			if host, ok := ServName.Load(fqdn); ok {
				servName = host.(string)
			} else {
				servName = fetchDNS(fqdn, ctx.RemotePort)
			}
		} else {
			servName = h.DefaultSNI
		}

		ctx.ServName = servName

		subConfig := goproxy.NewProxyHttpServer()

		tlsConfig, err := goproxy.TLSConfigFromCA(&tls.Certificate{
			Certificate: [][]byte{h.Cert.Raw},
			PrivateKey:  h.Key,
		})(servName, &goproxy.ProxyCtx{
			Proxy: subConfig,
		})

		if vhostClient != nil {
			ioClient = NewConn(tls.Server(vhostClient, tlsConfig))
		} else {
			ioClient = NewConn(tls.Server(ioClient, tlsConfig))
		}

		if _, err = ioClient.Reader.Read(buf); err != nil {
			yaklog.Errorf("%s remote server name invalid - %v", ctx.Preffix(), err)
			return
		}
	}

	if _, ok := HttpMethod[string(buf)]; ok {
		ctx.Request, err = http.ReadRequest(bufio.NewReader(ioClient))
		if err != nil {
			yaklog.Errorf("%s read http request failed - %v", ctx.Preffix(), err)
			return
		}

		ctx.RemoteHost, _, err = net.SplitHostPort(ctx.Request.Host)
		if err != nil {
			if strings.Contains(err.Error(), "missing port in address") {
				ctx.RemoteHost = ctx.Request.Host
			}
		}

		if ctx.IsTLS {
			if _, ok = NsLookup.Load(remoteIp); !ok && !IsDomain(remoteIp) {
				NsLookup.Store(remoteIp, ctx.RemoteHost)
			}

			if _, ok = ServName.Load(ctx.RemoteHost); !ok {
				ServName.Store(ctx.RemoteHost, ctx.ServName)
			}
		}

		if ctx.Request.Header.Get("Upgrade") == "websocket" {
			_ = h.handleWebSocket(ioClient, ctx)
		} else {
			ctx.Protocol, ctx.Request.URL.Scheme = "HTTP", "http"
			_ = h.handleHttp(ioClient, ctx)
		}
	} else {
		ctx.RemoteHost = remoteIp
		_ = h.handleTCP(ioClient, ctx)
	}
}
