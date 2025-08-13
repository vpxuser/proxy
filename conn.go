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
	total := 0

	if c.buf.Len() > 0 {
		n, _ := c.buf.Read(p)
		total += n
		if total == len(p) {
			return total, nil
		}
	}

	if total < len(p) {
		n, err := c.Conn.Read(p[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
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
