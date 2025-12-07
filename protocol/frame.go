package protocol

import (
	"encoding/binary"
	"fmt"
)

// DefaultHeaderSize 固定消息头的长度
var DefaultHeaderSize = 20
var ClientVersion int32 = 0

// MsgHeader 固定消息头
type MsgHeader struct {
	HeaderLength  int32
	ClientVersion int32
	Seq           int32
	CmdId         int32
	BodyLength    int32
}

// Frame 数据帧，最终转换为bytes数组发送到IM服务端
type Frame struct {
	Header *MsgHeader
	Body   []byte
}

func (f *Frame) Key() string {
	h := f.Header
	return fmt.Sprintf("%d_%d", h.Seq, h.CmdId)
}

// EncodeHeader 编码固定消息头
func EncodeHeader(h *MsgHeader) []byte {
	buf := make([]byte, DefaultHeaderSize)
	binary.BigEndian.PutUint32(buf[0:4], uint32(h.HeaderLength))
	binary.BigEndian.PutUint32(buf[4:8], uint32(h.ClientVersion))
	binary.BigEndian.PutUint32(buf[8:12], uint32(h.Seq))
	binary.BigEndian.PutUint32(buf[12:16], uint32(h.CmdId))
	binary.BigEndian.PutUint32(buf[16:20], uint32(h.BodyLength))
	return buf
}

// DecodeHeader 解码固定消息头
func DecodeHeader(data []byte) *MsgHeader {
	return &MsgHeader{
		HeaderLength:  int32(binary.BigEndian.Uint32(data[0:4])),
		ClientVersion: int32(binary.BigEndian.Uint32(data[4:8])),
		Seq:           int32(binary.BigEndian.Uint32(data[8:12])),
		CmdId:         int32(binary.BigEndian.Uint32(data[12:16])),
		BodyLength:    int32(binary.BigEndian.Uint32(data[16:20])),
	}
}
