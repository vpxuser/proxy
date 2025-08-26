package proxy

import (
	"bytes"
	"crypto/tls"
	"io"
	"net"
)

type Conn struct {
	net.Conn
	buf       *bytes.Buffer
	teeReader io.Reader
}

func (c *Conn) Read(p []byte) (int, error) {
	n, err := c.buf.Read(p)
	if err == io.EOF {
		m, err := c.Conn.Read(p[n:])
		return n + m, err
	}
	return n, err
}

func (c *Conn) Peek(n int) ([]byte, error) {
	buf := make([]byte, n)
	m := copy(buf, c.buf.Bytes())
	for m < n {
		readN, err := c.Conn.Read(buf[m:])
		if readN > 0 {
			c.buf.Write(buf[m : m+readN])
			m += readN
		}
		if err != nil || readN == 0 {
			if err == io.EOF && m > 0 {
				break
			}
			return buf[:m], err
		}
	}
	return buf[:m], nil
}

func (c *Conn) GetTeeReader() io.Reader { return c.teeReader }

func (c *Conn) IsTLS() bool {
	_, ok := c.Conn.(*tls.Conn)
	return ok
}

type teeReader struct {
	io.Reader
	buf *bytes.Buffer
}

func (r *teeReader) Read(p []byte) (int, error) {
	n := copy(p, r.buf.Bytes())
	if n < len(p) {
		m, err := r.Reader.Read(p[n:])
		if m > 0 {
			r.buf.Write(p[n : n+m])
			n += m
		}
		return n, err
	}
	return n, nil
}

func NewConn(inner net.Conn) *Conn {
	reader := &teeReader{
		Reader: inner,
		buf:    bytes.NewBuffer(nil),
	}
	return &Conn{
		Conn:      inner,
		buf:       reader.buf,
		teeReader: reader,
	}
}
