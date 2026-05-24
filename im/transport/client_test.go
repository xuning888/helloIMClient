package transport

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xuning888/helloIMClient/conf"
	"github.com/xuning888/helloIMClient/im/http"
	"github.com/xuning888/helloIMClient/im/proto"
	"github.com/xuning888/helloIMClient/im/protocol"
	"github.com/xuning888/helloIMClient/im/protocol/send"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

func TestNewClient(t *testing.T) {
	logger.InitLogger()
	conf.UserId = 1
	conf.UserName = "user1"
	http.Init("http://127.0.0.1:8087", time.Second*5)

	client := NewClient(testDispatch, &testAddrProvider{}, getSeq)
	if err := client.Connect(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	var n = 5000
	for i := 0; i < n; i++ {
		request := buildMsg(i, 1)
		now := time.Now()
		response, err2 := client.Send(context.Background(), request)
		cost := time.Since(now).Milliseconds()
		assert.Nil(t, err2)
		sendResponse, ok := response.(*send.SendAck)
		assert.True(t, ok)
		t.Logf("resp: %v, cost: %d ms", sendResponse, cost)
		time.Sleep(time.Millisecond * 10)
	}
}

func TestImClient_WriteMessage(t *testing.T) {
	var n = 1
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			writeMessage(i, t)
		}()
	}
	wg.Wait()
}

func writeMessage(i int, t *testing.T) {
	logger.InitLogger()
	http.Init("http://127.0.0.1:8087", time.Second*5)
	client := NewClient(testDispatch, &testAddrProvider{}, getSeq)
	if err := client.Connect(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	request := buildMsg(i, int64(i))
	now := time.Now()
	response, err2 := client.Send(context.Background(), request)
	cost := time.Since(now).Milliseconds()
	assert.Nil(t, err2)
	sendResponse, ok := response.(*send.SendAck)
	assert.True(t, ok)
	t.Logf("resp: %v, cost: %d ms", sendResponse, cost)
}

func buildMsg(i int, from int64) *send.SendMsg {
	return send.NewSendMsg(from, 2, 1, &helloim_proto.Payload{
		PayloadType: helloim_proto.PayloadType_TEXT,
		Content: &helloim_proto.Payload_Text{
			Text: &helloim_proto.TextPayload{Content: fmt.Sprintf("hello world %d", i)},
		},
	}, 0, 0)
}

func testDispatch(msg protocol.Message) {
	fmt.Printf("testDispatch: %v\n", msg)
}

type testAddrProvider struct{}

func (p *testAddrProvider) GetAddr(ctx context.Context) ([]string, error) {
	return []string{"127.0.0.1:9299"}, nil
}

var seq atomic.Int32 = atomic.Int32{}

func getSeq() int32 {
	load := seq.Load()
	seq.Add(1)
	return load
}
