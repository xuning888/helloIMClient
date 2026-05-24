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
	"github.com/xuning888/helloIMClient/internal/http"
	pb "github.com/xuning888/helloIMClient/internal/proto"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/protocol/send"
)

func TestNewClient(t *testing.T) {
	logger.InitLogger()
	conf.UserId = 1
	conf.UserName = "user1"
	http.Init("http://127.0.0.1:8087", time.Second*5)
	client, err := NewImClient(getSeq, testDispatch)
	if err != nil {
		t.Fatal(err)
	}
	if err = client.Start(); err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	t.Logf("ips: %v", client.Info.IpList)
	var n = 100
	for i := 0; i < n; i++ {
		request := buildMsg(i, 1)
		now := time.Now()
		response, err2 := client.WriteMessage(context.Background(), request)
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
	client, err := NewImClient(getSeq, testDispatch)
	if err != nil {
		t.Fatal(err)
	}
	if err = client.Start(); err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	request := buildMsg(i, int64(i))
	now := time.Now()
	response, err2 := client.WriteMessage(context.Background(), request)
	cost := time.Since(now).Milliseconds()
	assert.Nil(t, err2)
	sendResponse, ok := response.(*send.SendAck)
	assert.True(t, ok)
	t.Logf("resp: %v, cost: %d ms", sendResponse, cost)
}

func buildMsg(i int, from int64) *send.SendMsg {
	return send.NewSendMsg(from, 2, 1, &pb.Payload{
		PayloadType: pb.PayloadType_TEXT,
		Content: &pb.Payload_Text{
			Text: &pb.TextPayload{Content: fmt.Sprintf("hello world %d", i)},
		},
	}, 0, 0)
}

func testDispatch(result *Result) {
	fmt.Printf("testDispatch: %v\n", result)
}

var seq atomic.Int32 = atomic.Int32{}

func getSeq() int32 {
	load := seq.Load()
	seq.Add(1)
	return load
}
