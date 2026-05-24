package send

import (
	"fmt"
	"time"

	"github.com/xuning888/helloIMClient/internal/proto"
	"github.com/xuning888/helloIMClient/protocol"
	"google.golang.org/protobuf/proto"
)

var _ protocol.Request = &Request{}
var _ protocol.Response = &Response{}

// Request 统一上行消息（单聊/群聊共用 CMD_ID_SEND）
type Request struct {
	*helloim_proto.SendPktRequest
}

func (r *Request) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_SEND)
}

// Response 统一上行ACK
type Response struct {
	*helloim_proto.SendPktResponse
	msgSeq int32
}

func (r *Response) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_SEND)
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

// NewRequest 创建发送请求
// chatType: 1=单聊, 2=群聊
func NewRequest(from int64, chatId int64, chatType int32,
	payload *helloim_proto.Payload,
	fromUserType, toUserType int32) *Request {
	return &Request{
		SendPktRequest: &helloim_proto.SendPktRequest{
			From:          fmt.Sprintf("%d", from),
			FromUserType:  fromUserType,
			ChatId:        fmt.Sprintf("%d", chatId),
			ChatType:      chatType,
			SendTimestamp: time.Now().UnixMilli(),
			ToUserType:    toUserType,
			Payload:       payload,
		},
	}
}

func RequestDecode(frame *protocol.Frame) (protocol.Request, error) {
	req := &helloim_proto.SendPktRequest{}
	if err := proto.Unmarshal(frame.Body, req); err != nil {
		return nil, err
	}
	return &Request{SendPktRequest: req}, nil
}

func ResponseDecode(frame *protocol.Frame) (protocol.Response, error) {
	resp := &helloim_proto.SendPktResponse{}
	if err := proto.Unmarshal(frame.Body, resp); err != nil {
		return nil, err
	}
	return &Response{
		SendPktResponse: resp,
		msgSeq:          frame.Header.Seq,
	}, nil
}

func init() {
	protocol.RegisterDecoder(int32(helloim_proto.CmdId_CMD_ID_SEND), &protocol.Decoder{
		RequestDecode:  RequestDecode,
		ResponseDecode: ResponseDecode,
	})
}
