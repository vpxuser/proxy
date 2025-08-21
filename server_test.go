package proxy

import (
	"sync"
	"testing"
)

func testDispatch(t *testing.T) DispatchFn {
	return func(ctx *Context) error {
		buf := make([]byte, 1024)
		n, err := ctx.Conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("read client msg: %s", buf[:n])
		return nil
	}
}

func newTestCfg(t *testing.T) *Config { return &Config{Dispatcher: testDispatch(t)} }

func server(wg *sync.WaitGroup, addr string, t *testing.T) {
	defer wg.Done()
	if err := ListenAndServe(addr, newTestCfg(t)); err != nil {
		t.Fatal(err)
	}
}

func TestServer(t *testing.T) {
	network := "tcp"
	addr := "127.0.0.1:23999"
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go server(wg, addr, t)
	go client(wg, network, addr, t)
	wg.Wait()
}
