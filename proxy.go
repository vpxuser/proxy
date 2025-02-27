package proxy

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/gobwas/ws"
	yaklog "github.com/yaklang/yaklang/common/log"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"time"
)

var CA_CERTIFICATE *x509.Certificate

const CERTIFICATE = `-----BEGIN CERTIFICATE-----
MIIDuzCCAqOgAwIBAgIQBbdO7vYmnqfQFuurcuTtdDANBgkqhkiG9w0BAQsFADB2
MQswCQYDVQQGEwJVUzENMAsGA1UECBMEVXRhaDENMAsGA1UEBxMETGVoaTEXMBUG
A1UEChMORGlnaUNlcnQsIEluYy4xGTAXBgNVBAMTEHd3dy5kaWdpY2VydC5jb20x
FTATBgNVBAUTDDUyOTk1MzctMDE0MjAeFw0yNDA5MDkxMjM1MjdaFw0yNTA5MDkx
MjM1MjdaMHYxCzAJBgNVBAYTAlVTMQ0wCwYDVQQIEwRVdGFoMQ0wCwYDVQQHEwRM
ZWhpMRcwFQYDVQQKEw5EaWdpQ2VydCwgSW5jLjEZMBcGA1UEAxMQd3d3LmRpZ2lj
ZXJ0LmNvbTEVMBMGA1UEBRMMNTI5OTUzNy0wMTQyMIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEAt7+4AMVcG3L2J2VetGobBLqVgUOitr81Y2JNo9HQYjWM
e6WVaB2p/4YFaIx9VojgmnQQiHPjUdJQt4nNNhRUtiB3dgedyCL80vH2Q/rDK/vi
EZdK1KHdW5IXsH7ZwFJ/2QYW8ynw58Q+JEMjXTjFMGsonskzF0/GmXVYUow1TJ9L
EEph6ePawe/NGL17I1qMqeuY1zJJB5gPTrArVcZiAoRH5I/hrFAxcOgbIBqDK378
NA067V/z9dOVKQOGUiVcLrKLPELp7I7xOHAd74bFFAWbxbMup4xvg5c4JuYsvadO
ot0JzMZq6JUbV3qhl9pdjxftgrNxZUzjSTLoHvHvuQIDAQABo0UwQzAOBgNVHQ8B
Af8EBAMCBaAwEgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQUS7jP4grsrU+f
HS70jdnh7id1FWswDQYJKoZIhvcNAQELBQADggEBABnS87ylRipEasYHJWSOfXjQ
spzWIVhHbJz1kXq7EHxoDPcsAiB+S579i7znBsMFHkx7eQ12egg1AbqohB8+xJkh
L1kWCW1TL+6VaMdlQYBGwZz1/3hhuTyKICLOFpX7/p0ZBbU/apPjSSkkVlvbFNC7
a2roHGw1tMVjiuMsA3iLzPJYkIWJafpKS/z3Il5bYie1etMr7kXjrMI30zZLgPlT
/ac3eihnWejzjcDQHqAzM6XM/sVGGYuvrW448y+mOlz/NjadzbIVv8j+aKyeHiWT
k/BIYu/7Qf2giuBoMox+ynJk6zTUSYNu6dyh1C7gLzCdRpn8vkFp+q7VIn7WUdU=
-----END CERTIFICATE-----`

var CA_PRIVATE_KEY *rsa.PrivateKey

const PRIVATE_KEY = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAt7+4AMVcG3L2J2VetGobBLqVgUOitr81Y2JNo9HQYjWMe6WV
aB2p/4YFaIx9VojgmnQQiHPjUdJQt4nNNhRUtiB3dgedyCL80vH2Q/rDK/viEZdK
1KHdW5IXsH7ZwFJ/2QYW8ynw58Q+JEMjXTjFMGsonskzF0/GmXVYUow1TJ9LEEph
6ePawe/NGL17I1qMqeuY1zJJB5gPTrArVcZiAoRH5I/hrFAxcOgbIBqDK378NA06
7V/z9dOVKQOGUiVcLrKLPELp7I7xOHAd74bFFAWbxbMup4xvg5c4JuYsvadOot0J
zMZq6JUbV3qhl9pdjxftgrNxZUzjSTLoHvHvuQIDAQABAoIBAQCYe0eFlEHQIYF1
xeBmVRrgvLluUKNJhbkXJS+Kv4VuykMvJISsptk4y43XaaZlVicU5TYHFixQ7PfA
p2Ec/JgjnyOjVcSwnaAyqUoUtZaa/GZo9KTLeRtJbx2rgEjRGWUwwqXu2iIpIqfy
zageJws0F+jYg9ya/r+u/zrxHZrpXmrp0uJGJEvsUCRRQ9zXzV75ks1jVK+dP2Im
3UaKR6Uyz1L+uU+MP2H0JwPlEQwqYaLzb7DUBsmktwy62nK2PtyDhPp43932lYjj
WY8r+wJ/CWjWPziwqNGlUTAb5ashp2ykjQh8L1zgaby9hLjC994cIuoPD+25jkDF
OQvRgRNRAoGBANFSd10MGo2mt/EKj3APL+va1ZofYQpj0OCfjGhzy6Ojhpe5q/sD
uUP7w98Qei6Giij21x3t7aoZbOry8MRY20Ziy8cZdYG4/PWiiS5X2xgsOSduul91
mW7qhw61+xXkgV+oqPRhw2/x1PrvkUlLgohOUMsit9OZHzeh64fZEw3dAoGBAOC5
W66fC915APYkAssO2K1c/Vf6y6S158mAi8dUEjwj2154MKtMYrlXhPeJyYziPEAW
bQfb9ff++go28XmsRcPcjGSeWG3+bbT+FYGcgjxhTjfHXW9fIYzaN0udZI7SxaRb
9+lfNO2MqXEkavN7NaPDAHxTFdO05ZQZ5kayPDGNAoGADS5zQ9HCTk9EYBJ5K+ZY
7zJNpzH4q23TtoF1lxJLrZdbn2xazyjR3t0Y3ZAtEYb5ZlD9BO55u+9z82uvC75I
uKr6CCSrSKr8iv9cQiXYLzKBuuD3LZG7QsfUD3IYSK2mE/8L/K+3XfJNpiu163as
1qaP4erixplq5Nb2fQyHbaUCgYAQq++vTsFUluuJVzaV1e4hPmrVIhgFijE987lq
+kO4Dnjx0zzZGHuigGmu65v2Rbpujrtb/+eJlHL8WwMjIbKzSyNnO5AX6O4+pTL3
QKMw0483+CRoZMhaaL39cBnLtrtO7DvCJnwIu4y+hhMhKRzbn1Xj404VPLBjgmBh
EkwA6QKBgQCEtsnHrEeXeUtSntv1z647yIyMQguNIBDlk1YA+3KfN0tVuC1S7XHU
bPeSbqF3/C1h9KWfeZmSFrvRkS+T+HG+Sv1H1FgHqXnhCCNbR/e/T6Y1E49xVvDc
xbnsbTengZU5A1W3txFmgIQh508WFV1QqObyKLgNf1LRURrLFnGiuA==
-----END RSA PRIVATE KEY-----`

func init() {
	block, _ := pem.Decode([]byte(CERTIFICATE))
	if block == nil {
		yaklog.Error("decode proxy ca_certificate failed")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		yaklog.Errorf("parse proxy ca_certificate failed - %v", err)
	} else {
		CA_CERTIFICATE = cert
	}

	block, _ = pem.Decode([]byte(PRIVATE_KEY))
	if block == nil {
		yaklog.Error("decode proxy ca_private_key failed")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		yaklog.Errorf("parse proxy ca_certificate failed - %v", err)
	} else {
		CA_PRIVATE_KEY = key
	}
}

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
	reqHandlers       []HandleReq
	respHandlers      []HandleResp
	webSocketHandlers []HandleWebSocket
	rawHandlers       []HandleRaw
	hijackSet         map[string]struct{}
}

func NewHttpProxy() *HttpProxy {
	return &HttpProxy{
		Host:    "0.0.0.0",
		Port:    "1080",
		Threads: 100,
		Cert:    CA_CERTIFICATE,
		Key:     CA_PRIVATE_KEY,
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
		hijackSet: make(map[string]struct{}),
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
		hijackSet:  h.hijackSet,
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

func (h *HttpProxy) Serve(handleConn ConnectMode) {
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

			handleConn(client, h, ctx)
		}(client)
	}
}

func (h *HttpProxy) Hijack(hosts ...string) {
	for _, host := range hosts {
		h.hijackSet[host] = struct{}{}
	}
}
