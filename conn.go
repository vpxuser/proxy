package proxy

import (
	"bufio"
	"net"
)

type Conn struct {
	net.Conn
	r *bufio.Reader
}

func (c *Conn) Read(p []byte) (int, error)            { return c.r.Read(p) }
func (c *Conn) Peek(n int) ([]byte, error)            { return c.r.Peek(n) }
func (c *Conn) ReadString(delim byte) (string, error) { return c.r.ReadString(delim) }

func NewConn(inner net.Conn) *Conn { return &Conn{Conn: inner, r: bufio.NewReader(inner)} }
