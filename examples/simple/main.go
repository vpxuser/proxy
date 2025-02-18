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
	config := proxy.NewHttpProxy()
	config.Port = "8080"
	config.Serve(&proxy.Manual{})
}

func loadCert(config *proxy.HttpProxy) {
	cert, err := os.ReadFile("config/ca.crt")
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(cert)

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}

	config.Cert = certificate
}

func loadPrivateKey(config *proxy.HttpProxy) {
	key, err := os.ReadFile("config/ca.key")
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(key)

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	config.Key = privateKey
}

func snifferHTTP(config *proxy.HttpProxy) {
	config.OnRequest().Do(func(req *http.Request, ctx *proxy.Context) (*http.Request, *http.Response) {
		reqRaw, err := httputil.DumpRequest(req, true)
		if err != nil {
			log.Error(err)
			return req, nil
		}
		log.Debugf("HTTP Request : \n%s", reqRaw)
		return req, nil
	})

	config.OnResponse().Do(func(resp *http.Response, ctx *proxy.Context) *http.Response {
		respRaw, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Error(err)
			return resp
		}
		log.Debugf("HTTP Response : \n%s", respRaw)
		return resp
	})
}

func snifferWebSocket(config *proxy.HttpProxy) {
	config.OnWebSocket().Do(func(frame ws.Frame, reverse bool, ctx *proxy.Context) ws.Frame {
		log.Debugf("WebSocket Frame Payload : %s", frame.Payload)
		return frame
	})
}

func snifferTCPRaw(config *proxy.HttpProxy) {
	config.OnRaw().Do(func(raw []byte, reverse bool, ctx *proxy.Context) []byte {
		log.Debugf("TCP Raw : %s", raw)
		return raw
	})
}
