package push

import (
	"github.com/xuning888/helloIMClient/internal/proto"
	"github.com/xuning888/helloIMClient/protocol"
	"google.golang.org/protobuf/proto"
)

var _ protocol.Request = &Request{}
var _ protocol.Response = &Response{}

// Request push 下行 ACK，由客户端返回
type Request struct {
	*helloim_proto.PushPktResponse
}

func (r *Request) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_PUSH)
}

// Response 统一推送下行消息（单聊/群聊共用 CMD_ID_PUSH）
type Response struct {
	*helloim_proto.PushPktRequest
	msgSeq int32
}

func (r *Response) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_PUSH)
}

func (r *Response) ServerSeq() int64 {
	return r.GetServerSeq()
}

func (r *Response) MsgId() int64 {
	return r.GetMsgId()
}

func (r *Response) MsgSeq() int32 {
	return r.msgSeq
}

func RequestDecode(frame *protocol.Frame) (protocol.Request, error) {
	req := &helloim_proto.PushPktResponse{}
	if err := proto.Unmarshal(frame.Body, req); err != nil {
		return nil, err
	}
	return &Request{PushPktResponse: req}, nil
}

func ResponseDecode(frame *protocol.Frame) (protocol.Response, error) {
	resp := &helloim_proto.PushPktRequest{}
	if err := proto.Unmarshal(frame.Body, resp); err != nil {
		return nil, err
	}
	return &Response{
		PushPktRequest: resp,
		msgSeq:         frame.Header.Seq,
	}, nil
}

func init() {
	protocol.RegisterDecoder(int32(helloim_proto.CmdId_CMD_ID_PUSH), &protocol.Decoder{
		RequestDecode:  RequestDecode,
		ResponseDecode: ResponseDecode,
	})
}
