package proxy

import (
	"bytes"
	"crypto/tls"
	"io"
	"net"
)

type Conn struct {
	net.Conn
	PeekRd *PeekReader
}

func (c *Conn) Read(p []byte) (int, error) {
	multiRd := io.MultiReader(c.PeekRd.buf, c.Conn)
	return multiRd.Read(p)
}

func (c *Conn) Peek(n int) ([]byte, error) {
	buf := make([]byte, n)
	total := copy(buf, c.PeekRd.buf.Bytes())
	for total < n {
		readN, err := c.Conn.Read(buf[total:])
		if readN > 0 {
			c.PeekRd.buf.Write(buf[total : total+readN])
			total += readN
		}
		if err != nil || readN == 0 {
			if err == io.EOF && total > 0 {
				break
			}
			return buf[:total], err
		}
	}
	return buf[:total], nil
}

func (c *Conn) IsTLS() bool {
	_, ok := c.Conn.(*tls.Conn)
	return ok
}

func NewConn(inner net.Conn) *Conn {
	return &Conn{
		Conn: inner,
		PeekRd: &PeekReader{
			rd:  inner,
			buf: bytes.NewBuffer(nil),
			//bufRd: bufio.NewReader(inner),
		},
	}
}

type PeekReader struct {
	rd  io.Reader
	buf *bytes.Buffer
	//bufRd *bufio.Reader
}

func (r *PeekReader) Read(p []byte) (int, error) {
	//bytesRd := bytes.NewReader(r.buf.Bytes())
	//teeRd := io.TeeReader(r.rd, r.buf)
	//multiRd := io.MultiReader(bytesRd, teeRd)
	//return multiRd.Read(p)
	total := copy(p, r.buf.Bytes())
	if total < len(p) {
		readN, err := r.rd.Read(p[total:])
		if readN > 0 {
			r.buf.Write(p[total : total+readN])
			total += readN
		}
		return total, err
	}
	return total, nil
}
