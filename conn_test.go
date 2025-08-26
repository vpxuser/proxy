package proxy

import (
	"net"
	"sync"
	"testing"
)

func connServer(wg *sync.WaitGroup, network, addr string, t *testing.T) {
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

	conn := NewConn(inner)

	testPeek(conn, 6, t)
	testTee(conn, 6, t)
	testPeek(conn, 11, t)
	testTee(conn, 11, t)
	testRead(conn, 6, t)
	testPeek(conn, 5, t)
	testTee(conn, 5, t)
	testRead(conn, 5, t)
}

func testTee(conn *Conn, n int, t *testing.T) {
	buf := make([]byte, n)
	_, err := conn.GetTeeReader().Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("tee buffer length: %d , data: %s", len(buf), buf)
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

func client(wg *sync.WaitGroup, network, addr string, t *testing.T) {
	defer wg.Done()
	conn, err := net.Dial(network, addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if _, err = conn.Write([]byte("hello world")); err != nil {
		t.Fatal(err)
	}
}

func TestConn(t *testing.T) {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	network := "tcp"
	addr := "127.0.0.1:23999"
	go connServer(wg, network, addr, t)
	go client(wg, network, addr, t)
	wg.Wait()
}
