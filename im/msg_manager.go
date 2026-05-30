package im

import (
	"sync"

	"github.com/xuning888/helloIMClient/im/dbhelper"
)

// NewMsgCallback 新消息回调（已排序批次）
type NewMsgCallback func(msgs []*dbhelper.ChatMessage)

// SendACKCallback 发送确认回调
type SendACKCallback func(msg *dbhelper.ChatMessage)

// MessageManager 消息管理器，负责消息监听器管理和消息推送
type MessageManager struct {
	cli *Client

	mu               sync.RWMutex
	newMsgListeners  map[string]NewMsgCallback
	sendACKListeners map[string]SendACKCallback
}

func newMsgManager(cli *Client) *MessageManager {
	return &MessageManager{
		cli:              cli,
		newMsgListeners:  make(map[string]NewMsgCallback),
		sendACKListeners: make(map[string]SendACKCallback),
	}
}

// AddOnNewMsgListener 注册新消息回调，返回取消函数
func (mm *MessageManager) AddOnNewMsgListener(key string, cb NewMsgCallback) func() {
	mm.mu.Lock()
	mm.newMsgListeners[key] = cb
	mm.mu.Unlock()
	return func() {
		mm.mu.Lock()
		delete(mm.newMsgListeners, key)
		mm.mu.Unlock()
	}
}

// AddOnSendACKListener 注册发送确认回调，返回取消函数
func (mm *MessageManager) AddOnSendACKListener(key string, cb SendACKCallback) func() {
	mm.mu.Lock()
	mm.sendACKListeners[key] = cb
	mm.mu.Unlock()
	return func() {
		mm.mu.Lock()
		delete(mm.sendACKListeners, key)
		mm.mu.Unlock()
	}
}

// pushNewMsg SDK 内部使用，推送新消息至所有监听器
func (mm *MessageManager) pushNewMsg(msgs []*dbhelper.ChatMessage) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	for _, cb := range mm.newMsgListeners {
		go pushNewMsg0(cb, msgs)
	}
}

// pushSendACK SDK 内部使用，推送发送确认至所有监听器
func (mm *MessageManager) pushSendACK(msg *dbhelper.ChatMessage) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	for _, cb := range mm.sendACKListeners {
		go pushSendACK0(cb, msg)
	}
}

func pushNewMsg0(cb NewMsgCallback, msgs []*dbhelper.ChatMessage) {
	defer func() {
		if err := recover(); err != nil {
			println(err.(string))
		}
	}()
	cb(msgs)
}

func pushSendACK0(cb SendACKCallback, msg *dbhelper.ChatMessage) {
	defer func() {
		if err := recover(); err != nil {
			println(err.(string))
		}
	}()
	cb(msg)
}
