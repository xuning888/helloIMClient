package im

import (
	"sync"
	"sync/atomic"
)

// EventType 事件类型
type EventType int

const (
	// -----长连接相关的事件----

	EventConnected    EventType = iota // 连接简历
	EventDisconnected                  // 连接关闭
	EventConnecting                    // 连接建立中

	// -----消息相关的事件----

	EventMessageInsert   // 消息入库时间
	EventMessageReceived // 收到新消息
	EventMessageSent     // 消息发送成功

	// ---- 异常事件-------

	EventError // 异常事件
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
		go r.fire0(evt, cb)
	}
}

func (r *callbackRegistry) fire0(evt Event, cb EventCallback) {
	defer func() {
		if err := recover(); err != nil {
			println(err.(string))
		}
	}()
	cb(evt)
}
