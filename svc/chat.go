package svc

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type ChatType int

const (
	C2C ChatType = 1
	C2G ChatType = 2
)

type ChatInfo struct {
	Id                 string `json:"id"`
	UserID             int64  `json:"userId"`
	ChatID             int64  `json:"chatId"`
	ChatType           int32  `json:"chatType"`
	ChatTop            bool   `json:"chatTop"`
	ChatMute           bool   `json:"chatMute"`
	ChatDel            bool   `json:"chatDel"`
	UpdateTimestamp    int64  `json:"updateTimestamp"`
	DelTimestamp       int64  `json:"delTimestamp"`
	LastReadMsgID      int64  `json:"lastReadMsgId"`
	SubStatus          int32  `json:"subStatus"`
	JoinGroupTimestamp int64  `json:"joinGroupTimestamp"`
}

type Chat struct {
	Id                       int64        // 会话ID
	Type                     ChatType     // 会话类型
	Msgs                     *MsgSvc      // 会话中的消息
	LastChatMessage          *ChatMessage // 会话最后一条消息
	LastChatMessageTimestamp int64        // 会话最后一条消息的时间戳
	ChatName                 string       // 会话名称
	Timestamp                int64        // 会话的更新时间
	UnReadNum                int          // 会话未读数
}

func (c *Chat) Index() string {
	return fmt.Sprintf("%v_%d", c.Type, c.Id)
}

func NewChat(chatId int64, chatType ChatType, chatName string, msgs []*ChatMessage) *Chat {
	chat := &Chat{
		Id:       chatId,
		Type:     chatType,
		ChatName: chatName,
	}
	if len(msgs) != 0 {
		chat.Msgs = NewMsgSvc(msgs)
		chat.LastChatMessage = chat.Msgs.LastMessage()
		chat.LastChatMessageTimestamp = chat.Msgs.LastMessage().SendTime
		chat.UnReadNum = 0
	} else {
		chat.Msgs = NewMsgSvc(nil)
		chat.LastChatMessage = nil
		chat.LastChatMessageTimestamp = time.Now().UnixMilli()
		chat.UnReadNum = 0
	}
	return chat
}

type ChatSvc struct {
	mux       sync.RWMutex
	chats     []*Chat        // 会话列表
	chatIndex map[string]int // 会话索引
}

func (cc *ChatSvc) AddChat(chat *Chat) {
	cc.mux.Lock()
	defer cc.mux.Unlock()

	// 会话已经存在
	if _, exists := cc.chatIndex[chat.Index()]; exists {
		return
	}
	// 添加会话
	cc.chats = append(cc.chats, chat)
	cc.sortChats()
}

// SortChats 排序会话，并重建会话索引
func (cc *ChatSvc) sortChats() {
	sort.Slice(cc.chats, func(i, j int) bool {
		return cc.chats[i].LastChatMessageTimestamp > cc.chats[j].LastChatMessageTimestamp
	})
	cc.rebuildIndex()
}

func (cc *ChatSvc) rebuildIndex() {
	cc.chatIndex = make(map[string]int, len(cc.chats))
	for i, chat := range cc.chats {
		cc.chatIndex[chat.Index()] = i
	}
}

func (cc *ChatSvc) GetChat(chatId int64, chatType ChatType) *Chat {
	cc.mux.RLock()
	defer cc.mux.RUnlock()
	index := fmt.Sprintf("%v_%d", chatType, chatId)
	if i, exists := cc.chatIndex[index]; exists {
		return cc.chats[i]
	}
	return nil
}

func (cc *ChatSvc) GetChatByIndex(idx int) *Chat {
	cc.mux.RLock()
	defer cc.mux.RUnlock()
	return cc.chats[idx]
}

func (cc *ChatSvc) Update() {
	cc.mux.Lock()
	defer cc.mux.Unlock()
	cc.sortChats()
}

func (cc *ChatSvc) ChatRange(f func(i int, chat *Chat) bool) {
	cc.mux.RLock()
	defer cc.mux.RUnlock()
	if len(cc.chats) == 0 {
		return
	}
	for i, chat := range cc.chats {
		if !f(i, chat) {
			break
		}
	}
}

func (cc *ChatSvc) ChatLen() int {
	cc.mux.RLock()
	defer cc.mux.RUnlock()
	return len(cc.chats)
}

func newChatSvc(chats []*Chat) *ChatSvc {
	chatSvc := &ChatSvc{
		mux:       sync.RWMutex{},
		chatIndex: make(map[string]int),
	}
	if len(chats) != 0 {
		chatSvc.chats = chats
	}
	chatSvc.Update()
	return chatSvc
}
