package proxy

import (
	"context"
	yaklog "github.com/yaklang/yaklang/common/log"
	"io"
	"net"
)

type hook struct {
	w       io.Writer
	h       *HttpProxy
	reverse bool
	ctx     *Context
}

func (h *hook) Write(p []byte) (n int, err error) {
	_, err = h.w.Write(h.h.filterRaw(p, h.reverse, h.ctx))
	return len(p), err
}

func (h *hook) copy() *hook {
	return &hook{
		h:       h.h,
		reverse: h.reverse,
		ctx:     h.ctx,
	}
}

func (h *HttpProxy) handleTCP(client net.Conn, ctx *Context) error {
	remote, err := net.Dial("tcp", net.JoinHostPort(ctx.RemoteHost, ctx.RemotePort))
	if err != nil {
		yaklog.Errorf("%s connect to remote failed - %v", ctx.Preffix(), err)
		return err
	}
	defer remote.Close()

	errChan := make(chan error)
	cp := func(dst io.Writer, src io.Reader) {
		_, err = io.Copy(dst, src)
		errChan <- err
	}

	ho := &hook{h: h, ctx: ctx}

	remoteHook := ho.copy()
	remoteHook.w, remoteHook.reverse = remote, false

	go cp(remoteHook, client)

	clientHook := ho.copy()
	clientHook.w, clientHook.reverse = client, true

	go cp(clientHook, remote)

	<-errChan
	return nil
}

func (h *HttpProxy) handleTcp(addr string, client net.Conn, ctx *Context) {
	remote, err := net.Dial("tcp", addr)
	if err != nil {
		yaklog.Error(err)
		return
	}
	defer remote.Close()

	ho := &hook{h: h, ctx: ctx}

	remoteHook := ho.copy()
	remoteHook.w, remoteHook.reverse = remote, false
	signal, cancel := context.WithCancel(context.Background())

	go tcpCopy(remoteHook, client, cancel)

	clientHook := ho.copy()
	clientHook.w, clientHook.reverse = client, true

	go tcpCopy(clientHook, remote, cancel)

	<-signal.Done()
}

func tcpCopy(dst io.Writer, src io.Reader, cancel context.CancelFunc) {
	if _, err := io.Copy(dst, src); err != nil {
		cancel()
	}
}
