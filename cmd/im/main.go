package main

import (
	"github.com/xuning888/helloIMClient/app"
	"github.com/xuning888/helloIMClient/handler"
	"github.com/xuning888/helloIMClient/option"
	"github.com/xuning888/helloIMClient/pkg/logger"
	pb "github.com/xuning888/helloIMClient/proto"
	"github.com/xuning888/helloIMClient/transport"
	"log"
)

func main() {
	logger.InitLogger()
	user := &transport.ImUser{
		UserId:   2,
		UserType: 0,
		Token:    "",
	}
	imApp, err := app.NewApp(user,
		option.WithServerUrl("http://127.0.0.1:8087"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// 注册处理器
	register(imApp)
	// 启动app
	if err := imApp.Start(); err != nil {
		log.Fatal(err)
	}
	select {}
}

func register(imApp *app.ImApp) {
	imApp.Register(int32(pb.CmdId_CMD_ID_C2CPUSH), handler.C2cPushHandler)
}
