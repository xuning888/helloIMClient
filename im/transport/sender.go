package transport

import (
	"context"
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/xuning888/helloIMClient/im/protocol"
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
	dispatch func(protocol.Message)
}

type dispatchItem struct {
	frame *protocol.Frame
	conn  gnet.Conn
}

func newSender(getSeq GetSeq, dispatch func(protocol.Message)) *sender {
	ctx, cancel := context.WithCancel(context.Background())
	s := &sender{
		log:      logger.Named("sender"),
		requests: sync.Map{},
		getSeq:   getSeq,
		respChan: make(chan *dispatchItem, 5000),
		ctx:      ctx,
		cancel:   cancel,
		dispatch: dispatch,
	}
	s.startDispatchWorkers(10)
	return s
}

func (s *sender) startDispatchWorkers(n int) {
	for i := 0; i < n; i++ {
		go s.dispatchWorker()
	}
}

// send 停等协议：分配 seq，编码写出，等待 ACK
func (s *sender) send(ctx context.Context, conn gnet.Conn, msg protocol.Message, timeout time.Duration) (protocol.Message, error) {
	seq := s.getSeq()
	frame, err := protocol.EncodeMessageToFrame(seq, protocol.REQ, msg)
	if err != nil {
		return nil, err
	}
	p := newPromise()
	s.requests.Store(seq, p)
	defer s.requests.Delete(seq)

	data := protocol.ToBytes(frame)
	if err := writeFrame(conn, data); err != nil {
		return nil, err
	}
	return p.await(ctx, timeout)
}

// sendWithRetry 带重试的停等发送
func (s *sender) sendWithRetry(ctx context.Context, conn gnet.Conn, msg protocol.Message, timeout time.Duration, maxRetry int) (protocol.Message, error) {
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
func (s *sender) complete(frame *protocol.Frame) {
	seq := frame.Header.Seq
	if val, ok := s.requests.Load(seq); ok {
		if p, ok := val.(*promise); ok {
			p.complete(frame)
		}
	}
}

// dispatchFrame 异步分发推送消息（非阻塞，不够缓冲时起 goroutine 写）
func (s *sender) dispatchFrame(frame *protocol.Frame, conn gnet.Conn) {
	item := &dispatchItem{frame: frame, conn: conn}
	select {
	case s.respChan <- item:
	default:
		go func() { s.respChan <- item }()
	}
}

// sendAck 发送推送消息的 ACK
func sendAck(conn gnet.Conn, frame *protocol.Frame) {
	ack := protocol.MakeResFrame(frame)
	writeFrame(conn, ack)
}

// sendPing 发送心跳
func sendPing(conn gnet.Conn) error {
	ping := NewHeartbeatRequest()
	pingFrame, err := protocol.EncodeMessageToFrame(0, protocol.REQ, ping)
	if err != nil {
		return err
	}
	return writeFrame(conn, protocol.ToBytes(pingFrame))
}

func (s *sender) dispatchWorker() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case item := <-s.respChan:
			// 先回 ACK，减少服务端重试
			sendAck(item.conn, item.frame)
			msg, err := protocol.DecodeMessage(item.frame)
			if err != nil {
				s.log.Errorf("dispatchWorker: decode error: %v", err)
				continue
			}
			if s.dispatch != nil {
				s.dispatch(msg)
			}
		}
	}
}

func (s *sender) close() {
	s.cancel()
}

// promise 停等协议的等待原语
type promise struct {
	done chan struct{}
	resp *protocol.Frame
	err  error
	mu   sync.Mutex
}

func newPromise() *promise {
	return &promise{done: make(chan struct{})}
}

func (p *promise) await(ctx context.Context, timeout time.Duration) (protocol.Message, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	select {
	case <-p.done:
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.err != nil {
			return nil, p.err
		}
		return protocol.DecodeMessage(p.resp)
	case <-timeoutCtx.Done():
		return nil, timeoutCtx.Err()
	}
}

func (p *promise) complete(frame *protocol.Frame) {
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
