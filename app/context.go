package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xuning888/helloIMClient/protocol"
	"github.com/xuning888/helloIMClient/transport"
)

type ImContext struct {
	program  *tea.Program
	CmdId    int32               // 指令号
	imCli    *transport.ImClient // IM客户端
	response protocol.Response   // 接收到的消息
}

// DownMessage 获取下行消息
func (c *ImContext) DownMessage() protocol.Response {
	return c.response
}

// SendMessage 发送上行消息, 同步等待ACK
func (c *ImContext) SendMessage(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	return c.imCli.WriteMessage(ctx, request)
}

// AsyncSendMessageWithSeq 异步发送, 不关注ack
func (c *ImContext) AsyncSendMessageWithSeq(ctx context.Context, seq int32, request protocol.Request) error {
	return c.imCli.WriteMessageWithSeq(ctx, seq, request)
}

// reset 重制context
func (c *ImContext) reset() {
	c.CmdId = 0
	c.imCli = nil
	c.response = nil
	c.program = nil
}

func (c *ImContext) SendTuiCmd(cmds ...tea.Cmd) {
	for _, cmd := range cmds {
		msg := cmd()
		c.program.Send(msg)
	}
}
