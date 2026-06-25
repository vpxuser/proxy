package proxy

import (
	"net"
)

// clientConn connects to addr and writes "hello world", returning any error.
func clientConn(network, addr string) error {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte("hello world"))
	return err
}
