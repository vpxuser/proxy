package proxy

import (
	"bytes"
	"io"
	"net"
)

type Conn struct {
	net.Conn
	teeReader io.Reader
	buf       *bytes.Buffer
}

func (c *Conn) Read(p []byte) (int, error) {
	n, err := c.buf.Read(p)
	if err == io.EOF {
		m, err := c.Conn.Read(p[n:])
		n += m
		if err != nil {
			return n, err
		}
	}
	return n, err
}

func (c *Conn) TeeReader() io.Reader { return c.teeReader }

func NewConn(inner net.Conn) *Conn {
	c := &Conn{
		Conn: inner,
		buf:  bytes.NewBuffer(nil),
	}
	c.teeReader = io.TeeReader(inner, c.buf)
	return c
}
