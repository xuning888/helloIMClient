package c2cpush

import (
	"fmt"

	"github.com/xuning888/helloIMClient/internal/proto"
	"github.com/xuning888/helloIMClient/protocol"
	"google.golang.org/protobuf/proto"
)

var _ protocol.Request = &Request{}
var _ protocol.Response = &Response{}

// Request 单聊下行消息ACK，由客户端发送
type Request struct {
	*helloim_proto.C2CPushResponse
}

func (r *Request) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_C2CPUSH)
}

// Response 单聊下行消息，由IM推送
type Response struct {
	*helloim_proto.C2CPushRequest
	msgSeq int32
}

func (r *Response) MsgSeq() int32 {
	return r.msgSeq
}

func (r *Response) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_C2CPUSH)
}

func (r *Response) ServerSeq() int64 {
	return r.GetServerSeq()
}

func (r *Response) MsgId() int64 {
	return r.GetMsgId()
}

func NewRequest(from, to int64, msgId int64) *Request {
	req := &Request{
		C2CPushResponse: &helloim_proto.C2CPushResponse{
			From:  fmt.Sprintf("%d", from),
			To:    fmt.Sprintf("%d", to),
			MsgId: msgId,
		},
	}
	return req
}

func RequestDecode(frame *protocol.Frame) (protocol.Request, error) {
	body := frame.Body
	c2cPushResponse := helloim_proto.C2CPushResponse{}
	err := proto.Unmarshal(body, &c2cPushResponse)
	if err != nil {
		return nil, err
	}
	req := &Request{
		C2CPushResponse: &c2cPushResponse,
	}
	return req, nil
}

func ResponseDecode(frame *protocol.Frame) (protocol.Response, error) {
	body := frame.Body
	c2cPushRequest := helloim_proto.C2CPushRequest{}
	err := proto.Unmarshal(body, &c2cPushRequest)
	if err != nil {
		return nil, err
	}
	return &Response{
		C2CPushRequest: &c2cPushRequest,
		msgSeq:         frame.Header.Seq,
	}, nil
}

func init() {
	protocol.RegisterDecoder(int32(helloim_proto.CmdId_CMD_ID_C2CPUSH), &protocol.Decoder{
		RequestDecode:  RequestDecode,
		ResponseDecode: ResponseDecode,
	})
}
