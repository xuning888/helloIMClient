package net

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xuning888/helloIMClient/frame"
	pb "github.com/xuning888/helloIMClient/proto"
	"google.golang.org/protobuf/proto"
)

var user = &GateUser{
	Uid:      "1",
	UserType: 0,
}

var handlerMap = make(map[int]ReplyHandler)

func register(handler ReplyHandler) {
	cmdId := handler.CmdId()
	handlerMap[int(cmdId)] = handler
}

type ReplyHandler interface {
	CmdId() pb.CmdId
	Handle(reply *frame.Frame) error
}

func dispatch(reply *frame.Frame) {
	if reply == nil {
		return
	}
	cmdId := reply.Header.CmdId
	handler := handlerMap[int(cmdId)]
	if handler != nil {
		if err := handler.Handle(reply); err != nil {
			log.Printf("handle reply error: %v", err)
		}
	}
}

// EchoHandler 下行Echo处理器
type EchoHandler struct {
}

func (e *EchoHandler) CmdId() pb.CmdId {
	return pb.CmdId_CMD_ID_ECHO
}

func (e *EchoHandler) Handle(reply *frame.Frame) error {
	body := reply.Body
	response := &pb.EchoResponse{}
	if err := proto.Unmarshal(body, response); err != nil {
		return err
	}
	log.Printf("receive: %s\n", response)
	return nil
}

func init() {
	register(&EchoHandler{})
}

func Test_echo(t *testing.T) {
	// 新建cli, 并注册业务处理器
	client := NewImClient(dispatch)
	defer client.Close()

	// 建立长连接
	err2 := client.Connect("127.0.0.1:9300", user)
	if err2 != nil {
		t.Fatal(err2)
	}
	// 构造echo消息
	request, err := frame.MakeEchoRequest("hello world", GetSeq())
	if err != nil {
		t.Fatal(err)
	}

	stop := make(chan struct{})
	err = client.SendFrameWithCallback(request, func(request, response *frame.Frame, err error) {
		stop <- struct{}{}
	})
	if err != nil {
		t.Fatal(err)
	}
	<-stop
}

// 修改你的测试函数
func Test_c2csendACK(t *testing.T) {
	client := NewImClient(dispatch)
	defer client.Close()

	err2 := client.Connect("127.0.0.1:9300", user)
	if err2 != nil {
		t.Fatal(err2)
	}
	var wg sync.WaitGroup
	var n = 1000
	for i := 0; i < n; i++ {
		request, err3 := frame.MakeC2cSendMessage(user.Uid, "2", fmt.Sprintf("hello world_%d", i), 0, GetSeq())
		if err3 != nil {
			t.Fatal(err3)
		}
		wg.Add(1)
		go func(req *frame.Frame) {
			defer wg.Done()
			send(t, client, req)
		}(request)
	}
	wg.Wait()
}

func send(t *testing.T, client *ImClient, request *frame.Frame) {
	stop := make(chan struct{})
	log.Printf("发送上行消息: %v\n", request.Key())
	now := time.Now()
	err := client.SendFrameWithCallback(request, func(request, response *frame.Frame, err error) {
		defer func() { stop <- struct{}{} }()
		if err != nil {
			log.Printf("发送单聊上行消息失败, error: %v", err)
			return
		}
		ms := time.Since(now).Milliseconds()
		c2cSendResponse := &pb.C2CSendResponse{}
		if err3 := proto.Unmarshal(response.Body, c2cSendResponse); err3 != nil {
			return
		}
		assert.True(t, request.Key() == response.Key())

		log.Printf("上行消息ACK, request: %v, response: %v, cost: %vms\n", request.Key(), response.Key(), ms)
	})
	if err != nil {
		t.Fatal(err)
	}
	<-stop
}
