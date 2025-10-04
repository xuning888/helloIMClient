package protocol

import (
	"fmt"
	"google.golang.org/protobuf/proto"
)

// Message 是IM消息的抽象
type Message interface {
	proto.Message
	// CmdId 消息的指令号
	CmdId() int32
}

// Request 表示上行IM消息: 客户端发送给服务端到消息
type Request interface {
	Message
}

// Response 表示IM下行消息: IM服务端发送给客户端的消息
// Note: Response 有两种类型:
//  1. 上行消息的ACK: 请求响应模式下, 客户端发送消息, 同步等待IM服务端返回response
//  2. 下行消息推送: IM服务端主动推送消息。
//     比如: 用户u1发送消息，u2收到response，u2还需要向IM服务度发送一个request作为应答。
type Response interface {
	Message
	// ServerSeq 服务端为消息分配到序号
	ServerSeq() int64
	// MsgId 服务端为消息分配到唯一ID
	MsgId() int64
}

type DecodeResponse func(frame *Frame) (Response, error)
type DecodeRequest func(frame *Frame) (Request, error)

type Decoder struct {
	RequestDecode  DecodeRequest
	ResponseDecode DecodeResponse
}

var decoders = make(map[int32]*Decoder)

func RegisterDecoder(cmdId int32, decoder *Decoder) {
	decoders[cmdId] = decoder
}

func EncodeMessageToBytes(seq int32, message Message) ([]byte, error) {
	body, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	h := &MsgHeader{
		HeaderLength:  int32(DefaultHeaderSize),
		ClientVersion: ClientVersion,
		Seq:           seq,
		CmdId:         message.CmdId(),
		BodyLength:    int32(len(body)),
	}
	bytes := make([]byte, DefaultHeaderSize+len(body))
	copy(bytes, EncodeHeader(h))
	copy(bytes[20:], body)
	return bytes, nil
}

func DecodeResp(frame *Frame) (Response, error) {
	cmdId := frame.Header.CmdId
	decoder := decoders[cmdId]
	if decoder.ResponseDecode != nil {
		return decoder.ResponseDecode(frame)
	}
	return nil, fmt.Errorf("cmdId: %d, unsupport decode resposne", cmdId)
}
