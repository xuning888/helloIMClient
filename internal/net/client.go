package net

import (
	"log"
	"time"

	"github.com/panjf2000/gnet/v2"
)

type ImClient struct {
	gnet.BuiltinEventEngine
	cli     *gnet.Client // gnet客户端
	replies chan *Frame  // 接收缓冲区
	conn    gnet.Conn    // 长连接
}

func (imCli *ImClient) OnBoot(eng gnet.Engine) gnet.Action {
	return gnet.None
}

func (imCli *ImClient) OnOpen(conn gnet.Conn) (out []byte, action gnet.Action) {
	imCli.conn = conn // 连接建立
	return nil, gnet.None
}

func (imCli *ImClient) OnClose(conn gnet.Conn) gnet.Action {
	return gnet.None
}

func (imCli *ImClient) OnTraffic(conn gnet.Conn) gnet.Action {
	for {
		if conn.InboundBuffered() < DefaultHeaderLength {
			return gnet.None
		}
		buf, err := conn.Peek(DefaultHeaderLength)
		if err != nil {
			log.Printf("Peek Header error: %v\n", err)
			return gnet.Close
		}
		header := decodeHeader(buf)
		frameSize := int(header.BodyLength) + DefaultHeaderLength
		if frameSize < conn.InboundBuffered() {
			return gnet.None
		}
		if _, err2 := conn.Discard(DefaultHeaderLength); err2 != nil {
			log.Printf("Discard header failed, error: %v\n", err2)
			return gnet.Close
		}
		body := make([]byte, header.BodyLength)
		_, err = conn.Read(body)
		if err != nil {
			log.Printf("Read body failed, error: %v\n", err)
			return gnet.Close
		}
		frame := &Frame{
			Header: header,
			Body:   body,
		}
		imCli.replies <- frame
	}
}

func (imCli *ImClient) OnTick() (delay time.Duration, action gnet.Action) {
	return time.Second * 10, gnet.None
}

func NewImClient() *ImClient {
	return &ImClient{
		replies: make(chan *Frame, 100),
	}
}
