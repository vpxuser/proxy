package proxy

import (
	"errors"
	"net"
	"strconv"
)

type Negotiator interface {
	Handshake(*Context) error
}

type HandshakeFn func(*Context) error

func (f HandshakeFn) Handshake(ctx *Context) error { return f(ctx) }

var Socks5Handler HandshakeFn = func(ctx *Context) error { return socks5Handshake(ctx) }

func socks5Handshake(ctx *Context) error {
	buf := make([]byte, 2)
	if _, err := ctx.Conn.Read(buf); err != nil {
		return err
	}

	if buf[0] != 0x05 {
		return errors.New("unsupported version")
	}

	methods := make([]byte, buf[1])
	if _, err := ctx.Conn.Read(methods); err != nil {
		return err
	}

	if _, err := ctx.Conn.Write([]byte{0x05, 0x00}); err != nil {
		return err
	}

	buf = make([]byte, 4)
	if _, err := ctx.Conn.Read(buf); err != nil {
		return err
	}

	if buf[0] != 0x05 || buf[1] != 0x01 {
		return errors.New("unsupported request")
	}

	ctx.Tracef("Parsing SOCKS5 request: VER=0x%02X CMD=0x%02X", buf[0], buf[1])

	switch buf[3] {
	case 0x01:
		ipv4 := make([]byte, 4)
		if _, err := ctx.Conn.Read(ipv4); err != nil {
			return err
		}

		ctx.DstHost = net.IP(ipv4).String()
	case 0x03:
		aLen := make([]byte, 1)
		if _, err := ctx.Conn.Read(aLen); err != nil {
			return err
		}

		domain := make([]byte, aLen[0])
		if _, err := ctx.Conn.Read(domain); err != nil {
			return err
		}

		ctx.DstHost = string(domain)
	case 0x04:
		ipv6 := make([]byte, 16)
		if _, err := ctx.Conn.Read(ipv6); err != nil {
			return err
		}

		ctx.DstHost = net.IP(ipv6).String()
	default:
		return errors.New("unsupported address type")
	}

	port := make([]byte, 2)
	if _, err := ctx.Conn.Read(port); err != nil {
		return err
	}

	ctx.DstPort = strconv.Itoa(int(port[0])<<8 | int(port[1]))

	_, _ = ctx.Conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	return nil
}
