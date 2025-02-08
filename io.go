package proxy

import (
	"bytes"
	"io"
	"net"
	"sync"
)

type Conn struct {
	sync.Mutex
	Buffer *bytes.Buffer
	io.Reader
	net.Conn
}

func (c *Conn) Read(p []byte) (n int, err error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if c.Buffer == nil || c.Buffer.Len() == 0 {
		return c.Conn.Read(p)
	}

	n, err = c.Buffer.Read(p)
	if err == io.EOF {
		c.Buffer.Reset()
		var size int
		size, err = c.Conn.Read(p[n:])
		n += size
	}
	return n, err
}

func NewConn(conn net.Conn) *Conn {
	c := &Conn{
		Conn:   conn,
		Buffer: bytes.NewBuffer(make([]byte, 0, 1024)),
	}
	c.Reader = io.TeeReader(conn, c.Buffer)
	return c
}
