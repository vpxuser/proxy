package proxy

import (
	"github.com/google/uuid"
	"net"
	"strings"
)

func ListenAndServe(addr string, cfg *Config) error {
	inner, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return Serve(inner, cfg)
}

func Serve(ln net.Listener, cfg *Config) error {
	return NewListener(ln, cfg).Serve()
}

func (ln *Listener) Serve() error {
	defer ln.Close()
	for {
		id := uuid.New().String()
		id = strings.ReplaceAll(id, "-", "")
		ctx := NewContext(ctxLogger, id[:16], ln.cfg)
		inner, err := ln.Accept()
		if err != nil {
			ctx.Error(err)
			continue
		}
		ctx.Conn = NewConn(inner)
		go func() {
			defer ctx.Conn.Close()
			if ctx.Negotiator != nil { //代理协议握手
				err = ctx.Negotiator.Handshake(ctx)
				if err != nil {
					ctx.Error(err)
					return
				}
			}
			_ = ctx.Dispatcher.Dispatch(ctx)
		}()
	}
}
