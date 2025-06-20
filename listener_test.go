package proxy

import (
	"sync"
	"testing"
)

func listenServer(network, addr string, t *testing.T) {
	l, err := Listen(network, addr, newTestCfg(t))
	if err != nil {
		t.Error(err)
	}
	defer l.Close()

	if err = l.Serve(); err != nil {
		t.Error(err)
	}
}

func TestListener(t *testing.T) {
	wg := new(sync.WaitGroup)
	network := "tcp"
	addr := "127.0.0.1:23999"
	wg.Add(1)
	go client(wg, network, addr, t)
	listenServer(network, addr, t)
	wg.Wait()
}
