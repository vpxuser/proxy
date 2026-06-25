package proxy

import (
	"sync"
	"testing"
)

func TestListener(t *testing.T) {
	cfg := &Config{Dispatcher: DispatchFn(func(ctx *Context) error {
		buf := make([]byte, 1024)
		n, err := ctx.Conn.Read(buf)
		if err != nil {
			return err
		}
		t.Logf("read client msg: %s", buf[:n])
		return nil
	})}

	l, err := Listen("tcp", "127.0.0.1:0", cfg)
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
		conn := NewConn(inner)
		buf := make([]byte, 1024)
		_, err = conn.Read(buf)
		if err != nil {
			errCh <- err
			return
		}
		t.Logf("read client msg: %s", buf)
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
