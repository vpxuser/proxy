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
		n += m
		if err != nil {
			return n, err
		}
	}
	return n, err
}

func (c *Conn) Peek(n int) ([]byte, error) {
	buf := make([]byte, n)
	copy(buf, c.buf.Bytes())

	if bufLen := c.buf.Len(); n > bufLen {
		_, err := io.TeeReader(c.Conn, c.buf).Read(buf[bufLen:])
		if err != nil {
			return buf, err
		}
	}

	return buf, nil
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
	copy(p, r.buf.Bytes())
	bufLen := r.buf.Len()
	if len(p) > bufLen {
		n, err := io.TeeReader(r.Reader, r.buf).Read(p[bufLen:])
		return n + bufLen, err
	}
	return bufLen, nil
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
