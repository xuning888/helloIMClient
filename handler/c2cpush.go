package handler

import (
	"fmt"
	"github.com/xuning888/helloIMClient/app"
	"github.com/xuning888/helloIMClient/protocol/c2cpush"
)

var _ app.Handler = C2cPushHandler

func C2cPushHandler(ctx *app.ImContext) error {
	var response *c2cpush.Response = nil
	var ok bool = false
	if response, ok = ctx.DownMessage().(*c2cpush.Response); !ok {
		return nil
	}
	fmt.Printf("c2cPushMessage, response: %v\n", response)
	return nil
}
