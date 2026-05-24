package im

import (
	"context"
	"strconv"

	"github.com/xuning888/helloIMClient/im/payload"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	pb "github.com/xuning888/helloIMClient/internal/proto"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/protocol"
	"github.com/xuning888/helloIMClient/protocol/push"
	"github.com/xuning888/helloIMClient/transport"
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

func (d *dispatcher) dispatch(result *transport.Result) {
	resp := result.GetResp()
	if resp == nil {
		if err := result.GetErr(); err != nil {
			d.events.fire(Event{Type: EventError, Data: err})
		}
		return
	}
	cmdId := resp.CmdId()
	switch cmdId {
	case int32(pb.CmdId_CMD_ID_PUSH):
		d.handlePush(resp)
	default:
		logger.Infof("dispatcher: unhandled push message, cmdId: %d", cmdId)
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
