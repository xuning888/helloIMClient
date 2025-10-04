package c2csend

import (
	pb "github.com/xuning888/helloIMClient/proto"
	"github.com/xuning888/helloIMClient/protocol"
	"google.golang.org/protobuf/proto"
)

var _ protocol.Request = &Request{}
var _ protocol.Response = &Response{}

// Request 单聊上行
type Request struct {
	*pb.C2CSendRequest
}

func (req *Request) CmdId() int32 {
	return int32(pb.CmdId_CMD_ID_C2CSEND)
}

// Response 单聊下行
// Note: 单聊上行和单聊下行的seq和cmdId是相同的
type Response struct {
	*pb.C2CSendResponse
}

func (res *Response) CmdId() int32 {
	return int32(pb.CmdId_CMD_ID_C2CSEND)
}

func (res *Response) ServerSeq() int64 {
	return res.C2CSendResponse.ServerSeq
}

func (res *Response) MsgId() int64 {
	return res.C2CSendResponse.MsgId
}

func RequestDecode(frame *protocol.Frame) (protocol.Request, error) {
	body := frame.Body
	c2cSendRequest := pb.C2CSendRequest{}
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
	c2cSendResponse := pb.C2CSendResponse{}
	err := proto.Unmarshal(body, &c2cSendResponse)
	if err != nil {
		return nil, err
	}
	return &Response{
		C2CSendResponse: &c2cSendResponse,
	}, nil
}

func init() {
	protocol.RegisterDecoder(int32(pb.CmdId_CMD_ID_C2CSEND), &protocol.Decoder{
		RequestDecode:  RequestDecode,
		ResponseDecode: ResponseDecode,
	})
}
