package im

import (
	"context"

	"github.com/xuning888/helloIMClient/im/protocol"
)

// msgManager 消息管理器
type msgManager struct {
	cli *Client
}

func newMsgManager(cli *Client) *msgManager {
	return &msgManager{cli: cli}
}

// Send 通过SDK发送消息
func (mm *msgManager) Send(ctx context.Context, request protocol.Message) (protocol.Message, error) {
	return mm.cli.SendMessage(ctx, request)
}

// AddNewMsgListener 注册新消息回调，返回取消函数
func (mm *msgManager) AddNewMsgListener(cb EventCallback) func() {
	return mm.cli.events.subscribe(func(evt Event) {
		if evt.Type == EventMessageReceived {
			cb(evt)
		}
	})
}

// AddOnSendMsgListener 注册消息发送回调
func (mm *msgManager) AddOnSendMsgListener(cb EventCallback) func() {
	return mm.cli.events.subscribe(func(evt Event) {
		if evt.Type == EventMessageSent {
			cb(evt)
		}
	})
}
