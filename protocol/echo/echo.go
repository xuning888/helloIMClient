package echo

import (
	"github.com/xuning888/helloIMClient/internal/proto"
	"github.com/xuning888/helloIMClient/protocol"
	"google.golang.org/protobuf/proto"
)

var _ protocol.Request = &Request{}
var _ protocol.Response = &Response{}

type Request struct {
	*helloim_proto.EchoRequest
}

func (r *Request) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_ECHO)
}

type Response struct {
	*helloim_proto.EchoResponse
	f *protocol.Frame
}

func (r *Response) MsgSeq() int32 {
	return r.f.Header.Seq
}

func (r *Response) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_ECHO)
}

func (r *Response) ServerSeq() int64 {
	return 0
}

func (r *Response) MsgId() int64 {
	return 0
}

// NewRequest 构造一个空包
func NewRequest() *Request {
	req := &Request{
		EchoRequest: &helloim_proto.EchoRequest{
			Msg: "",
		},
	}
	return req
}

func RequestDecode(frame *protocol.Frame) (protocol.Request, error) {
	body := frame.Body
	pbRequest := helloim_proto.EchoRequest{}
	err := proto.Unmarshal(body, &pbRequest)
	if err != nil {
		return nil, err
	}
	req := &Request{
		&pbRequest,
	}
	return req, nil
}

func ResponseDecode(frame *protocol.Frame) (protocol.Response, error) {
	body := frame.Body
	pbResponse := helloim_proto.EchoResponse{}
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
	protocol.RegisterDecoder(int32(helloim_proto.CmdId_CMD_ID_ECHO), &protocol.Decoder{
		RequestDecode:  RequestDecode,
		ResponseDecode: ResponseDecode,
	})
}
