package app

import (
	"context"

	"github.com/xuning888/helloIMClient/internal/model"
	"github.com/xuning888/helloIMClient/protocol"
	"github.com/xuning888/helloIMClient/svc"
	"github.com/xuning888/helloIMClient/transport"
)

type ImContext struct {
	CmdId     int32               // 指令号
	imCli     *transport.ImClient // IM客户端
	response  protocol.Response   // 接收到的消息
	commonSvc *svc.CommonSvc      // 操作数据
}

// DownMessage 获取下行消息
func (c *ImContext) DownMessage() protocol.Response {
	return c.response
}

// SendMessage 发送上行消息, 同步等待ACK
func (c *ImContext) SendMessage(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	return c.imCli.WriteMessage(ctx, request)
}

func (c *ImContext) AppendMessage(msg *model.ChatMessage) {
	chat := c.commonSvc.GetChat(msg.ChatId, msg.ChatType)
	if chat != nil {
		chat.Msgs.AppendMsg(msg)
	}
}

// reset 重制context
func (c *ImContext) reset() {
	c.CmdId = 0
	c.imCli = nil
	c.response = nil
}
