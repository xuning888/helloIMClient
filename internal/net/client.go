package net

import (
	"errors"
	"fmt"
	"log"

	"github.com/panjf2000/gnet/v2"
)

type ImClient struct {
	gnet.BuiltinEventEngine
	addr    string
	cli     *gnet.Client
	conn    gnet.Conn
	replies chan *Frame
}

func (imCli *ImClient) OnTraffic(conn gnet.Conn) gnet.Action {
	for {
		for {
			if conn.InboundBuffered() < DefaultHeaderSize {
				return gnet.None
			}
			buf, err := conn.Peek(DefaultHeaderSize)
			if err != nil {
				log.Printf("Peek Header error: %v\n", err)
				return gnet.Close
			}
			header := DecodeHeader(buf)
			frameSize := int(header.BodyLength) + DefaultHeaderSize
			if frameSize < conn.InboundBuffered() {
				return gnet.None
			}
			if _, err2 := conn.Discard(DefaultHeaderSize); err2 != nil {
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
}

func (imCli *ImClient) Start() error {
	cli, err := gnet.NewClient(imCli)
	if err != nil {
		return err
	}
	imCli.cli = cli
	err = cli.Start()
	if err != nil {
		return err
	}
	conn, err := cli.Dial("tcp", imCli.addr)
	if err != nil {
		return err
	}
	imCli.conn = conn
	return nil
}

func (imCli *ImClient) SendFrame(frame *Frame) error {
	if frame == nil {
		return errors.New("frame is nil")
	}
	msg := ToBytes(frame)
	_, err := imCli.conn.Write(msg)
	if err != nil {
		return err
	}
	return nil
}

func (imCli *ImClient) Process() {
	go func() {
		for {
			select {
			case reply := <-imCli.replies:
				if reply.Header.CmdId == 1 {
					fmt.Printf("echo: %s", string(reply.Body))
				}
			default:
			}
		}
	}()
}

func NewImClient(addr string) *ImClient {
	return &ImClient{
		addr:    addr,
		replies: make(chan *Frame, 100),
	}
}
