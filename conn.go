package proxy

import (
	"bufio"
	"bytes"
	"io"
	"net"
)

type Conn struct {
	net.Conn
	reader    *bufio.Reader
	buf       *bytes.Buffer
	teeReader io.Reader
}

func (c *Conn) Read(p []byte) (int, error) {
	n, err := c.buf.Read(p)
	if err == io.EOF {
		m, err := c.reader.Read(p[n:])
		n += m
		if err != nil {
			return n, err
		}
	}
	return n, err
}

func (c *Conn) Reader() *bufio.Reader { return c.reader }
func (c *Conn) TeeReader() io.Reader  { return c.teeReader }

func NewConn(inner net.Conn) *Conn {
	c := &Conn{
		Conn:   inner,
		reader: bufio.NewReader(inner),
		buf:    bytes.NewBuffer(nil),
	}
	c.teeReader = io.TeeReader(c.reader, c.buf)
	return c
}
