package auth

import (
	pb "github.com/xuning888/helloIMClient/proto"
	"github.com/xuning888/helloIMClient/protocol"
	"google.golang.org/protobuf/proto"
)

var _ protocol.Request = &Request{}
var _ protocol.Response = &Response{}

type Request struct {
	*pb.AuthRequest
}

func (r *Request) CmdId() int32 {
	return int32(pb.CmdId_CMD_ID_AUTH)
}

type Response struct {
	*pb.AuthResponse
}

func (r *Response) CmdId() int32 {
	return int32(pb.CmdId_CMD_ID_AUTH)
}

// ServerSeq Auth 信令不需要在通道中传递不需要serverSeq
func (r *Response) ServerSeq() int64 {
	return 0
}

// MsgId Auth 信令不需要在通道中传输也不需要存储, 所以msgId也传个0
func (r *Response) MsgId() int64 {
	return 0
}

func RequestDecode(frame *protocol.Frame) (protocol.Request, error) {
	body := frame.Body
	pbReq := pb.AuthRequest{}
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
	pbResp := pb.AuthResponse{}
	err := proto.Unmarshal(body, &pbResp)
	if err != nil {
		return nil, err
	}
	return &Response{
		&pbResp,
	}, nil
}

func init() {
	protocol.RegisterDecoder(int32(pb.CmdId_CMD_ID_AUTH), &protocol.Decoder{
		RequestDecode:  RequestDecode,
		ResponseDecode: ResponseDecode,
	})
}
