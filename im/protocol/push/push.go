package push

import (
	"github.com/xuning888/helloIMClient/im/proto"
	"github.com/xuning888/helloIMClient/im/protocol"
	"google.golang.org/protobuf/proto"
)

// RecvMsg 统一推送下行消息（单聊/群聊共用 CMD_ID_PUSH）
type RecvMsg struct {
	*helloim_proto.PushPktRequest
	msgSeq int32
}

func (m *RecvMsg) CmdId() int32     { return int32(helloim_proto.CmdId_CMD_ID_PUSH) }
func (m *RecvMsg) MsgId() int64     { return m.GetMsgId() }
func (m *RecvMsg) MsgSeq() int32    { return m.msgSeq }
func (m *RecvMsg) ServerSeq() int64 { return m.GetServerSeq() }

// RecvAck push 下行 ACK，由客户端返回
type RecvAck struct {
	*helloim_proto.PushPktResponse
}

func (m *RecvAck) CmdId() int32 { return int32(helloim_proto.CmdId_CMD_ID_PUSH) }

func decodePush(frame *protocol.Frame) (protocol.Message, error) {
	resp := &helloim_proto.PushPktRequest{}
	if err := proto.Unmarshal(frame.Body, resp); err != nil {
		return nil, err
	}
	return &RecvMsg{PushPktRequest: resp, msgSeq: frame.Header.Seq}, nil
}

func init() {
	protocol.RegisterDecoder(int32(helloim_proto.CmdId_CMD_ID_PUSH), decodePush)
}
