package heartbeat

import (
	"github.com/xuning888/helloIMClient/internal/proto"
	"github.com/xuning888/helloIMClient/protocol"
	"google.golang.org/protobuf/proto"
)

var _ protocol.Request = &Request{}
var _ protocol.Response = &Response{}

type Request struct {
	*helloim_proto.EmptyRequest
}

func (r *Request) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_HEARTBEAT)
}

type Response struct {
	*helloim_proto.EmptyResponse
	f *protocol.Frame
}

func (r *Response) MsgSeq() int32 {
	return r.f.Header.Seq
}

func (r *Response) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_HEARTBEAT)
}

func (r *Response) ServerSeq() int64 {
	return 0
}

func (r *Response) MsgId() int64 {
	return 0
}

func NewRequest() *Request {
	req := &Request{
		&helloim_proto.EmptyRequest{},
	}
	return req
}

func RequestDecode(frame *protocol.Frame) (protocol.Request, error) {
	body := frame.Body
	emptReq := helloim_proto.EmptyRequest{}
	err := proto.Unmarshal(body, &emptReq)
	if err != nil {
		return nil, err
	}
	req := &Request{
		&emptReq,
	}
	return req, nil
}

func ResponseDecode(frame *protocol.Frame) (protocol.Response, error) {
	body := frame.Body
	pbResponse := helloim_proto.EmptyResponse{}
	err := proto.Unmarshal(body, &pbResponse)
	if err != nil {
		return nil, err
	}
	return &Response{
		&pbResponse,
		frame,
	}, nil
}

func init() {
	protocol.RegisterDecoder(int32(helloim_proto.CmdId_CMD_ID_HEARTBEAT), &protocol.Decoder{
		RequestDecode:  RequestDecode,
		ResponseDecode: ResponseDecode,
	})
}
