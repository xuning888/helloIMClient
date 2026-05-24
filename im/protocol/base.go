package protocol

import (
	im "github.com/xuning888/helloIMClient/internal/proto"
	"github.com/xuning888/helloIMClient/protocol"
)

const (
	DefaultHeaderLength int = 14
)

type BaseMsg struct {
	Seq       int32               // 客户端消息序号
	CmdId     int32               // 指令号
	MsgId     int64               // 消息唯一id
	ServerSeq int64               // 服务端消息序号
	MsgFrom   string              // 消息发送着
	ChatId    string              // 会话id
	ChatType  int32               // 会话类型
	Header    *protocol.MsgHeader // 消息头
	Payload   *im.Payload         // 消息正文
}

func (bm *BaseMsg) FixedHeaderLength() int {
	return DefaultHeaderLength
}
