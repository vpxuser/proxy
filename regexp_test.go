package proxy

import "testing"

func TestIsDomain(t *testing.T) {
	host := "www.baidu.com"
	t.Logf("%s: %v", host, IsDomain(host))
	host = "192.168.0.1"
	t.Logf("%s: %v", host, IsDomain(host))
	host = "[::1]"
	t.Logf("%s: %v", host, IsDomain(host))
}
