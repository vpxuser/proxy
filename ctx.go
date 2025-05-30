package proxy

import (
	"fmt"
	"github.com/google/uuid"
	"net"
	"net/http"
	"strings"
)

type Context struct {
	Id         string
	ClientAddr string
	RemoteIp   string
	RemoteHost string
	RemotePort string
	ServName   string
	Request    *http.Request
	Response   *http.Response
	IsTLS      bool
	Protocol   string
	ThirdCtx   interface{}
	HttpProxy  *HttpProxy
}

func NewContext() *Context {
	id := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	return &Context{
		Id:       id,
		IsTLS:    false,
		Protocol: "TCP",
	}
}

func (c *Context) Preffix(args ...interface{}) (preffix string) {
	preffixSB := strings.Builder{}
	if c.Id != "" {
		preffixSB.WriteString("[" + c.Id + "]")
	}

	if len(args) > 0 {
		if args[0].(bool) {
			preffixSB.WriteString(" " + fmt.Sprintf("[%s => %s]", net.JoinHostPort(c.RemoteHost, c.RemotePort), c.ClientAddr) + " [Response]")
		} else {
			preffixSB.WriteString(" " + fmt.Sprintf("[%s => %s]", c.ClientAddr, net.JoinHostPort(c.RemoteHost, c.RemotePort)) + " [Request]")
		}
	} else {
		if c.ClientAddr != "" {
			preffixSB.WriteString(" " + "[" + c.ClientAddr + "]")
		}

		if c.RemoteHost != "" {
			preffixSB.WriteString(" " + "[" + net.JoinHostPort(c.RemoteHost, c.RemotePort) + "]")
		}
	}

	if c.Protocol != "" {
		preffixSB.WriteString(" " + "[" + c.Protocol + "]")
	}

	return preffixSB.String()
}
