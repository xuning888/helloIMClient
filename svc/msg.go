package svc

import (
	"sort"
	"sync"

	"github.com/xuning888/helloIMClient/internal/model"
	"github.com/xuning888/helloIMClient/pkg"
)

// OrderedMsg 处理无序到达，有序输出
type OrderedMsg struct {
	hash   map[int64]*model.ChatMessage
	ptr    int64
	maxSeq int64
}

func NewOrderedMsg() *OrderedMsg {
	return &OrderedMsg{
		hash: make(map[int64]*model.ChatMessage),
		ptr:  1,
	}
}

func (o *OrderedMsg) Insert(msg *model.ChatMessage) (msgs []*model.ChatMessage, minSeq, maxSeq int64) {
	var seq = msg.ServerSeq
	if seq < 1 {
		msgs = []*model.ChatMessage{}
		minSeq, maxSeq = o.missingSeq()
		return
	}
	// 重复消息
	if seq < o.ptr {
		msgs = []*model.ChatMessage{}
		minSeq, maxSeq = o.missingSeq()
		return
	}
	o.hash[seq] = msg
	o.maxSeq = pkg.Max(o.maxSeq, seq)
	msgs = make([]*model.ChatMessage, 0)
	if o.ptr == seq {
		delKey := make([]int64, 0)
		for {
			if v, ok := o.hash[o.ptr]; ok {
				msgs = append(msgs, v)
				delKey = append(delKey, o.ptr)
				o.ptr++
			} else {
				break
			}
		}
		for _, k := range delKey {
			delete(o.hash, k)
		}
	}
	minSeq, maxSeq = o.missingSeq()
	return
}

func (o *OrderedMsg) missingSeq() (minSeq, maxSeq int64) {
	return o.ptr, o.maxSeq
}

type MsgSvc struct {
	mux      sync.RWMutex
	messages []*model.ChatMessage
	dup      map[int64]struct{} // 消息去重
	ordered  *OrderedMsg
}

func (c *MsgSvc) AppendMsg(msg *model.ChatMessage) (minSeq int64, maxSeq int64) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if _, exists := c.dup[msg.MsgId]; exists {
		return c.ordered.missingSeq()
	}
	c.messages = append(c.messages, msg)
	return minSeq, maxSeq
}

func (c *MsgSvc) LastMessage() *model.ChatMessage {
	c.mux.RLock()
	defer c.mux.RUnlock()
	messages := c.messages
	if len(messages) == 0 {
		return nil
	}
	lastMessage := messages[len(messages)-1]
	return lastMessage
}

func (c *MsgSvc) Range(f func(i int, msg *model.ChatMessage) bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	for i, msg := range c.messages {
		if !f(i, msg) {
			break
		}
	}
}

func (c *MsgSvc) Sort() {
	c.mux.Lock()
	defer c.mux.Unlock()
	if len(c.messages) == 0 {
		return
	}
	sort.Slice(c.messages, func(i, j int) bool {
		return c.messages[i].ServerSeq < c.messages[j].ServerSeq
	})
}

func NewMsgSvc(msgs []*model.ChatMessage) *MsgSvc {
	msgSvc := &MsgSvc{
		mux:      sync.RWMutex{},
		messages: make([]*model.ChatMessage, 0),
		ordered:  NewOrderedMsg(),
	}
	if len(msgs) != 0 {
		msgSvc.messages = msgs
	} else {
		msgSvc.messages = make([]*model.ChatMessage, 0)
	}
	msgSvc.Sort()
	var initPtr, maxSeq int64
	if len(msgSvc.messages) == 0 {
		initPtr = 1
		maxSeq = 0
	} else {
		message := msgSvc.LastMessage()
		initPtr = message.ServerSeq + 1
		maxSeq = message.ServerSeq
	}
	msgSvc.ordered.ptr = initPtr
	msgSvc.ordered.maxSeq = maxSeq
	return msgSvc
}
