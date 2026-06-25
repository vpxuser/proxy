package proxy

import (
	"net"
	"sync"
	"testing"
)

func TestServer(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()

	errCh := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		inner, err := l.Accept()
		if err != nil {
			errCh <- err
			return
		}
		defer inner.Close()

		buf := make([]byte, 1024)
		n, err := inner.Read(buf)
		if err != nil {
			errCh <- err
			return
		}
		t.Logf("read client msg: %s", buf[:n])
		errCh <- nil
	}()

	go func() {
		defer wg.Done()
		errCh <- clientConn("tcp", addr)
	}()

	wg.Wait()
	l.Close()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
}
