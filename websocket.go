package proxy

import (
	"bufio"
	"context"
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

func (h *HttpProxy) handleWebsocket(req *http.Request, https bool, addr string, client net.Conn, ctx *Context) {
	req.Header.Del("Sec-WebSocket-Extensions")

	remote, err := h.dialWebSocket(https, addr)
	if err != nil {
		yaklog.Error(err)
		return
	}
	defer remote.Close()

	if err = req.Write(remote); err != nil {
		yaklog.Error(err)
		return
	}

	resp, err := http.ReadResponse(bufio.NewReader(remote), req)
	if err != nil {
		yaklog.Error(err)
		return
	}

	if err = resp.Write(client); err != nil {
		yaklog.Error(err)
		return
	}

	signal, cancel := context.WithCancel(context.Background())

	go h.wsCopy(remote, client, false, ctx, signal, cancel)
	go h.wsCopy(client, remote, true, ctx, signal, cancel)
	<-signal.Done()
}

func (h *HttpProxy) dialWebSocket(https bool, addr string) (net.Conn, error) {
	if h.Dialer != nil {
		h.Dialer.(Dialer).SetTLS(https)
		return h.Dialer.Dial("tcp", addr)
	}

	if https {
		return tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	}

	return net.Dial("tcp", addr)
}

func (h *HttpProxy) wsCopy(dst io.Writer, src io.Reader, reverse bool, ctx *Context, signal context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-signal.Done():
			return
		default:
			frame, err := ws.ReadFrame(src)
			if err != nil {
				if !handleEOF(err) {
					yaklog.Error(err)
				}
				cancel()
				return
			}

			frame = h.filterWebSocket(frame, reverse, ctx)
			if err = ws.WriteFrame(dst, frame); err != nil {
				if !handleEOF(err) {
					yaklog.Error(err)
				}
				cancel()
				return
			}
		}
	}
}
