package frame

import (
	"encoding/binary"
	"fmt"
)

var DefaultHeaderSize = 20

type MsgHeader struct {
	HeaderLength  int32
	ClientVersion int32
	Seq           int32
	CmdId         int32
	BodyLength    int32
}

type Frame struct {
	Header *MsgHeader
	Body   []byte
}

func NewMsgHeader(clientVersion, seq int32, cmdId, bodyLength int) *MsgHeader {
	h := &MsgHeader{
		HeaderLength:  int32(DefaultHeaderSize),
		ClientVersion: clientVersion,
		Seq:           seq,
		CmdId:         int32(cmdId),
		BodyLength:    int32(bodyLength),
	}
	return h
}

func EncodeHeader(h *MsgHeader) []byte {
	buf := make([]byte, DefaultHeaderSize)
	binary.BigEndian.PutUint32(buf[0:4], uint32(h.HeaderLength))
	binary.BigEndian.PutUint32(buf[4:8], uint32(h.ClientVersion))
	binary.BigEndian.PutUint32(buf[8:12], uint32(h.Seq))
	binary.BigEndian.PutUint32(buf[12:16], uint32(h.CmdId))
	binary.BigEndian.PutUint32(buf[16:20], uint32(h.BodyLength))
	return buf
}

func DecodeHeader(data []byte) *MsgHeader {
	return &MsgHeader{
		HeaderLength:  int32(binary.BigEndian.Uint32(data[0:4])),
		ClientVersion: int32(binary.BigEndian.Uint32(data[4:8])),
		Seq:           int32(binary.BigEndian.Uint32(data[8:12])),
		CmdId:         int32(binary.BigEndian.Uint32(data[12:16])),
		BodyLength:    int32(binary.BigEndian.Uint32(data[16:20])),
	}
}

func (frame *Frame) ToBytes() []byte {
	if frame == nil {
		return nil
	}
	h := frame.Header
	bytes := append([]byte{}, EncodeHeader(h)...)
	bytes = append(bytes, frame.Body...)
	return bytes
}

func (frame *Frame) ToBytesV2() []byte {
	if frame == nil {
		return nil
	}
	h := frame.Header
	bytes := make([]byte, DefaultHeaderSize+len(frame.Body))
	copy(bytes, EncodeHeader(h))
	copy(bytes[20:], frame.Body)
	return bytes
}

// Key 对于客户端来说cmdId + seq 就可以确定一条消息的request 和 response
func (frame *Frame) Key() string {
	cmdId, seq := frame.Header.CmdId, frame.Header.Seq
	return fmt.Sprintf("%d_%d", cmdId, seq)
}
