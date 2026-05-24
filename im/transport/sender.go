package transport

import (
	"context"
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2"
	protocol2 "github.com/xuning888/helloIMClient/im/protocol"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

// GetSeq 序号分配器
type GetSeq func() int32

type sender struct {
	log      logger.Logger
	requests sync.Map // key: int32(seq) → *promise
	getSeq   GetSeq

	// dispatch
	respChan chan *dispatchItem
	ctx      context.Context
	cancel   context.CancelFunc
	dispatch func(protocol2.Message)
}

type dispatchItem struct {
	frame *protocol2.Frame
	conn  gnet.Conn
}

func newSender(getSeq GetSeq, dispatch func(protocol2.Message)) *sender {
	ctx, cancel := context.WithCancel(context.Background())
	s := &sender{
		log:      logger.Named("sender"),
		requests: sync.Map{},
		getSeq:   getSeq,
		respChan: make(chan *dispatchItem, 100),
		ctx:      ctx,
		cancel:   cancel,
		dispatch: dispatch,
	}
	go s.dispatchLoop()
	return s
}

// send 停等协议：分配 seq，编码写出，等待 ACK
func (s *sender) send(ctx context.Context, conn gnet.Conn, msg protocol2.Message, timeout time.Duration) (protocol2.Message, error) {
	seq := s.getSeq()
	frame, err := protocol2.EncodeMessageToFrame(seq, protocol2.REQ, msg)
	if err != nil {
		return nil, err
	}
	p := newPromise()
	s.requests.Store(seq, p)
	defer s.requests.Delete(seq)

	data := protocol2.ToBytes(frame)
	if err := writeFrame(conn, data); err != nil {
		return nil, err
	}
	return p.await(ctx, timeout)
}

// sendWithRetry 带重试的停等发送
func (s *sender) sendWithRetry(ctx context.Context, conn gnet.Conn, msg protocol2.Message, timeout time.Duration, maxRetry int) (protocol2.Message, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetry; attempt++ {
		if attempt > 0 {
			delay := backoff(attempt, 100*time.Millisecond, 1*time.Second)
			time.Sleep(delay)
		}
		resp, err := s.send(ctx, conn, msg, timeout)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		s.log.Errorf("sendWithRetry attempt %d: %v", attempt, err)
	}
	return nil, lastErr
}

// complete 完成 promise（ACK 到达时调用）
func (s *sender) complete(frame *protocol2.Frame) {
	seq := frame.Header.Seq
	if val, ok := s.requests.Load(seq); ok {
		if p, ok := val.(*promise); ok {
			p.complete(frame)
		}
	}
}

// dispatchFrame 异步分发推送消息
func (s *sender) dispatchFrame(frame *protocol2.Frame, conn gnet.Conn) {
	select {
	case s.respChan <- &dispatchItem{frame: frame, conn: conn}:
	default:
		s.log.Errorf("dispatchFrame: respChan full, dropping frame")
	}
}

// sendAck 发送推送消息的 ACK
func sendAck(conn gnet.Conn, frame *protocol2.Frame) {
	ack := protocol2.MakeResFrame(frame)
	writeFrame(conn, ack)
}

// sendPing 发送心跳
func sendPing(conn gnet.Conn) error {
	ping := NewHeartbeatRequest()
	pingFrame, err := protocol2.EncodeMessageToFrame(0, protocol2.REQ, ping)
	if err != nil {
		return err
	}
	return writeFrame(conn, protocol2.ToBytes(pingFrame))
}

func (s *sender) dispatchLoop() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case item := <-s.respChan:
			msg, err := protocol2.DecodeMessage(item.frame)
			if err != nil {
				s.log.Errorf("dispatchLoop: decode error: %v", err)
				continue
			}
			if s.dispatch != nil {
				s.dispatch(msg)
			}
			// 推送消息回复 ACK
			sendAck(item.conn, item.frame)
		}
	}
}

func (s *sender) close() {
	s.cancel()
}

// promise 停等协议的等待原语
type promise struct {
	done chan struct{}
	resp *protocol2.Frame
	err  error
	mu   sync.Mutex
}

func newPromise() *promise {
	return &promise{done: make(chan struct{})}
}

func (p *promise) await(ctx context.Context, timeout time.Duration) (protocol2.Message, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	select {
	case <-p.done:
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.err != nil {
			return nil, p.err
		}
		return protocol2.DecodeMessage(p.resp)
	case <-timeoutCtx.Done():
		return nil, timeoutCtx.Err()
	}
}

func (p *promise) complete(frame *protocol2.Frame) {
	p.mu.Lock()
	p.resp = frame
	p.mu.Unlock()
	close(p.done)
}

func (p *promise) fail(err error) {
	p.mu.Lock()
	p.err = err
	p.mu.Unlock()
	close(p.done)
}

func backoff(attempt int, min, max time.Duration) time.Duration {
	d := time.Duration(attempt*attempt) * min
	if d > max {
		d = max
	}
	return d
}
