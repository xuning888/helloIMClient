package protocol

import (
	"encoding/binary"
	"fmt"
)

// DefaultHeaderSize 固定消息头的长度
var DefaultHeaderSize byte = 14

const (
	REQ = iota
	RES
)

// MsgHeader 固定消息头
type MsgHeader struct {
	HeaderLength byte
	Req          byte
	Seq          int32
	CmdId        int32
	BodyLength   int32
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
	buf[0] = DefaultHeaderSize // 固定消息头的大小
	buf[1] = h.Req             // req or res
	binary.BigEndian.PutUint32(buf[2:6], uint32(h.Seq))
	binary.BigEndian.PutUint32(buf[6:10], uint32(h.CmdId))
	binary.BigEndian.PutUint32(buf[10:14], uint32(h.BodyLength))
	return buf
}

// DecodeHeader 解码固定消息头
func DecodeHeader(data []byte) *MsgHeader {
	return &MsgHeader{
		HeaderLength: data[0],
		Req:          data[1],
		Seq:          int32(binary.BigEndian.Uint32(data[2:6])),
		CmdId:        int32(binary.BigEndian.Uint32(data[6:10])),
		BodyLength:   int32(binary.BigEndian.Uint32(data[10:14])),
	}
}
