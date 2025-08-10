package net

import "encoding/binary"

var DefaultHeaderLength = 20

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

func encodeHeader(h *MsgHeader) []byte {
	buf := make([]byte, DefaultHeaderLength)
	binary.BigEndian.PutUint32(buf[0:4], uint32(h.HeaderLength))
	binary.BigEndian.PutUint32(buf[4:8], uint32(h.ClientVersion))
	binary.BigEndian.PutUint32(buf[8:12], uint32(h.Seq))
	binary.BigEndian.PutUint32(buf[12:16], uint32(h.CmdId))
	binary.BigEndian.PutUint32(buf[16:20], uint32(h.BodyLength))
	return buf
}

func decodeHeader(data []byte) *MsgHeader {
	return &MsgHeader{
		HeaderLength:  int32(binary.BigEndian.Uint32(data[0:4])),
		ClientVersion: int32(binary.BigEndian.Uint32(data[4:8])),
		Seq:           int32(binary.BigEndian.Uint32(data[8:12])),
		CmdId:         int32(binary.BigEndian.Uint32(data[12:16])),
		BodyLength:    int32(binary.BigEndian.Uint32(data[16:20])),
	}
}
