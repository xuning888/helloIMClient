package http

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_Users(t *testing.T) {
	Init("http://127.0.0.1:8087", time.Second*3)
	users, err := Users(context.Background())
	assert.Nil(t, err)
	t.Log(users)
}

func TestClient_LastMessage(t *testing.T) {
	Init("http://127.0.0.1:8087", time.Second*3)
	message, err := LastMessage(1, 2, 1)
	assert.Nil(t, err)
	t.Log(message)
}
