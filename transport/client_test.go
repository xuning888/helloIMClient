package transport

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	pb "github.com/xuning888/helloIMClient/proto"
	"github.com/xuning888/helloIMClient/protocol/c2csend"
	"sync"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	user := ImUser{
		UserId:   1,
		UserType: 0,
		Token:    "token",
	}
	client, err := NewImClient(&user, testDispatch, WithServerUrl("http://127.0.0.1:8087"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	t.Logf("ips: %v", client.Info.IpList)
	var n = 1000
	for i := 0; i < n; i++ {
		request := buildMsg(i)
		now := time.Now()
		response, err2 := client.WriteMessage(context.Background(), request)
		cost := time.Since(now).Milliseconds()
		assert.Nil(t, err2)
		c2cSendResponse, ok := response.(*c2csend.Response)
		assert.True(t, ok)
		t.Logf("resp: %v, cost: %d ms", c2cSendResponse, cost)
	}
}

func TestImClient_WriteMessage(t *testing.T) {
	var n = 100
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
	user := ImUser{
		UserId:   int64(i),
		UserType: 0,
		Token:    "token",
	}
	client, err := NewImClient(&user, testDispatch, WithServerUrl("http://127.0.0.1:8087"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	request := buildMsg(i)
	now := time.Now()
	response, err2 := client.WriteMessage(context.Background(), request)
	cost := time.Since(now).Milliseconds()
	assert.Nil(t, err2)
	c2cSendResponse, ok := response.(*c2csend.Response)
	assert.True(t, ok)
	t.Logf("resp: %v, cost: %d ms", c2cSendResponse, cost)
}

func buildMsg(i int) *c2csend.Request {
	request := &c2csend.Request{
		C2CSendRequest: &pb.C2CSendRequest{
			From:          "1",
			To:            "2",
			Content:       fmt.Sprintf("hello world %d", i),
			ContentType:   0,
			SendTimestamp: time.Now().UnixMilli(),
			FromUserType:  0,
			ToUserType:    0,
		},
	}
	return request
}

func testDispatch(result *Result) {
	fmt.Printf("testDispatch: %v\n", result)
}
