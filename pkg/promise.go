package pkg

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	ErrorTimeout = errors.New("error promise timeout")
	ErrorClosed  = errors.New("error promise closed")
)

type Promise[T any] struct {
	result  T
	err     error
	done    chan struct{}
	handler []func(v T, err error)
	closed  bool
	lock    sync.Mutex // 保护close和handler
}

func NewPromise[T any]() *Promise[T] {
	return &Promise[T]{
		done: make(chan struct{}),
	}
}

func (p *Promise[T]) Then(h func(v T, err error)) *Promise[T] {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.closed {
		var res = p.result
		var err error = nil
		if p.err != nil {
			err = fmt.Errorf("promise: %v, error: %v", p.err, ErrorClosed)
		} else {
			err = ErrorClosed
		}
		go h(res, err)
		return p
	}
	p.handler = append(p.handler, h)
	return p
}

func (p *Promise[T]) Get(timeout time.Duration) (T, error) {
	p.lock.Lock()
	closed := p.closed
	res := p.result
	err := p.err
	p.lock.Unlock()

	if closed {
		return res, err
	}
	var timeoutCh <-chan time.Time = nil
	if timeout > 0 {
		timeoutCh = time.After(timeout)
	} else {
		timeoutCh = make(<-chan time.Time)
	}
	select {
	case <-p.done:
		p.lock.Lock()
		defer p.lock.Unlock()
		return p.result, p.err
	case <-timeoutCh:
		return *new(T), ErrorTimeout
	}
}

func (p *Promise[T]) Complete(res T, err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.closed {
		return
	}

	p.result = res
	p.err = err
	p.closed = true

	close(p.done)

	p.notify()
}

func (p *Promise[T]) notify() {
	f := func(handler func(v T, err error)) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("promise callback panic: %v", r)
			}
		}()
		handler(p.result, p.err)
	}
	if len(p.handler) == 0 {
		return
	}
	// 异步回调
	go func() {
		for _, h := range p.handler {
			f(h)
		}
	}()
}
