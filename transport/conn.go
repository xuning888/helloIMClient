package transport

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/panjf2000/gnet/v2"
	"github.com/xuning888/helloIMClient/conf"
	"github.com/xuning888/helloIMClient/protocol"
	"github.com/xuning888/helloIMClient/protocol/auth"
)

var (
	ErrNoAuth      = errors.New("conn no auth")
	ErrNotActivate = errors.New("conn not activate")
)

type Conn struct {
	conn     gnet.Conn
	activate int32 // 是否可用
	authed   int32 // 是否验证通过
	resp     chan *protocol.Frame
	requests sync.Map
	dispatch Dispatch
	closeCh  chan struct{}
}

func (c *Conn) asyncWrite(item *syncItem) error {
	if !c.Activate() {
		// 连接不可用
		return ErrNotActivate
	}
	if !c.Authed() {
		// 连接未经过认证
		return ErrNoAuth
	}
	return c.asyncWrite0(item)
}

func (c *Conn) asyncWrite0(item *syncItem) error {
	seq, request := item.seq, item.request
	bytes, err := protocol.EncodeMessageToBytes(seq, request)
	if err != nil {
		return err
	}
	_, _ = c.requests.LoadOrStore(seq, item)
	err = c.conn.AsyncWrite(bytes, c.writeCallback)
	if err != nil {
		// 加入写出队列失败
		return err
	}
	return nil
}

func (c *Conn) authReq(ctx context.Context) error {
	// TODO token
	authRequest := auth.NewRequest(conf.UserId, 0, "")
	item := newSyncItem(authRequest)
	if err := c.asyncWrite0(item); err != nil {
		return err
	}
	if err := item.await(ctx); err != nil {
		return err
	}
	response := item.response.(*auth.Response)
	if response.AuthResponse.Success {
		atomic.StoreInt32(&c.activate, 1)
		atomic.StoreInt32(&c.authed, 1)
		return nil
	}
	return ErrNoAuth
}

func (c *Conn) writeCallback(conn gnet.Conn, err error) error {
	if err != nil {
		log.Printf("writeCallback error: %v\n", err)
		c.Close()
	}
	return err
}

func (c *Conn) Close() {
	if atomic.CompareAndSwapInt32(&c.activate, 1, 0) {
		c.closeCh <- struct{}{}
		c.conn.Close()
		close(c.resp)
		close(c.closeCh)
	}
}

func (c *Conn) Activate() bool {
	return atomic.LoadInt32(&c.activate) == 1
}

func (c *Conn) Authed() bool {
	return atomic.LoadInt32(&c.authed) == 1
}

func (c *Conn) doDispatch() {
	for {
		select {
		case frame := <-c.resp:
			c.process(frame)
		case <-c.closeCh:
			return
		}
	}
}

func (c *Conn) process(frame *protocol.Frame) {
	seq := frame.Header.Seq
	response, err := protocol.DecodeResp(frame)
	if value, loaded := c.requests.LoadAndDelete(seq); loaded {
		item := value.(*syncItem)
		item.err = err
		item.complete(response)
	} else {
		r := &Result{
			resp: response,
			err:  err,
		}
		c.dispatch(r)
	}
}

func (c *Conn) run(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println(r)
			}
		}()
		f()
	}()
}

// Decode 从socket读取数据并解码
func (c *Conn) Decode() (action gnet.Action) {
	conn := c.conn
	for {
		if conn.InboundBuffered() < protocol.DefaultHeaderSize {
			action = gnet.None
			return
		}
		buf, err := conn.Peek(protocol.DefaultHeaderSize)
		if err != nil {
			action = gnet.None
			return
		}
		header := protocol.DecodeHeader(buf)
		frameSize := int(header.BodyLength) + protocol.DefaultHeaderSize
		if conn.InboundBuffered() < frameSize {
			action = gnet.None
			return
		}
		if _, err = conn.Discard(protocol.DefaultHeaderSize); err != nil {
			action = gnet.Close
			return
		}
		body := make([]byte, header.BodyLength)
		if _, err = conn.Read(body); err != nil {
			action = gnet.Close
			return
		}
		c.resp <- &protocol.Frame{
			Header: header,
			Body:   body,
		}
	}
}

func newConn(conn gnet.Conn, dispatch Dispatch) *Conn {
	pConn := &Conn{
		conn:     conn,
		activate: 1,
		resp:     make(chan *protocol.Frame, 1),
		requests: sync.Map{},
		dispatch: dispatch,
		closeCh:  make(chan struct{}, 1),
	}
	pConn.run(pConn.doDispatch)
	return pConn
}
