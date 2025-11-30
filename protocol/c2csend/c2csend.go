package c2csend

import (
	"fmt"
	"time"

	"github.com/xuning888/helloIMClient/internal/proto"
	"github.com/xuning888/helloIMClient/protocol"
	"google.golang.org/protobuf/proto"
)

var _ protocol.Request = &Request{}
var _ protocol.Response = &Response{}

// Request 单聊上行
type Request struct {
	*helloim_proto.C2CSendRequest
}

func (req *Request) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_C2CSEND)
}

// Response 单聊下行
// Note: 单聊上行和单聊下行的seq和cmdId是相同的
type Response struct {
	*helloim_proto.C2CSendResponse
	msgSeq int32
}

func (res *Response) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_C2CSEND)
}

func (res *Response) ServerSeq() int64 {
	return res.C2CSendResponse.ServerSeq
}

func (res *Response) MsgId() int64 {
	return res.C2CSendResponse.MsgId
}

func (res *Response) MsgSeq() int32 {
	return res.msgSeq
}

func NewRequest(from, to int64, content string,
	contentType, fromUserType, toUserType int32) *Request {
	req := &Request{
		C2CSendRequest: &helloim_proto.C2CSendRequest{
			From:          fmt.Sprintf("%d", from),
			To:            fmt.Sprintf("%d", to),
			Content:       content,
			ContentType:   contentType,
			SendTimestamp: time.Now().UnixMilli(),
			FromUserType:  fromUserType,
			ToUserType:    toUserType,
		},
	}
	return req
}

func RequestDecode(frame *protocol.Frame) (protocol.Request, error) {
	body := frame.Body
	c2cSendRequest := helloim_proto.C2CSendRequest{}
	err := proto.Unmarshal(body, &c2cSendRequest)
	if err != nil {
		return nil, err
	}
	req := &Request{
		C2CSendRequest: &c2cSendRequest,
	}
	return req, nil
}

func ResponseDecode(frame *protocol.Frame) (protocol.Response, error) {
	body := frame.Body
	c2cSendResponse := helloim_proto.C2CSendResponse{}
	err := proto.Unmarshal(body, &c2cSendResponse)
	if err != nil {
		return nil, err
	}
	return &Response{
		C2CSendResponse: &c2cSendResponse,
		msgSeq:          frame.Header.Seq,
	}, nil
}

func init() {
	protocol.RegisterDecoder(int32(helloim_proto.CmdId_CMD_ID_C2CSEND), &protocol.Decoder{
		RequestDecode:  RequestDecode,
		ResponseDecode: ResponseDecode,
	})
}
