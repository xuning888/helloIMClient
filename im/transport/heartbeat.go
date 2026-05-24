package transport

import (
	"github.com/xuning888/helloIMClient/im/proto"
	"github.com/xuning888/helloIMClient/im/protocol"
	"google.golang.org/protobuf/proto"
)

type HeartbeatRequest struct {
	*helloim_proto.EmptyRequest
}

func (r *HeartbeatRequest) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_HEARTBEAT)
}

type HeartbeatResponse struct {
	*helloim_proto.EmptyResponse
	f *protocol.Frame
}

func (r *HeartbeatResponse) CmdId() int32  { return int32(helloim_proto.CmdId_CMD_ID_HEARTBEAT) }
func (r *HeartbeatResponse) MsgSeq() int32 { return r.f.Header.Seq }

func NewHeartbeatRequest() *HeartbeatRequest {
	return &HeartbeatRequest{EmptyRequest: &helloim_proto.EmptyRequest{}}
}

func decodeHeartbeat(frame *protocol.Frame) (protocol.Message, error) {
	resp := &helloim_proto.EmptyResponse{}
	if err := proto.Unmarshal(frame.Body, resp); err != nil {
		return nil, err
	}
	return &HeartbeatResponse{EmptyResponse: resp, f: frame}, nil
}

func init() {
	protocol.RegisterDecoder(int32(helloim_proto.CmdId_CMD_ID_HEARTBEAT), decodeHeartbeat)
}
