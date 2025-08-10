package net

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
