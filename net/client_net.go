package net

import (
	"log"

	"github.com/panjf2000/gnet/v2"
	"github.com/xuning888/helloIMClient/frame"
)

func (imCli *ImClient) OnClose(conn gnet.Conn, err error) (action gnet.Action) {
	log.Printf("conn closed error: %v", err)
	return
}

// OnTraffic 长连接可读
// 读取消息，然后放到接收缓冲区中
func (imCli *ImClient) OnTraffic(conn gnet.Conn) gnet.Action {
	for {
		for {
			if conn.InboundBuffered() < frame.DefaultHeaderSize {
				return gnet.None
			}
			buf, err := conn.Peek(frame.DefaultHeaderSize)
			if err != nil {
				log.Printf("Peek Header error: %v\n", err)
				return gnet.Close
			}
			header := frame.DecodeHeader(buf)
			frameSize := int(header.BodyLength) + frame.DefaultHeaderSize
			if conn.InboundBuffered() < frameSize {
				return gnet.None // 数据还不够
			}
			// 检查是否存在一个完整的frame
			if conn.InboundBuffered() < frameSize {
				return gnet.None
			}
			if _, err2 := conn.Discard(frame.DefaultHeaderSize); err2 != nil {
				log.Printf("Discard header failed, error: %v\n", err2)
				return gnet.Close
			}
			body := make([]byte, header.BodyLength)
			_, err = conn.Read(body)
			if err != nil {
				log.Printf("Read body failed, error: %v\n", err)
				return gnet.Close
			}
			reply := &frame.Frame{
				Header: header,
				Body:   body,
			}
			imCli.replies <- reply
		}
	}
}

// process 接受来自IM服务端的消息，然后分发出去
func (imCli *ImClient) process() {
	for {
		select {
		case f, ok := <-imCli.replies:
			if !ok {
				return
			}
			imCli.dispatch(f)
		case _ = <-imCli.stop:
			return
		}
	}
}

func (imCli *ImClient) dispatch(reply *frame.Frame) {
	// 处理消息的ACK
	imCli.inflightQ.Ack(reply)
	imCli.handle(reply)
}
