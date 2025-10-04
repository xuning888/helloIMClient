package protocol

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	pb "github.com/xuning888/helloIMClient/proto"
	"github.com/xuning888/helloIMClient/protocol/c2csend"
	"testing"
	"time"
)

func Test_EncodeMessageToBytes(t *testing.T) {
	var request Request = &c2csend.Request{
		C2CSendRequest: pb.C2CSendRequest{
			From:          "1",
			To:            "2",
			Content:       "hello world",
			ContentType:   0,
			SendTimestamp: time.Now().UnixMilli(),
			FromUserType:  0,
			ToUserType:    0,
		},
	}
	bytes, err := EncodeMessageToBytes(1, request)
	assert.Nil(t, err)
	fmt.Println(bytes)
}
