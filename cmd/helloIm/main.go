package main

import (
	"flag"
	"log"
	"time"

	"github.com/xuning888/helloIMClient/app"
	"github.com/xuning888/helloIMClient/conf"
	"github.com/xuning888/helloIMClient/im"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

func init() {
	flag.Int64Var(&conf.UserId, "userId", 0, "-userId userId")
	flag.StringVar(&conf.UserName, "username", "", "-username username")
	flag.StringVar(&conf.ServerUrl, "serverUrl", "http://127.0.0.1:8087", "-serverUrl http://127.0.0.1:8087")
}

func main() {
	flag.Parse()
	if conf.UserId == 0 {
		log.Fatal("请输userId")
	}
	if conf.UserName == "" {
		log.Fatal("请输入用户名")
	}
	if conf.ServerUrl == "" {
		log.Fatal("请输入服务器地址")
	}
	if err := logger.InitLogger(); err != nil {
		log.Fatal(err)
	}

	// 使用 SDK 创建客户端
	sdk, err := im.New(conf.ServerUrl,
		im.WithUID(conf.UserId),
		im.WithConnectTimeout(time.Second*10),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 创建应用并启动
	imApp := app.New(sdk)
	if err = imApp.Start(); err != nil {
		log.Fatal(err)
	}
}
