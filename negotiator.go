package proxy

import (
	"errors"
	"net"
	"strconv"
)

// Negotiator defines the handshake behavior for various protocols.
// Negotiator 接口定义了各种协议的握手行为。
type Negotiator interface {
	Handshake(*Context) error // Performs handshake negotiation. 执行握手协商。
}

// HandshakeFn is a function adapter that implements the Negotiator interface.
// HandshakeFn 是一个实现 Negotiator 接口的函数适配器。
type HandshakeFn func(*Context) error

// Handshake calls the function itself.
// Handshake 方法直接调用函数本体。
func (f HandshakeFn) Handshake(ctx *Context) error { return f(ctx) }

// Socks5Negotiator handles SOCKS5 protocol negotiation as per RFC 1928.
// Socks5Negotiator 按照 RFC 1928 实现 SOCKS5 协议握手。
var Socks5Negotiator HandshakeFn = func(ctx *Context) error {
	buf := make([]byte, 2)
	if _, err := ctx.Conn.Read(buf); err != nil {
		return err
	}

	// Check SOCKS version
	// 检查 SOCKS 版本
	if buf[0] != 0x05 {
		return errors.New("unsupported version") // 不支持的 SOCKS 版本
	}

	// Read supported authentication methods
	// 读取客户端支持的认证方法
	methods := make([]byte, buf[1])
	if _, err := ctx.Conn.Read(methods); err != nil {
		return err
	}

	// Send method selection response (no auth)
	// 回复选择的认证方法（0x00 表示不需要认证）
	if _, err := ctx.Conn.Write([]byte{0x05, 0x00}); err != nil {
		return err
	}

	// Read connection request header
	// 读取连接请求头
	buf = make([]byte, 4)
	if _, err := ctx.Conn.Read(buf); err != nil {
		return err
	}

	// Validate request type: 0x01 means CONNECT
	// 校验请求类型：0x01 表示 CONNECT
	if buf[0] != 0x05 || buf[1] != 0x01 {
		return errors.New("unsupported request") // 不支持的请求类型
	}

	// Logging version and command
	// 日志记录版本和命令字段
	ctx.Debugf("Parsing SOCKS5 request: VER=0x%02X CMD=0x%02X", buf[0], buf[1])

	switch buf[3] {
	case 0x01: // IPv4 address
		ipv4 := make([]byte, 4)
		if _, err := ctx.Conn.Read(ipv4); err != nil {
			return err
		}
		ctx.DstHost = net.IP(ipv4).String()

	case 0x03: // Domain name
		aLen := make([]byte, 1)
		if _, err := ctx.Conn.Read(aLen); err != nil {
			return err
		}

		domain := make([]byte, aLen[0])
		if _, err := ctx.Conn.Read(domain); err != nil {
			return err
		}
		ctx.DstHost = string(domain)

	case 0x04: // IPv6 address
		ipv6 := make([]byte, 16)
		if _, err := ctx.Conn.Read(ipv6); err != nil {
			return err
		}
		ctx.DstHost = net.IP(ipv6).String()

	default:
		return errors.New("unsupported address type") // 不支持的地址类型
	}

	// Read destination port (2 bytes)
	// 读取目标端口（2 字节）
	port := make([]byte, 2)
	if _, err := ctx.Conn.Read(port); err != nil {
		return err
	}
	ctx.DstPort = strconv.Itoa(int(port[0])<<8 | int(port[1]))

	// Send success response
	// 发送连接成功的响应
	_, _ = ctx.Conn.Write([]byte{
		0x05, 0x00, 0x00, 0x01, // Version, success, reserved, address type (IPv4)
		0x00, 0x00, 0x00, 0x00, // BIND address (ignored here)
		0x00, 0x00, // BIND port
	})

	return nil
}
