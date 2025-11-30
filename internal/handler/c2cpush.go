package handler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/xuning888/helloIMClient/app"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/protocol/c2cpush"
)

var _ app.Handler = C2cPushHandler

func C2cPushHandler(ctx *app.ImContext) error {
	var response *c2cpush.Response = nil
	var ok = false
	if response, ok = ctx.DownMessage().(*c2cpush.Response); !ok {
		return nil
	}
	fmt.Printf("c2cPushMessage, response: %v\n", response)
	msgTo, _ := strconv.ParseInt(response.To, 10, 64)
	msgFrom, _ := strconv.ParseInt(response.From, 10, 64)
	serverSeq := response.ServerSeq()
	message := sqllite.NewMessage(1, msgFrom, response.MsgId(),
		msgFrom, msgTo,
		response.FromUserType, response.ToUserType, response.MsgSeq(), response.Content, response.ContentType,
		response.CmdId(),
		time.Now().UnixMilli(), 0, serverSeq)
	context.Background()
	err := sqllite.SaveOrUpdateMessage(context.Background(), message)
	if err != nil {
		logger.Errorf("C2cPushHandler.SaveOrUpdateMessage error: %v", err)
	}
	return nil
}
