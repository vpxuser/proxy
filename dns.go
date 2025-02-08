package proxy

import (
	"crypto/tls"
	"net"
	"sync"
)

var (
	ServName = new(sync.Map)
	NsLookup = new(sync.Map)
)

func fetchDNS(host string, port string) string {
	tlsRemote, err := tls.Dial("tcp", host+":"+port, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return ""
	}
	defer tlsRemote.Close()

	ip, _, _ := net.SplitHostPort(tlsRemote.RemoteAddr().String())
	NsLookup.Store(ip, host)

	if err = tlsRemote.Handshake(); err != nil {
		return ""
	}

	sni := tlsRemote.ConnectionState().ServerName

	ServName.Store(host, sni)

	return sni
}
