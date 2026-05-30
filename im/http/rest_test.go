package http

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newTestClient() *WebApiClient {
	return NewRestClient("http://127.0.0.1:8087", time.Second*5)
}

func TestWebApiClient_IpList(t *testing.T) {
	cli := newTestClient()
	ips, err := cli.IpList(context.Background())
	assert.Nil(t, err)
	assert.NotEmpty(t, ips)
	t.Logf("ips: %v", ips)
}

func TestWebApiClient_GetAddr(t *testing.T) {
	cli := newTestClient()
	ips, err := cli.GetAddr(context.Background())
	assert.Nil(t, err)
	assert.NotEmpty(t, ips)
	t.Logf("addrs: %v", ips)
}

func TestWebApiClient_Users(t *testing.T) {
	cli := newTestClient()
	users, err := cli.Users(context.Background())
	assert.Nil(t, err)
	t.Logf("users count: %d", len(users))
	for _, u := range users {
		t.Logf("user: %+v", u)
	}
}

func TestWebApiClient_GetAllChat(t *testing.T) {
	cli := newTestClient()
	chats, err := cli.GetAllChat(context.Background(), 1)
	assert.Nil(t, err)
	t.Logf("chats count: %d", len(chats))
	for _, c := range chats {
		t.Logf("chat: %+v", c)
	}
}

func TestWebApiClient_LastMessage(t *testing.T) {
	cli := newTestClient()
	msg, err := cli.LastMessage(context.Background(), 1, 2, 1)
	assert.Nil(t, err)
	t.Logf("last message: %+v", msg)
}

func TestWebApiClient_PullOfflineMsg(t *testing.T) {
	cli := newTestClient()
	msgs, err := cli.PullOfflineMsg(context.Background(), 1, 2, 1, 0, 100)
	assert.Nil(t, err)
	t.Logf("offline msgs count: %d", len(msgs))
}

func TestWebApiClient_GetLatestOfflineMessages(t *testing.T) {
	cli := newTestClient()
	msgs, err := cli.GetLatestOfflineMessages(context.Background(), 1, 2, 1, 10)
	assert.Nil(t, err)
	t.Logf("latest offline msgs count: %d", len(msgs))
}
