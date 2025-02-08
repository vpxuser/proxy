package proxy

import (
	"bufio"
	"crypto/tls"
	"errors"
	"github.com/gobwas/ws"
	yaklog "github.com/yaklang/yaklang/common/log"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
)

func (h *HttpProxy) handleWebSocket(client net.Conn, ctx *Context) (err error) {
	ctx.Protocol = "WebSocket"

	requestRaw, err := httputil.DumpRequest(ctx.Request, true)
	if err == nil {
		yaklog.Infof("%s [Handshake] [Request]\n%s", ctx.Preffix(false), requestRaw)
	}

	ctx.Request.Header.Del("Sec-WebSocket-Extensions")

	var remote net.Conn
	if h.Dialer != nil {
		h.Dialer.(Dialer).SetTLS(ctx.IsTLS)

		proxyConn, err := h.Dialer.Dial("tcp", net.JoinHostPort(ctx.RemoteHost, ctx.RemotePort))
		if err != nil {
			yaklog.Errorf("%s connect to proxy failed - %v", ctx.Preffix(), err)
			return err
		}
		remote = proxyConn
	} else {
		if ctx.IsTLS {
			tlsRemote, err := tls.Dial("tcp", net.JoinHostPort(ctx.RemoteHost, ctx.RemotePort), &tls.Config{InsecureSkipVerify: true})
			if err != nil {
				yaklog.Errorf("%s tls connect to remote failed - %v", ctx.Preffix(), err)
				return err
			}
			remote = tlsRemote
		} else {
			netRemote, err := net.Dial("tcp", net.JoinHostPort(ctx.RemoteHost, ctx.RemotePort))
			if err != nil {
				yaklog.Errorf("%s connect to remote failed - %v", ctx.Preffix(), err)
			}
			remote = netRemote
		}
	}

	defer func() {
		_ = remote.Close()
	}()

	if err := ctx.Request.Write(remote); err != nil {
		yaklog.Errorf("%s send request to remote failed - %v", ctx.Preffix(), err)
		return err
	}

	response, err := http.ReadResponse(bufio.NewReader(remote), ctx.Request)
	if err != nil {
		yaklog.Errorf("%s read response from remote failed - %v", ctx.Preffix(), err)
		return err
	}

	if err = response.Write(client); err != nil {
		yaklog.Errorf("%s send response to client failed - %v", ctx.Preffix(), err)
		return err
	}

	responseRaw, err := httputil.DumpResponse(response, true)
	if err == nil {
		yaklog.Infof("%s [Handshake] [Response]\n%s", ctx.Preffix(true), responseRaw)
	}

	errChan := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader, reverse bool) {
		for {
			frame, err := ws.ReadFrame(src)
			if err != nil {
				if !handleEOF(err) {
					yaklog.Errorf("%s read frame failed - %v", ctx.Preffix(reverse), err)
				}
				errChan <- err
				break
			}

			frame = h.filterWebSocket(frame, reverse, ctx)

			if err = ws.WriteFrame(dst, frame); err != nil {
				if !handleEOF(err) {
					yaklog.Errorf("%s send frame failed - %v", ctx.Preffix(reverse), err)
				}
				errChan <- err
				break
			}
		}
	}

	go cp(remote, client, false)
	go cp(client, remote, true)
	<-errChan

	return nil
}

func handleEOF(err error) (isEOF bool) {
	return err == io.EOF || errors.Is(err, net.ErrClosed)
}
