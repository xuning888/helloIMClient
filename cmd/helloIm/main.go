package main

import (
	"flag"
	"log"
	"time"

	"github.com/xuning888/helloIMClient/app"
	"github.com/xuning888/helloIMClient/conf"
	"github.com/xuning888/helloIMClient/internal/dal"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/handler"
	"github.com/xuning888/helloIMClient/internal/http"
	pb "github.com/xuning888/helloIMClient/internal/proto"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

var (
	userId    int64
	username  string
	serverUrl string
)

func init() {
	flag.Int64Var(&userId, "userId", 0, "-userId userId")
	flag.StringVar(&username, "username", "", "-username username")
	flag.StringVar(&serverUrl, "serverUrl", "http://127.0.0.1:8087", "-serverUrl http://127.0.0.1:8087")
}

func main() {
	flag.Parse()
	if userId == 0 {
		log.Fatal("请输userId")
	}
	if username == "" {
		log.Fatal("请输入用户名")
	}
	if serverUrl == "" {
		log.Fatal("请输入服务器地址")
	}
	if err := dal.Init(); err != nil {
		log.Fatal(err)
	}
	if err := logger.InitLogger(); err != nil {
		log.Fatal(err)
	}
	// 初始化HTTP客户端
	http.Init(serverUrl, time.Second*10)
	u := sqllite.ImUser{
		UserID:   userId,
		UserName: username,
		UserType: 0,
	}
	conf.UserId = userId
	imApp, err := app.NewApp(&u)
	if err != nil {
		log.Fatal(err)
	}
	// 注册指令
	register(imApp)
	// start App
	if err = imApp.Start(); err != nil {
		log.Fatal(err)
	}
}

func register(imApp *app.ImApp) {
	imApp.Register(int32(pb.CmdId_CMD_ID_C2CPUSH), handler.C2cPushHandler)
}
