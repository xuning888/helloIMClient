package im

import (
	"context"
	"strconv"

	"github.com/xuning888/helloIMClient/im/dbhelper"
	"github.com/xuning888/helloIMClient/im/payload"
	pb "github.com/xuning888/helloIMClient/im/proto"
	"github.com/xuning888/helloIMClient/im/protocol"
	"github.com/xuning888/helloIMClient/im/protocol/push"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

type msgHandler struct {
	cli    *Client
	events *callbackRegistry
}

func newMsgHandler(cli *Client, events *callbackRegistry) *msgHandler {
	return &msgHandler{
		cli:    cli,
		events: events,
	}
}

func (d *msgHandler) dispatch(msg protocol.Message) {
	if msg == nil {
		return
	}
	switch msg.CmdId() {
	case int32(pb.CmdId_CMD_ID_PUSH):
		d.handlePush(msg)
	default:
		logger.Infof("msgHandler: unhandled push message, cmdId: %d", msg.CmdId())
	}
}

func (d *msgHandler) handlePush(resp protocol.Message) {
	response, ok := resp.(*push.RecvMsg)
	if !ok {
		return
	}

	msgTo, err := strconv.ParseInt(response.GetChatId(), 10, 64)
	if err != nil {
		logger.Errorf("msgHandler Push: parse chatId error: %v", err)
		return
	}

	msgFrom, err := strconv.ParseInt(response.GetFrom(), 10, 64)
	if err != nil {
		logger.Errorf("msgHandler Push: parse from error: %v", err)
		return
	}

	logger.Infof("msgHandler Push: received message, msgId: %v, chatType: %d", response.MsgId(), response.GetChatType())

	content, contentType := payload.ExtractContent(response.GetPayload())
	chatType := response.GetChatType()

	message := dbhelper.NewMessage(chatType, msgFrom, response.MsgId(),
		msgFrom, msgTo,
		response.GetFromUserType(), response.GetToUserType(),
		response.MsgSeq(), content, contentType,
		response.CmdId(),
		response.GetSendTimestamp(), 0, response.ServerSeq())

	if err := d.cli.store.Messages.Save(context.Background(), message); err != nil {
		logger.Errorf("msgHandler Push: save message error: %v", err)
		d.events.fire(Event{Type: EventError, Data: err})
		return
	}

	d.cli.store.Chats.UpdateVersion(context.Background(), msgFrom, chatType)

	// 通过内部缓存排序 + 缺口补偿后，推送已排序的消息列表
	sortedMessages := d.cli.store.Messages.ProcessIncomingMessage(message)
	d.cli.MessageManager.pushNewMsg(sortedMessages)
}
