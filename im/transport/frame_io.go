package transport

import (
	"github.com/panjf2000/gnet/v2"
	"github.com/xuning888/helloIMClient/im/protocol"
)

// readFrame 从 socket 读取一个完整帧
func readFrame(conn gnet.Conn) (*protocol.Frame, int, gnet.Action) {
	hsize := int(protocol.DefaultHeaderSize)
	if conn.InboundBuffered() < hsize {
		return nil, 0, gnet.None
	}
	buf, err := conn.Peek(hsize)
	if err != nil {
		return nil, 0, gnet.None
	}
	header := protocol.DecodeHeader(buf)
	frameSize := int(header.BodyLength) + hsize
	if conn.InboundBuffered() < frameSize {
		return nil, 0, gnet.None
	}
	if _, err = conn.Discard(hsize); err != nil {
		return nil, 0, gnet.Close
	}
	body := make([]byte, header.BodyLength)
	if _, err = conn.Read(body); err != nil {
		return nil, 0, gnet.Close
	}
	return &protocol.Frame{Header: header, Body: body}, frameSize, gnet.None
}

// writeFrame 写字节到 socket
func writeFrame(conn gnet.Conn, data []byte) error {
	return conn.AsyncWrite(data, nil)
}
