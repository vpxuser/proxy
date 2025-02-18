package handler

import (
	"github.com/gobwas/ws"
	"github.com/vpxuser/proxy"
	yaklog "github.com/yaklang/yaklang/common/log"
)

var opType = map[ws.OpCode]string{
	ws.OpContinuation: "Continuation",
	ws.OpText:         "Text",
	ws.OpBinary:       "Binary",
	ws.OpClose:        "Close",
	ws.OpPing:         "Ping",
	ws.OpPong:         "Pong",
}

var DumpWebSocket proxy.HandleWebSocket = func(frame ws.Frame, reverse bool, ctx *proxy.Context) ws.Frame {
	payload := frame.Payload
	if frame.Header.Masked {
		payload = ws.UnmaskFrame(frame).Payload
	}

	yaklog.Infof("%s [%s]\n%s", ctx.Preffix(reverse), opType[frame.Header.OpCode], payload)

	return frame
}
