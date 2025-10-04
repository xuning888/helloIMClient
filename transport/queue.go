package transport

import (
	"context"
	"github.com/xuning888/helloIMClient/protocol"
	"sync"
	"sync/atomic"
)

type seqAdder struct {
	value atomic.Int32
}

func (a *seqAdder) seq() int32 {
	return a.value.Add(1)
}

type syncItem struct {
	mux      sync.Mutex
	request  protocol.Request
	response protocol.Response
	done     chan struct{}
	closed   bool
	err      error
	seq      int32
}

func newSyncItem(request protocol.Request) *syncItem {
	return &syncItem{
		mux:     sync.Mutex{},
		request: request,
		done:    make(chan struct{}),
	}
}

func (s *syncItem) await(ctx context.Context) error {
	select {
	case <-s.done:
		return s.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *syncItem) complete(res protocol.Response) {
	if s.closed {
		return
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	s.response = res
	close(s.done)
	s.closed = true
}

func (s *syncItem) errors(err error) {
	if s.closed {
		return
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	s.err = err
	close(s.done)
	s.closed = true
}

type syncQueue struct {
	queue  []*syncItem
	mutex  *sync.Mutex
	cond   *sync.Cond
	closed bool
	adder  *seqAdder
}

func (s *syncQueue) Put(request *syncItem) bool {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	defer s.cond.Broadcast()
	if s.closed {
		return false
	}
	request.seq = s.adder.seq()
	s.queue = append(s.queue, request)
	return true
}

func (s *syncQueue) Get() *syncItem {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	for len(s.queue) == 0 && !s.closed {
		s.cond.Wait()
	}
	if len(s.queue) == 0 {
		return nil
	}
	item := s.queue[0]
	s.queue[0] = nil
	s.queue = s.queue[1:]
	return item
}

func (s *syncQueue) Close() {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	defer s.cond.Broadcast()
	s.closed = true
}

func newSyncQueue(initialSize int, seq int32) *syncQueue {
	syncQ := &syncQueue{
		queue: make([]*syncItem, 0, initialSize),
		mutex: &sync.Mutex{},
		cond:  &sync.Cond{},
	}
	syncQ.cond.L = syncQ.mutex
	adder := &seqAdder{}
	if seq > 0 {
		adder.value.Store(seq)
	}
	syncQ.adder = adder
	return syncQ
}
