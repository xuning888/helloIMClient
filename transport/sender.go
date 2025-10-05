package transport

import (
	"context"
	"errors"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/protocol"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrClosed = errors.New("imClient closed")
)

type sender struct {
	transport   *Transport    // gnet 传输层
	queue       *syncQueue    // 飞行队列
	lingerMs    time.Duration // 等待时间
	maxAttempts int           // 最大充实次数

	closeOnce sync.Once      // 确保只关闭一次
	wg        sync.WaitGroup // 等待goroutine退出
	closed    int32          // 原子关闭标志
}

func (s *sender) writeMessage(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	item := newSyncItem(request)
	if putSuccess := s.queue.put(item); !putSuccess {
		return nil, ErrClosed
	}
	timeout, cancelFunc := context.WithTimeout(ctx, time.Second*3)
	defer cancelFunc()
	if err := item.await(timeout); err != nil {
		return nil, err
	}
	return item.response, item.err
}

func (s *sender) writeRequest() {
	for {
		item := s.queue.get()
		if item == nil {
			select {
			// 没有消息就等待在这里等待下
			case <-time.After(s.lingerMs):
				// 如果sender关闭就退出
				if atomic.LoadInt32(&s.closed) == 1 {
					return
				}
				continue
			}
		}
		// 处理请求
		s.doWrite(item)
	}
}

func (s *sender) close() {
	s.closeOnce.Do(func() {
		atomic.StoreInt32(&s.closed, 1)
		s.queue.Close()
		waitDone := make(chan struct{})
		go func() {
			s.wg.Wait()
			close(waitDone)
		}()
		select {
		case <-waitDone:
		case <-time.After(5 * time.Second):
		}
	})
}

func (s *sender) run(f func()) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		f()
	}()
}

func (s *sender) doWrite(item *syncItem) {
	if item == nil {
		return
	}
	var lastErr error
	for attempt, maxAttempts := 0, s.maxAttempts; attempt < maxAttempts; attempt++ {
		if attempt != 0 {
			// 指数退避
			delay := backoff(attempt, 100*time.Millisecond, 1*time.Second)
			time.Sleep(delay)
		}
		// 发送消息
		lastErr = s.transport.roundTrip(context.Background(), item)
		if lastErr == nil {
			break
		}
		logger.Infof("doWrite")
	}
	if lastErr != nil {
		if item.response == nil {
			item.errors(lastErr)
		}
	}
}

func backoff(attempt int, min time.Duration, max time.Duration) time.Duration {
	d := time.Duration(attempt*attempt) * min
	if d > max {
		d = max
	}
	return d
}

func newSender(transport *Transport, lingerMs time.Duration, maxAttempts int, seq int32) *sender {
	sender := &sender{
		transport:   transport,
		lingerMs:    lingerMs,
		maxAttempts: maxAttempts,
		queue:       newSyncQueue(1000, seq),
		wg:          sync.WaitGroup{},
	}
	// 开启一个 goroutine 处理写请求
	sender.run(sender.writeRequest)
	return sender
}
