package c2cpush

import (
	"testing"

	"github.com/xuning888/helloIMClient/protocol"
)

func TestResponseDecode(t *testing.T) {
	body := []byte{
		10, 1, 57, 18, 1, 49, 26, 1, 49, 40,
		164, 161, 128, 228, 176, 254, 178, 125, 48, 61, 56, 174, 181, 150, 191, 177, 51,
	}
	h := &protocol.MsgHeader{
		BodyLength: int32(len(body)),
	}
	frame := &protocol.Frame{
		Header: h,
		Body:   body,
	}
	response, _ := ResponseDecode(frame)
	r := response.(*Response)
	t.Logf("response: %v", r)
}
