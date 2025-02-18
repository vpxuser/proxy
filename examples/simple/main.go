package main

import (
	"crypto/x509"
	"encoding/pem"
	"github.com/gobwas/ws"
	"github.com/vpxuser/proxy"
	"github.com/yaklang/yaklang/common/log"
	"net/http"
	"net/http/httputil"
	"os"
)

func main() {
	// init proxy server config
	httpProxy := proxy.NewHttpProxy()

	// set proxy server host and port
	httpProxy.Host = "0.0.0.0"
	httpProxy.Port = "8080"

	// -----------------------------
	cert, err := os.ReadFile("config/ca.crt")
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(cert)

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}

	// set mitm ca certificate
	httpProxy.Cert = certificate

	//--------------------------------
	key, err := os.ReadFile("config/ca.key")
	if err != nil {
		panic(err)
	}

	block, _ = pem.Decode(key)

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	// set mitm ca privatekey
	httpProxy.Key = privateKey

	// set http request handler witch conditions
	httpProxy.OnRequest(proxy.ReqHostIs("www.baidu.com", "www.google.com"),
		proxy.ReqWildcardIs("*.qq.com", "*.aliyun.com")).
		Do(func(req *http.Request, ctx *proxy.Context) (*http.Request, *http.Response) {
			reqRaw, err := httputil.DumpRequest(req, true)
			if err != nil {
				log.Error(err)
				return req, nil
			}
			log.Debugf("HTTP Request : \n%s", reqRaw)
			return req, nil
		})

	// set http response handler
	httpProxy.OnResponse().Do(func(resp *http.Response, ctx *proxy.Context) *http.Response {
		respRaw, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Error(err)
			return resp
		}
		log.Debugf("HTTP Response : \n%s", respRaw)
		return resp
	})

	// set websocket handler
	httpProxy.OnWebSocket().Do(func(frame ws.Frame, reverse bool, ctx *proxy.Context) ws.Frame {
		log.Debugf("WebSocket Frame Payload : %s", frame.Payload)
		return frame
	})

	// set tcp raw handler
	httpProxy.OnRaw().Do(func(raw []byte, reverse bool, ctx *proxy.Context) []byte {
		log.Debugf("TCP Raw : %s", raw)
		return raw
	})

	httpProxy.Serve(&proxy.Manual{})
}
