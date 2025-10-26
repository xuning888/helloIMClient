package http

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClient_Users(t *testing.T) {
	cli := NewClient("http://127.0.0.1:8087", time.Second*3)
	users, err := cli.Users(context.Background())
	assert.Nil(t, err)
	t.Log(users)
}
