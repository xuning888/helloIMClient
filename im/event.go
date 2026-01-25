package im

import (
	"github.com/panjf2000/gnet/v2"
)

var _ gnet.EventHandler = &eventHandler{}

type OnResult func(conn gnet.Conn)

type eventHandler struct {
	gnet.BuiltinEventEngine
	onResult OnResult
}

func (cc *eventHandler) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	cc.onResult(c)
	return nil, gnet.None
}

func (cc *eventHandler) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	return gnet.None
}

func (cc *eventHandler) OnTraffic(c gnet.Conn) (action gnet.Action) {
	if hc := getHiConn(c); hc != nil {
		return hc.onMessage()
	}
	return gnet.Close
}

func getHiConn(c gnet.Conn) *HiConn {
	ctx := c.Context()
	if ctx == nil {
		return nil
	}
	if hc, ok := ctx.(*HiConn); ok {
		return hc
	}
	return nil
}

func newEventHandler(connResult OnResult) *eventHandler {
	return &eventHandler{
		onResult: connResult,
	}
}
