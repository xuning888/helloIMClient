package protocol

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// Message IM 消息
type Message interface {
	proto.Message
	CmdId() int32
}

type DecodeFunc func(frame *Frame) (Message, error)

var decoders = make(map[int32]DecodeFunc)

func RegisterDecoder(cmdId int32, decode DecodeFunc) {
	decoders[cmdId] = decode
}

func DecodeMessage(frame *Frame) (Message, error) {
	decode := decoders[frame.Header.CmdId]
	if decode == nil {
		return nil, fmt.Errorf("unsupported cmdId: %d", frame.Header.CmdId)
	}
	return decode(frame)
}

func EncodeMessageToBytes(seq int32, req byte, message Message) ([]byte, error) {
	body, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	h := &MsgHeader{
		HeaderLength: DefaultHeaderSize,
		Req:          req,
		Seq:          seq,
		CmdId:        message.CmdId(),
		BodyLength:   int32(len(body)),
	}
	bytes := make([]byte, int(DefaultHeaderSize)+len(body))
	copy(bytes, EncodeHeader(h))
	copy(bytes[DefaultHeaderSize:], body)
	return bytes, nil
}

func EncodeMessageToFrame(seq int32, req byte, message Message) (*Frame, error) {
	body, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	h := &MsgHeader{
		HeaderLength: DefaultHeaderSize,
		Req:          req,
		Seq:          seq,
		CmdId:        message.CmdId(),
		BodyLength:   int32(len(body)),
	}
	return &Frame{Header: h, Body: body}, nil
}

func MakeResFrame(frame *Frame) []byte {
	h := frame.Header
	header := &MsgHeader{
		HeaderLength: DefaultHeaderSize,
		Req:          RES,
		Seq:          h.Seq,
		CmdId:        h.CmdId,
		BodyLength:   0,
	}
	bytes := make([]byte, int(DefaultHeaderSize))
	copy(bytes, EncodeHeader(header))
	return bytes
}

func ToBytes(frame *Frame) []byte {
	h, body := frame.Header, frame.Body
	bytes := make([]byte, int(DefaultHeaderSize)+len(body))
	copy(bytes, EncodeHeader(h))
	copy(bytes[DefaultHeaderSize:], body)
	return bytes
}
