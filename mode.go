package proxy

import (
	"bufio"
	"crypto/tls"
	"github.com/inconshreveable/go-vhost"
	yaklog "github.com/yaklang/yaklang/common/log"
	"net"
	"net/http"
	"strings"
)

var Http HandleConn = func(client net.Conn, ctx *Context) {
	req, err := http.ReadRequest(bufio.NewReader(client))
	if err != nil {
		yaklog.Errorf("%s read http request failed - %v", ctx.Preffix(), err)
		return
	}

	ctx.Request = req

	remoteHost, remotePort, err := net.SplitHostPort(req.Host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			remoteHost, remotePort = req.Host, "80"
		} else {
			yaklog.Error(err)
			return
		}
	}

	https := false
	if req.Method == http.MethodConnect {
		https = true

		if _, err = client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
			yaklog.Error(err)
			return
		}

		tlsConfig, err := ctx.HttpProxy.GetTLSConfig(remoteHost)
		if err != nil {
			yaklog.Error(err)
			return
		}

		client = tls.Server(client, tlsConfig)

		req, err = http.ReadRequest(bufio.NewReader(client))
		if err != nil {
			yaklog.Error(err)
			return
		}

		ctx.Request = req
	}

	if ctx.Request.Header.Get("Upgrade") == "websocket" {
		addr := net.JoinHostPort(remoteHost, remotePort)
		ctx.HttpProxy.handleWebsocket(req, https, addr, client, ctx)
		return
	}

	ctx.HttpProxy.handleHTTP(req, https, client, ctx)
}

func Socks5Hook(hosts ...string) HandleConn {
	todo := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		todo[host] = struct{}{}
	}

	return func(client net.Conn, ctx *Context) {
		host, port, err := Socks5Handshake(client)
		if err != nil {
			yaklog.Error(err)
			return
		}
		addr := net.JoinHostPort(host, port)

		yaklog.Debugf("[%s] origin dst addr: %s", ctx.Id, addr)

		ioClient, buf := NewConn(client), make([]byte, 3)
		if _, err = ioClient.Reader.Read(buf); err != nil {
			yaklog.Error(err)
			return
		}

		var domain, sni string
		https := false
		if buf[0] == 0x16 {
			https = true

			var vhostTLS *vhost.TLSConn
			if IsDomain(host) {
				domain = host
			} else {
				if record, ok := NsLookup.Load(host); ok {
					domain = record.(string)
				} else {
					vhostTLS, err = vhost.TLS(ioClient)
					if err != nil {
						yaklog.Error(err)
						return
					}
					domain = vhostTLS.Host()
					ioClient = NewConn(vhostTLS)
				}
			}

			if _, ok := todo[domain]; !ok && len(todo) > 0 {
				yaklog.Debugf("[%s] whitelist domain: %s", ctx.Id, domain)

				ctx.HttpProxy.handleTcp(addr, ioClient, ctx)
				return
			}

			if domain != "" {
				yaklog.Debugf("[%s] origin dst domain: %s", ctx.Id, domain)

				if record, ok := ServName.Load(domain); ok {
					sni = record.(string)
				} else {
					sni = fetchDNS(domain, ctx.RemotePort)
				}
			} else {
				yaklog.Debugf("[%s] domain is empty: %s", ctx.Id, domain)

				ctx.HttpProxy.handleTcp(addr, ioClient, ctx)
				return
			}

			tlsConfig, err := ctx.HttpProxy.GetTLSConfig(sni)
			if err != nil {
				yaklog.Error(err)
				return
			}

			ioClient = NewConn(tls.Server(ioClient, tlsConfig))
			if _, err = ioClient.Reader.Read(buf); err != nil {
				yaklog.Error(err)
				return
			}
		}

		if _, ok := HttpMethod[string(buf)]; ok {
			req, err := http.ReadRequest(bufio.NewReader(ioClient))
			if err != nil {
				yaklog.Error(err)
				return
			}

			ctx.Request = req

			domain, port, err = net.SplitHostPort(req.Host)
			if err != nil {
				if strings.Contains(err.Error(), "missing port in address") {
					domain, port = req.Host, "80"
					if https {
						port = "443"
					}
				}
				yaklog.Warn(err)
			}

			if https {
				if _, ok = NsLookup.Load(host); !ok && !IsDomain(host) {
					NsLookup.Store(host, domain)
				}

				if _, ok = ServName.Load(domain); !ok {
					ServName.Store(domain, sni)
				}
			}

			if req.Header.Get("Upgrade") == "websocket" {
				ctx.HttpProxy.handleWebsocket(req, https, addr, ioClient, ctx)
				return
			}

			ctx.HttpProxy.handleHTTP(req, https, ioClient, ctx)
			return
		}

		ctx.HttpProxy.handleTcp(addr, ioClient, ctx)
	}
}

type ConnectMode func(client net.Conn, h *HttpProxy, ctx *Context)

var Manual ConnectMode = func(client net.Conn, h *HttpProxy, ctx *Context) {
	req, err := http.ReadRequest(bufio.NewReader(client))
	if err != nil {
		yaklog.Errorf("%s read http request failed - %v", ctx.Preffix(), err)
		return
	}

	req.URL.Scheme, ctx.Protocol = "http", "HTTP"

	ctx.RemoteHost, ctx.RemotePort, err = net.SplitHostPort(req.Host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			ctx.RemoteHost, ctx.RemotePort = req.Host, "80"
		} else {
			yaklog.Errorf("%s split remote host failed - %v", ctx.Preffix(), err)
			return
		}
	}

	ctx.Request = req

	if ctx.Request.Method == http.MethodConnect {
		ctx.IsTLS, ctx.Request.URL.Scheme, ctx.Protocol = true, "https", "HTTPS"

		if _, err = client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
			yaklog.Errorf("%s write http response failed - %v", ctx.Preffix(), err)
			return
		}

		tlsConfig, err := h.GetTLSConfig(ctx.RemoteHost)
		if err != nil {
			yaklog.Errorf("%s generate tls config failed - %v", ctx.Preffix(), err)
			return
		}

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

var Transparent ConnectMode = func(client net.Conn, h *HttpProxy, ctx *Context) {
	req, err := http.ReadRequest(bufio.NewReader(client))
	if err != nil {
		yaklog.Errorf("%s read http connect request failed - %v", ctx.Preffix(), err)
		return
	}

	if req.Method == http.MethodConnect {
		if _, err = client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
			yaklog.Errorf("%s write http connect response failed - %v", ctx.Preffix(), err)
			return
		}
	}

	ctx.RemoteHost, ctx.RemotePort, err = net.SplitHostPort(req.Host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			ctx.RemoteHost, ctx.RemotePort = req.Host, "80"
		} else {
			yaklog.Errorf("%s split remote host failed - %v", ctx.Preffix(), err)
			return
		}
	}

	ctx.Request = req

	remoteIp := ctx.RemoteHost

	ioClient, buf := NewConn(client), make([]byte, 3)
	if _, err = ioClient.Reader.Read(buf); err != nil {
		yaklog.Errorf("%s peek buf failed - %v", ctx.Preffix(), err)
		return
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

		tlsConfig, err := h.GetTLSConfig(ctx.RemoteHost)
		if err != nil {
			yaklog.Errorf("%s generate tls config failed - %v", ctx.Preffix(), err)
			return
		}

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

func SelectMitmManual(hosts ...string) ConnectMode {
	hostSet := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		hostSet[host] = struct{}{}
	}

	return func(client net.Conn, h *HttpProxy, ctx *Context) {
		req, err := http.ReadRequest(bufio.NewReader(client))
		if err != nil {
			yaklog.Errorf("%s read http request failed - %v", ctx.Preffix(), err)
			return
		}

		req.URL.Scheme, req.RequestURI, ctx.Protocol = "http", "", "HTTP"

		ctx.RemoteHost, ctx.RemotePort, err = net.SplitHostPort(req.Host)
		if err != nil {
			if strings.Contains(err.Error(), "missing port in address") {
				ctx.RemoteHost, ctx.RemotePort = req.Host, "80"
			} else {
				yaklog.Errorf("%s split remote host failed - %v", ctx.Preffix(), err)
				return
			}
		}

		ctx.Request = req

		if ctx.Request.Method == http.MethodConnect {
			ctx.IsTLS, ctx.Request.URL.Scheme, ctx.Protocol = true, "https", "HTTPS"

			if _, err = client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
				yaklog.Errorf("%s write http response failed - %v", ctx.Preffix(), err)
				return
			}

			if _, ok := hostSet[ctx.RemoteHost]; !ok {
				_ = h.handleTCP(client, ctx)
				return
			}

			tlsConfig, err := h.GetTLSConfig(ctx.RemoteHost)
			if err != nil {
				yaklog.Errorf("%s generate tls config failed - %v", ctx.Preffix(), err)
				return
			}

			client = tls.Server(client, tlsConfig)

			ctx.Request, err = http.ReadRequest(bufio.NewReader(client))
			if err != nil {
				yaklog.Errorf("%s read https request failed - %v", ctx.Preffix(), err)
				return
			}
		}

		if _, ok := hostSet[ctx.RemoteHost]; ok {
			if ctx.Request.Header.Get("Upgrade") == "websocket" {
				_ = h.handleWebSocket(client, ctx)
			} else {
				_ = h.handleHttp(client, ctx)
			}
			return
		}

		resp, err := h.HTTPClient.Do(req)
		if err != nil {
			yaklog.Errorf("%s read response form remote failed - %v", ctx.Preffix(), err)
			return
		}

		if err = resp.Write(client); err != nil {
			yaklog.Errorf("%s write response to client failed - %v", ctx.Preffix(), err)
			return
		}

		_ = h.handleTCP(client, ctx)
	}
}

func Socks5(hosts ...string) ConnectMode {
	whiteList := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		whiteList[host] = struct{}{}
	}

	return func(client net.Conn, h *HttpProxy, ctx *Context) {
		remoteHost, remotePort, err := Socks5Handshake(client)
		if err != nil {
			yaklog.Errorf("%s socks5 handshake failed - %v", ctx.Preffix(), err)
			return
		}

		ctx.RemoteHost, ctx.RemotePort = remoteHost, remotePort

		ioClient, buf := NewConn(client), make([]byte, 3)
		if _, err := ioClient.Reader.Read(buf); err != nil {
			yaklog.Errorf("%s peek buf[:3] failed - %v", ctx.Preffix(), err)
			return
		}

		if buf[0] == 0x16 {
			ctx.IsTLS, ctx.Protocol = true, "TLS"

			var (
				vhostClient *vhost.TLSConn
				fqdn        string
				servName    string
			)

			if IsDomain(remoteHost) {
				fqdn = remoteHost
			} else {
				if host, ok := NsLookup.Load(remoteHost); ok {
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

			tlsConfig, err := h.GetTLSConfig(ctx.RemoteHost)
			if err != nil {
				yaklog.Errorf("%s generate tls config failed - %v", ctx.Preffix(), err)
				return
			}

			if vhostClient != nil {
				ioClient = NewConn(tls.Server(vhostClient, tlsConfig))
			} else {
				ioClient = NewConn(tls.Server(ioClient, tlsConfig))
			}

			if _, ok := whiteList[fqdn]; !ok && len(whiteList) > 0 {
				_ = h.handleTCP(ioClient, ctx)
				return
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
				if _, ok = NsLookup.Load(remoteHost); !ok && !IsDomain(remoteHost) {
					NsLookup.Store(remoteHost, ctx.RemoteHost)
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
			ctx.RemoteHost = remoteHost
			_ = h.handleTCP(ioClient, ctx)
		}
	}
}
