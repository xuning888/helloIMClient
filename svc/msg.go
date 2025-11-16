package svc

import (
	"fmt"
	"github.com/xuning888/helloIMClient/pkg"
	"sort"
	"sync"
)

type ChatMessage struct {
	ChatType      int32  `json:"chatType,omitempty"`
	ChatId        int64  `json:"chatId,omitempty"`
	ChatIdStr     string `json:"chatIdStr,omitempty"`
	MsgId         int64  `json:"msgId,omitempty"`
	MsgIdStr      string `json:"msgIdStr,omitempty"`
	MsgFrom       int64  `json:"msgFrom,omitempty"`
	MsgFromStr    string `json:"msgFromStr,omitempty"`
	FromUserType  int32  `json:"fromUserType,omitempty"`
	MsgTo         int64  `json:"msgTo,omitempty"`
	MsgToStr      string `json:"msgToStr,omitempty"`
	ToUserType    int32  `json:"toUserType,omitempty"`
	GroupId       int64  `json:"groupId,omitempty"`
	GroupIdStr    string `json:"groupIdStr,omitempty"`
	MsgSeq        int32  `json:"msgSeq,omitempty"`
	MsgContent    string `json:"msgContent,omitempty"`
	ContentType   int32  `json:"contentType,omitempty"`
	CmdId         int32  `json:"cmdId,omitempty"`
	SendTime      int64  `json:"sendTime,omitempty"`
	ReceiptStatus int32  `json:"receiptStatus,omitempty"`
	ServerSeq     int64  `json:"serverSeq,omitempty"`
}

func (m *ChatMessage) String() string {
	return fmt.Sprintf("%d_%v", m.ServerSeq, m.MsgContent)
}

// OrderedMsg 处理无序到达，有序输出
type OrderedMsg struct {
	hash   map[int64]*ChatMessage
	ptr    int64
	maxSeq int64
}

func NewOrderedMsg() *OrderedMsg {
	return &OrderedMsg{
		hash: make(map[int64]*ChatMessage),
		ptr:  1,
	}
}

func (o *OrderedMsg) Insert(msg *ChatMessage) (msgs []*ChatMessage, minSeq, maxSeq int64) {
	var seq = msg.ServerSeq
	if seq < 1 {
		msgs = []*ChatMessage{}
		minSeq, maxSeq = o.missingSeq()
		return
	}
	// 重复消息
	if seq < o.ptr {
		msgs = []*ChatMessage{}
		minSeq, maxSeq = o.missingSeq()
		return
	}
	o.hash[seq] = msg
	o.maxSeq = pkg.Max(o.maxSeq, seq)
	msgs = make([]*ChatMessage, 0)
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
	messages []*ChatMessage
	dup      map[int64]struct{} // 消息去重
	ordered  *OrderedMsg
}

func (c *MsgSvc) AppendMsg(msg *ChatMessage) (minSeq int64, maxSeq int64) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if _, exists := c.dup[msg.MsgId]; exists {
		return c.ordered.missingSeq()
	}
	msgs, minSeq, maxSeq := c.ordered.Insert(msg)
	if len(msgs) == 0 {
		return minSeq, maxSeq
	}
	c.messages = append(c.messages, msgs...)
	return minSeq, maxSeq
}

func (c *MsgSvc) LastMessage() *ChatMessage {
	c.mux.RLock()
	defer c.mux.RUnlock()
	messages := c.messages
	if len(messages) == 0 {
		return nil
	}
	lastMessage := messages[len(messages)-1]
	return lastMessage
}

func (c *MsgSvc) Range(f func(i int, msg *ChatMessage) bool) {
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
	defer c.mux.RLock()
	if len(c.messages) == 0 {
		return
	}
	sort.Slice(c.messages, func(i, j int) bool {
		return c.messages[i].ServerSeq < c.messages[j].ServerSeq
	})
}

func NewMsgSvc(msgs []*ChatMessage) *MsgSvc {
	msgSvc := &MsgSvc{
		mux:      sync.RWMutex{},
		messages: make([]*ChatMessage, 0),
		ordered:  NewOrderedMsg(),
	}
	if len(msgs) != 0 {
		msgSvc.messages = msgs
	} else {
		msgSvc.messages = make([]*ChatMessage, 0)
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
