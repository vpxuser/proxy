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

type Mode interface {
	HandleConnect(client net.Conn, h *HttpProxy, ctx *Context) (err error)
}

type Manual struct{}

func (m *Manual) HandleConnect(client net.Conn, h *HttpProxy, ctx *Context) (err error) {
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

	if len(h.WhiteList) > 0 {
		if _, ok := h.WhiteList[ctx.RemoteHost]; !ok {
			return h.handleTCP(client, ctx)
		}
	}

	if ctx.Request.Method == http.MethodConnect {
		ctx.IsTLS, ctx.Request.URL.Scheme, ctx.Protocol = true, "https", "HTTPS"

		if _, err = client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
			yaklog.Errorf("%s write http response failed - %v", ctx.Preffix(), err)
			return err
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
			return err
		}
	}

	if ctx.Request.Header.Get("Upgrade") == "websocket" {
		_ = h.handleWebSocket(client, ctx)
	} else {
		_ = h.handleHttp(client, ctx)
	}

	return nil
}

var HttpMethod = map[string]struct{}{
	http.MethodGet[:3]:     {},
	http.MethodHead[:3]:    {},
	http.MethodPost[:3]:    {},
	http.MethodPut[:3]:     {},
	http.MethodPatch[:3]:   {},
	http.MethodConnect[:3]: {},
	http.MethodDelete[:3]:  {},
	http.MethodOptions[:3]: {},
	http.MethodTrace[:3]:   {},
}

type Transparent struct{}

func (t *Transparent) HandleConnect(client net.Conn, h *HttpProxy, ctx *Context) (err error) {
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

	if len(h.WhiteList) > 0 {
		if _, ok := h.WhiteList[ctx.RemoteHost]; !ok {
			return h.handleTCP(client, ctx)
		}
	}

	remoteIp := ctx.RemoteHost

	ioClient, buf := NewConn(client), make([]byte, 3)
	if _, err = ioClient.Reader.Read(buf); err != nil {
		yaklog.Errorf("%s peek buf failed - %v", ctx.Preffix(), err)
		return err
	}

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
					return err
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
			return err
		}
	}

	if _, ok := HttpMethod[string(buf)]; ok {
		ctx.Request, err = http.ReadRequest(bufio.NewReader(ioClient))
		if err != nil {
			yaklog.Errorf("%s read http request failed - %v", ctx.Preffix(), err)
			return err
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

	return nil
}
