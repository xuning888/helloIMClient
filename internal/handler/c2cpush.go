package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/xuning888/helloIMClient/app"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/protocol/c2cpush"
	"github.com/xuning888/helloIMClient/tui"
)

var _ app.Handler = C2cPushHandler

func C2cPushHandler(ctx *app.ImContext) error {
	var response *c2cpush.Response = nil
	var ok = false
	if response, ok = ctx.DownMessage().(*c2cpush.Response); !ok {
		return nil
	}
	var msgTo int64
	var err error
	if msgTo, err = strconv.ParseInt(response.To, 10, 64); err != nil {
		return err
	}
	var msgFrom int64
	if msgFrom, err = strconv.ParseInt(response.From, 10, 64); err != nil {
		return err
	}
	logger.Infof("C2cPushHandler 接收到单聊下行消息, msgId: %v", response.MsgId())
	// 收到消息立刻回复ACK
	request := c2cpush.NewRequest(msgFrom, msgTo, response.MsgId())
	if err2 := ctx.AsyncSendMessageWithSeq(context.Background(), response.MsgSeq(), request); err2 != nil {
		logger.Errorf("C2cPushHandler sendAck error: %v", err2)
	}
	serverSeq := response.ServerSeq()
	message := sqllite.NewMessage(1, msgFrom, response.MsgId(),
		msgFrom, msgTo,
		response.FromUserType, response.ToUserType, response.MsgSeq(), response.Content, response.ContentType,
		response.CmdId(),
		time.Now().UnixMilli(), 0, serverSeq)
	if err = sqllite.SaveOrUpdateMessage(context.Background(), message); err != nil {
		logger.Errorf("C2cPushHandler.SaveOrUpdateMessage error: %v", err)
		return err
	}
	//service.UpdateChatVersion(msgFrom)
	// 更新tui
	ctx.SendTuiCmd(
		tui.FetchUpdateMessage(msgFrom), // 发送更新消息的cmd
		tui.FetchUpdatedChatListCmd(),   // 发送更新会话列表的cmd
	)
	return nil
}
