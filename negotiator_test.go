package proxy

import (
	"net"
	"net/http"
	"sync"
	"testing"
)

type testCase func(*testing.T, net.Conn)

func serverWitchTestCase(wg *sync.WaitGroup, network, addr string, t *testing.T, testCase testCase) {
	defer wg.Done()
	listener, err := net.Listen(network, addr)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	inner, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer inner.Close()
	testCase(t, inner)
}

func printAddrResult(t *testing.T, conn net.Conn) {
	ctx := NewContext(ctxLogger, "test", nil)
	ctx.Conn = NewConn(conn)
	t.Logf("target addr: %s:%s", ctx.DstHost, ctx.DstPort)
}

func clientWithTestCase(wg *sync.WaitGroup, network, addr string, t *testing.T, testCase testCase) {
	defer wg.Done()
	conn, err := net.Dial(network, addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	testCase(t, conn)
}

func testConnectReq(t *testing.T, conn net.Conn) {
	req, err := http.NewRequest(http.MethodConnect, "https://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = req.Write(conn)
	if err != nil {
		t.Fatal(err)
	}
}

func testCommonReq(t *testing.T, conn net.Conn) {
	req, err := http.NewRequest(http.MethodGet, "http://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = req.WriteProxy(conn)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHttpHandshake(t *testing.T) {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	t.Log("test CONNECT method")
	go serverWitchTestCase(wg, "tcp", "127.0.0.1:52234", t, printAddrResult)
	go clientWithTestCase(wg, "tcp", "127.0.0.1:52234", t, testConnectReq)
	wg.Wait()

	wg.Add(2)
	t.Log("test GET method")
	go serverWitchTestCase(wg, "tcp", "127.0.0.1:52234", t, printAddrResult)
	go clientWithTestCase(wg, "tcp", "127.0.0.1:52234", t, testCommonReq)
	wg.Wait()
}
