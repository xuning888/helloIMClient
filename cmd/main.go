package main

import (
	"fmt"
	"log"

	"github.com/xuning888/helloIMClient/frame"
	"github.com/xuning888/helloIMClient/im"
	"github.com/xuning888/helloIMClient/net"
	pb "github.com/xuning888/helloIMClient/proto"
	"google.golang.org/protobuf/proto"
)

var Me = &im.User{}

func handleReply(reply *frame.Frame) {
	if reply.Header.CmdId == int32(pb.CmdId_CMD_ID_C2CPUSH) {
		msg := &pb.C2CPushRequest{}
		if err := proto.Unmarshal(reply.Body, msg); err != nil {
			return
		}
		fmt.Println(msg)
	}
}

func main() {
	// 新建cli, 并注册业务处理器
	client := net.NewImClient(handleReply)
	defer client.Close()

	// 建立长连接
	err2 := client.Connect("127.0.0.1:9300", &net.GateUser{
		Uid:      "2",
		UserType: 0,
	})
	if err2 != nil {
		log.Fatal(err2)
	}
	select {}
}
