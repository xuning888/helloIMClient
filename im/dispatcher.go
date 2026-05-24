package im

import (
	"context"
	"strconv"

	"github.com/xuning888/helloIMClient/im/dal/sqllite"
	"github.com/xuning888/helloIMClient/im/payload"
	pb "github.com/xuning888/helloIMClient/im/proto"
	"github.com/xuning888/helloIMClient/im/protocol"
	"github.com/xuning888/helloIMClient/im/protocol/push"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

type dispatcher struct {
	store  *Store
	events *callbackRegistry
}

func newDispatcher(store *Store, events *callbackRegistry) *dispatcher {
	return &dispatcher{
		store:  store,
		events: events,
	}
}

func (d *dispatcher) dispatch(msg protocol.Message) {
	if msg == nil {
		return
	}
	switch msg.CmdId() {
	case int32(pb.CmdId_CMD_ID_PUSH):
		d.handlePush(msg)
	default:
		logger.Infof("dispatcher: unhandled push message, cmdId: %d", msg.CmdId())
	}
}

func (d *dispatcher) handlePush(resp protocol.Message) {
	response, ok := resp.(*push.RecvMsg)
	if !ok {
		return
	}

	msgTo, err := strconv.ParseInt(response.GetChatId(), 10, 64)
	if err != nil {
		logger.Errorf("dispatcher Push: parse chatId error: %v", err)
		return
	}

	msgFrom, err := strconv.ParseInt(response.GetFrom(), 10, 64)
	if err != nil {
		logger.Errorf("dispatcher Push: parse from error: %v", err)
		return
	}

	logger.Infof("dispatcher Push: received message, msgId: %v, chatType: %d", response.MsgId(), response.GetChatType())

	content, contentType := payload.ExtractContent(response.GetPayload())
	chatType := response.GetChatType()

	message := sqllite.NewMessage(chatType, msgFrom, response.MsgId(),
		msgFrom, msgTo,
		response.GetFromUserType(), response.GetToUserType(),
		response.MsgSeq(), content, contentType,
		response.CmdId(),
		response.GetSendTimestamp(), 0, response.ServerSeq())

	if err := d.store.Messages.Save(context.Background(), message); err != nil {
		logger.Errorf("dispatcher Push: save message error: %v", err)
		d.events.fire(Event{Type: EventError, Data: err})
		return
	}

	d.store.Chats.UpdateVersion(context.Background(), msgFrom, chatType)
	d.events.fire(Event{Type: EventMessageReceived, Data: message})
}
