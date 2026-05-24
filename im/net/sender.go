package im

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/protocol"
)

type DispatchMsg func(frame *protocol.Frame)

type MsgSender struct {
	log         logger.Logger
	respChan    chan *protocol.Frame
	ctx         context.Context
	cancel      context.CancelFunc
	requests    sync.Map
	mux         sync.Mutex
	dispatchMsg DispatchMsg
}

func NewMsgSender(dispatchMsg DispatchMsg) *MsgSender {
	ms := &MsgSender{
		log:         logger.Named("sender"),
		respChan:    make(chan *protocol.Frame, 1),
		requests:    sync.Map{},
		mux:         sync.Mutex{},
		dispatchMsg: dispatchMsg,
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	ms.ctx = ctx
	ms.cancel = cancelFunc
	go ms.dispatch()
	return ms
}

func (mm *MsgSender) syncWithRetry(ctx context.Context,
	conn *HiConn, request *protocol.Frame, readTimeout time.Duration, maxRetry int) (*protocol.Frame, error) {

	var resp *protocol.Frame
	var err error
	for attempt := 0; attempt < maxRetry; attempt++ {
		if attempt != 0 {
			delay := backoff(attempt, 100*time.Millisecond, 1*time.Second)
			time.Sleep(delay)
		}
		resp, err = mm.sync(ctx, conn, request, readTimeout)
		if err != nil {
			mm.log.Errorf("sync error: %v", err)
			continue
		}
		break
	}
	return resp, err
}

func (mm *MsgSender) sync(ctx context.Context,
	conn *HiConn, request *protocol.Frame, readTimeout time.Duration) (resp *protocol.Frame, err error) {
	seq := request.Header.Seq
	// 创建promise
	var p promise = make(async, 1)
	mm.requests.Store(seq, p)
	defer mm.requests.Delete(seq)
	bytes := protocol.ToBytes(request)
	// 写出数据
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err := mm.write(conn, bytes); err != nil {
			return nil, err
		}
	}
	// 等待响应
	timeoutCtx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()
	return p.await(timeoutCtx)
}

func (mm *MsgSender) oneway(ctx context.Context, conn *HiConn, request *protocol.Frame) error {
	bytes := protocol.ToBytes(request)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		err := mm.write(conn, bytes)
		if err != nil {
			return fmt.Errorf("write to connection failed: %w", err)
		}
		return nil
	}
}

// write 同步写出
func (mm *MsgSender) write(conn *HiConn, msg []byte) error {
	resultChan := make(chan error, 1)
	err := conn.asyncMessage(msg, func(err error) {
		resultChan <- err
	})
	if err != nil {
		close(resultChan)
		return err
	}
	select {
	case err = <-resultChan:
		return err
	}
}

// onMessage 收到消息
func (mm *MsgSender) onMessage(hc *HiConn) (action gnet.Action) {
	conn := hc.getConn()
	var hsize = int(protocol.DefaultHeaderSize)
	for {
		if conn.InboundBuffered() < hsize {
			action = gnet.None
			return
		}
		buf, err := conn.Peek(hsize)
		if err != nil {
			action = gnet.None
			return
		}
		header := protocol.DecodeHeader(buf)
		frameSize := int(header.BodyLength) + hsize
		if conn.InboundBuffered() < frameSize {
			action = gnet.None
			return
		}
		if _, err = conn.Discard(hsize); err != nil {
			action = gnet.Close
			return
		}
		body := make([]byte, header.BodyLength)
		if _, err = conn.Read(body); err != nil {
			action = gnet.Close
			return
		}
		mm.respChan <- &protocol.Frame{
			Header: header,
			Body:   body,
		}
	}
}

func (mm *MsgSender) dispatch() {
	for {
		select {
		case <-mm.ctx.Done():
			close(mm.respChan)
			return
		case resp := <-mm.respChan:
			if resp.Header.Req == protocol.RES {
				mm.complete(resp)
			} else {
				mm.dispatchMsg(resp)
			}
		}
	}
}

func (mm *MsgSender) complete(ack *protocol.Frame) {
	seq := ack.Header.Seq
	if value, ok := mm.requests.Load(seq); ok {
		if p, ok := value.(async); ok {
			p.complete(ack)
		}
	}
}

type promise interface {
	await(ctx context.Context) (resp *protocol.Frame, err error)
}

type async chan interface{}

func (p async) await(ctx context.Context) (resp *protocol.Frame, err error) {
	select {
	case x := <-p:
		switch v := x.(type) {
		case nil:
			return nil, nil
		case *protocol.Frame:
			return v, nil
		case error:
			return nil, err
		default:
			return nil, fmt.Errorf("未知的类型: %T", v)
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p async) complete(resp *protocol.Frame) {
	p <- resp
}

func (p async) error(err error) {
	p <- err
}

func backoff(attempt int, min time.Duration, max time.Duration) time.Duration {
	d := time.Duration(attempt*attempt) * min
	if d > max {
		d = max
	}
	return d
}
