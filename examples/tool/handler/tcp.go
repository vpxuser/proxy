package handler

import (
	"github.com/vpxuser/proxy"
	yaklog "github.com/yaklang/yaklang/common/log"
)

var DumpRaw proxy.HandleRaw = func(raw []byte, reverse bool, ctx *proxy.Context) []byte {
	yaklog.Infof("%s\n%s", ctx.Preffix(reverse), raw)

	return raw
}
