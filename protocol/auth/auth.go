package auth

import (
	"fmt"

	"github.com/xuning888/helloIMClient/internal/proto"
	"github.com/xuning888/helloIMClient/protocol"
	"google.golang.org/protobuf/proto"
)

var _ protocol.Request = &Request{}
var _ protocol.Response = &Response{}

type Request struct {
	*helloim_proto.AuthRequest
}

func (r *Request) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_AUTH)
}

type Response struct {
	*helloim_proto.AuthResponse
	f *protocol.Frame
}

func (r *Response) MsgSeq() int32 {
	return r.f.Header.Seq
}

func (r *Response) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_AUTH)
}

// ServerSeq Auth 信令不需要在通道中传递不需要serverSeq
func (r *Response) ServerSeq() int64 {
	return 0
}

// MsgId Auth 信令不需要在通道中传输也不需要存储, 所以msgId也传个0
func (r *Response) MsgId() int64 {
	return 0
}

func NewRequest(userId int64, userType int32, token string) *Request {
	req := &Request{
		AuthRequest: &helloim_proto.AuthRequest{
			Uid:      fmt.Sprintf("%d", userId),
			UserType: userType,
			Token:    token,
		},
	}
	return req
}

func RequestDecode(frame *protocol.Frame) (protocol.Request, error) {
	body := frame.Body
	pbReq := helloim_proto.AuthRequest{}
	err := proto.Unmarshal(body, &pbReq)
	if err != nil {
		return nil, err
	}
	req := &Request{
		&pbReq,
	}
	return req, nil
}

func ResponseDecode(frame *protocol.Frame) (protocol.Response, error) {
	body := frame.Body
	pbResp := helloim_proto.AuthResponse{}
	err := proto.Unmarshal(body, &pbResp)
	if err != nil {
		return nil, err
	}
	return &Response{
		&pbResp,
		frame,
	}, nil
}

func init() {
	protocol.RegisterDecoder(int32(helloim_proto.CmdId_CMD_ID_AUTH), &protocol.Decoder{
		RequestDecode:  RequestDecode,
		ResponseDecode: ResponseDecode,
	})
}
