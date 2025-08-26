package main

import (
	"bufio"
	"fmt"
	"github.com/vpxuser/proxy"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
)

//func main() {
//	proxy.SetLogLevel(proxy.TraceLevel)
//	listen, err := net.Listen("tcp", ":8084")
//	if err != nil {
//		proxy.Fatal(err)
//		return
//	}
//	defer listen.Close()
//	for {
//		inner, err := listen.Accept()
//		if err != nil {
//			proxy.Error(err)
//			continue
//		}
//		client := proxy.NewConn(inner)
//		//client := inner
//		go func() {
//			req, err := http.ReadRequest(bufio.NewReader(client))
//			if err != nil {
//				proxy.Error(err)
//				return
//			}
//
//			if req.Method == http.MethodConnect {
//				status := "Connection Established"
//				resp := fmt.Sprintf("%s %d %s\r\n\r\n",
//					req.Proto, http.StatusOK, status)
//				_, err = client.Write([]byte(resp))
//				if err != nil {
//					proxy.Error(err)
//					return
//				}
//			}
//
//			dstHost := req.URL.Hostname()
//			dstPort := req.URL.Port()
//			if dstPort == "" {
//				switch req.Method {
//				case http.MethodConnect:
//					dstPort = "443"
//				default:
//					dstPort = "80"
//				}
//			}
//
//			dstConn, err := net.Dial("tcp", dstHost+":"+dstPort)
//			if err != nil {
//				proxy.Error(err)
//				return
//			}
//			defer dstConn.Close()
//
//			if req.Method != http.MethodConnect {
//				err := req.Write(dstConn)
//				if err != nil {
//					proxy.Error(err)
//					return
//				}
//			}
//
//			wg := new(sync.WaitGroup)
//			cp := func(wg *sync.WaitGroup, dst, src net.Conn, str string) {
//				defer wg.Done()
//				n, err := io.Copy(dst, src)
//				proxy.Errorf("%s %d %v", str, n, err)
//			}
//
//			wg.Add(2)
//			go cp(wg, dstConn, client, fmt.Sprintf("%s => %s:%s", client.RemoteAddr(), dstHost, dstPort))
//			go cp(wg, client, dstConn, fmt.Sprintf("%s:%s => %s", dstHost, dstPort, client.RemoteAddr()))
//			wg.Wait()
//		}()
//	}
//}

func main() {
	proxy.SetLogLevel(proxy.TraceLevel)

	cfg := proxy.NewConfig(proxy.FromCA(Cert, Key))
	cfg.DefaultSNI = Cfg.SAN
	cfg.ClientTLSConfig.InsecureSkipVerify = true

	cfg.Dispatcher = tcpForward

	cfg.WithReqMatcher().Handle(func(req *http.Request, ctx *proxy.Context) (*http.Request, *http.Response) {
		dump, err := httputil.DumpRequest(req, true)
		if err != nil {
			ctx.Error(err)
			dump, _ = httputil.DumpRequest(req, false)
		}
		ctx.Infof("是否为TLS：%v,\n%s", ctx.Conn.IsTLS(), dump)
		return req, nil
	})

	cfg.WithRespMatcher().Handle(func(resp *http.Response, ctx *proxy.Context) *http.Response {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			ctx.Error(err)
			dump, _ = httputil.DumpResponse(resp, false)
		}
		ctx.Infof("\n%s", dump)
		return resp
	})

	//cfg.WithWsMatcher().Handle(func(frame ws.Frame, ctx *proxy.Context) ws.Frame {
	//	ctx.Infof("\n%s", frame.Payload)
	//	return frame
	//})
	//
	//cfg.WithRawMatcher().Handle(func(raw []byte, ctx *proxy.Context) []byte {
	//	ctx.Infof("\n%s", raw)
	//	return raw
	//})

	if err := proxy.ListenAndServe(net.JoinHostPort(Cfg.Host, Cfg.Port), cfg); err != nil {
		proxy.Fatal(err)
	}
}

var tcpForward proxy.DispatchFn = func(ctx *proxy.Context) error {
	_, err := http.ReadRequest(bufio.NewReader(ctx.Conn.PeekRd))
	if err != nil {
		raw, err := ctx.Conn.Peek(2)
		if err != nil && len(raw) > 0 {
			ctx.Error(err)
			return err
		}
		return ctx.TcpHandler.HandleTcp(ctx)
	}
	return ctx.HttpHandler.HandleHttp(ctx)
}

var ioCopyForward proxy.DispatchFn = func(ctx *proxy.Context) error {
	dstConn, err := net.Dial("tcp", ctx.DstHost+":"+ctx.DstPort)
	if err != nil {
		ctx.Error(err)
		return err
	}
	defer dstConn.Close()

	wg := new(sync.WaitGroup)
	cp := func(dst, src net.Conn, str string) {
		defer wg.Done()
		n, err := io.Copy(dst, io.TeeReader(src, os.Stdout))
		ctx.Errorf("%s %d %v", str, n, err)
	}

	wg.Add(2)
	go cp(dstConn, ctx.Conn, fmt.Sprintf("%s => %s:%s", ctx.Conn.RemoteAddr(), ctx.DstHost, ctx.DstPort))
	go cp(ctx.Conn, dstConn, fmt.Sprintf("%s:%s => %s", ctx.DstHost, ctx.DstPort, ctx.Conn.RemoteAddr()))
	wg.Wait()

	return nil
}
