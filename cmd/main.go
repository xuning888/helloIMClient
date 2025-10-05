package main

import (
	"context"
	"github.com/xuning888/helloIMClient/option"
	"github.com/xuning888/helloIMClient/protocol/c2csend"
	"github.com/xuning888/helloIMClient/transport"
	"log"
)

func main() {
	user := transport.ImUser{
		UserId:   2,
		UserType: 0,
		Token:    "",
	}
	imClient, err := transport.NewImClient(
		&user,
		dispatch,
		option.WithServerUrl("http://127.0.0.1:8087"),
	)
	if err != nil {
		log.Fatal(err)
	}
	toUser := transport.ImUser{
		UserId:   1,
		UserType: 0,
		Token:    "",
	}

	request := c2csend.NewRequest(user.UserId, toUser.UserId, "你好", 0,
		int32(user.UserType), int32(toUser.UserType))

	ack, err := imClient.WriteMessage(context.Background(), request)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("消息发送成功,ack: %v\n", ack)
}

func dispatch(result *transport.Result) {
}
