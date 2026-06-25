package proxy

import (
	"net"
	"net/http"
	"sync"
	"testing"
)

type testCase func(*testing.T, net.Conn)

func TestHttpHandshake(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()

	t.Run("CONNECT", func(t *testing.T) {
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

			ctx := NewContext(ctxLogger, "test", nil)
			ctx.Conn = NewConn(inner)
			t.Logf("target addr: %s:%s", ctx.DstHost, ctx.DstPort)
			errCh <- nil
		}()

		go func() {
			defer wg.Done()
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				errCh <- err
				return
			}
			defer conn.Close()

			req, err := http.NewRequest(http.MethodConnect, "https://www.google.com", nil)
			if err != nil {
				errCh <- err
				return
			}
			errCh <- req.Write(conn)
		}()

		wg.Wait()
		close(errCh)
		for err := range errCh {
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("GET", func(t *testing.T) {
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

			ctx := NewContext(ctxLogger, "test", nil)
			ctx.Conn = NewConn(inner)
			t.Logf("target addr: %s:%s", ctx.DstHost, ctx.DstPort)
			errCh <- nil
		}()

		go func() {
			defer wg.Done()
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				errCh <- err
				return
			}
			defer conn.Close()

			req, err := http.NewRequest(http.MethodGet, "http://www.google.com", nil)
			if err != nil {
				errCh <- err
				return
			}
			errCh <- req.WriteProxy(conn)
		}()

		wg.Wait()
		close(errCh)
		for err := range errCh {
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	l.Close()
}
