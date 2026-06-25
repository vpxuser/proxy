package proxy

import (
	"net"
	"sync"
	"testing"
)

func TestConn(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()

	errCh := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			errCh <- err
			return
		}
		defer listener.Close()

		inner, err := listener.Accept()
		if err != nil {
			errCh <- err
			return
		}

		conn := NewConn(inner)
		testPeek(conn, 6, t)
		testPeek(conn, 11, t)
		testRead(conn, 6, t)
		testPeek(conn, 5, t)
		testRead(conn, 5, t)
		errCh <- nil
	}()

	go func() {
		defer wg.Done()
		errCh <- clientConn("tcp", addr)
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testRead(conn *Conn, n int, t *testing.T) {
	buf := make([]byte, n)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("read buffer length: %d , data: %s", len(buf[:n]), buf[:n])
}

func testPeek(conn *Conn, n int, t *testing.T) {
	buf, err := conn.Peek(n)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("peek buffer length: %d , data: %s", len(buf), buf)
}
