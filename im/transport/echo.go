package transport

import (
	"github.com/xuning888/helloIMClient/im/proto"
	"github.com/xuning888/helloIMClient/im/protocol"
	"google.golang.org/protobuf/proto"
)

type EchoRequest struct {
	*helloim_proto.EchoRequest
}

func (r *EchoRequest) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_ECHO)
}

type EchoResponse struct {
	*helloim_proto.EchoResponse
	f *protocol.Frame
}

func (r *EchoResponse) CmdId() int32  { return int32(helloim_proto.CmdId_CMD_ID_ECHO) }
func (r *EchoResponse) MsgSeq() int32 { return r.f.Header.Seq }

func NewEchoRequest() *EchoRequest {
	return &EchoRequest{EchoRequest: &helloim_proto.EchoRequest{Msg: ""}}
}

func decodeEcho(frame *protocol.Frame) (protocol.Message, error) {
	resp := &helloim_proto.EchoResponse{}
	if err := proto.Unmarshal(frame.Body, resp); err != nil {
		return nil, err
	}
	return &EchoResponse{EchoResponse: resp, f: frame}, nil
}

func init() {
	protocol.RegisterDecoder(int32(helloim_proto.CmdId_CMD_ID_ECHO), decodeEcho)
}
