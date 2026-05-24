package send

import (
	"fmt"
	"time"

	"github.com/xuning888/helloIMClient/im/proto"
	"github.com/xuning888/helloIMClient/im/protocol"
	"google.golang.org/protobuf/proto"
)

// SendMsg 统一上行消息（单聊/群聊共用 CMD_ID_SEND）
type SendMsg struct {
	*helloim_proto.SendPktRequest
}

func (m *SendMsg) CmdId() int32 { return int32(helloim_proto.CmdId_CMD_ID_SEND) }

// SendAck 统一上行 ACK
type SendAck struct {
	*helloim_proto.SendPktResponse
	msgSeq int32
}

func (m *SendAck) CmdId() int32     { return int32(helloim_proto.CmdId_CMD_ID_SEND) }
func (m *SendAck) MsgId() int64     { return m.GetMsgId() }
func (m *SendAck) MsgSeq() int32    { return m.msgSeq }
func (m *SendAck) ServerSeq() int64 { return m.GetServerSeq() }

// chatType: 1=单聊, 2=群聊
func NewSendMsg(from int64, chatId int64, chatType int32,
	payload *helloim_proto.Payload,
	fromUserType, toUserType int32) *SendMsg {
	return &SendMsg{
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

func decodeSend(frame *protocol.Frame) (protocol.Message, error) {
	resp := &helloim_proto.SendPktResponse{}
	if err := proto.Unmarshal(frame.Body, resp); err != nil {
		return nil, err
	}
	return &SendAck{SendPktResponse: resp, msgSeq: frame.Header.Seq}, nil
}

func init() {
	protocol.RegisterDecoder(int32(helloim_proto.CmdId_CMD_ID_SEND), decodeSend)
}
