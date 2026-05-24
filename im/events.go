package im

import (
	"sync"
	"sync/atomic"
)

// EventType 事件类型
type EventType int

const (
	EventConnected EventType = iota
	EventDisconnected
	EventConnecting
	EventMessageReceived
	EventMessageSent
	EventError
)

// Event SDK 事件
type Event struct {
	Type EventType
	Data interface{}
}

// EventCallback 事件回调函数
type EventCallback func(Event)

type callbackRegistry struct {
	mu        sync.RWMutex
	listeners map[int64]EventCallback
	counter   atomic.Int64
}

func newCallbackRegistry() *callbackRegistry {
	return &callbackRegistry{
		listeners: make(map[int64]EventCallback),
	}
}

func (r *callbackRegistry) subscribe(cb EventCallback) func() {
	id := r.counter.Add(1)
	r.mu.Lock()
	r.listeners[id] = cb
	r.mu.Unlock()
	return func() {
		r.mu.Lock()
		delete(r.listeners, id)
		r.mu.Unlock()
	}
}

func (r *callbackRegistry) fire(evt Event) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, cb := range r.listeners {
		cb(evt)
	}
}
