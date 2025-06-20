package proxy

import (
	"net"
)

func ListenAndServe(network, addr string, cfg *Config) error {
	inner, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	Infof("Proxy server started on %s", addr)
	return Serve(inner, cfg)
}

func Serve(inner net.Listener, cfg *Config) error {
	return NewListener(inner, cfg).Serve()
}

func (l *Listener) Serve() error {
	defer l.Close()

	for {
		ctx := NewContext()
		ctx.Config = l.cfg

		inner, err := l.Accept()
		if err != nil {
			ctx.Error(err)
			continue
		}

		ctx.Infof("New connection from %s", inner.RemoteAddr())

		ctx.Conn = inner.(*Conn)
		go func() {
			defer inner.Close()

			if ctx.negotiator != nil {
				if err = ctx.negotiator.Handshake(ctx); err != nil {
					ctx.Error(err)
					return
				}
			}

			ctx.Debugf("Dispatching connection to %s:%s", ctx.DstHost, ctx.DstPort)

			ctx.dispatcher.Dispatch(ctx)
		}()
	}
}
