package svc

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/xuning888/helloIMClient/internal/model"
)

type ChatType int

const (
	C2C ChatType = 1
	C2G ChatType = 2
)

type Chat struct {
	UserId                   int64
	ChatId                   int64
	ChatType                 int32
	ChatTop                  bool
	ChatMute                 bool
	ChatDel                  bool
	UpdateTimestamp          int64
	DelTimestamp             int64
	LastReadMsgID            int64
	SubStatus                int32
	JoinGroupTimestamp       int64
	Msgs                     *MsgSvc
	LastChatMessage          *model.ChatMessage
	LastChatMessageTimestamp int64
	ChatName                 string
	Timestamp                int64
	UnReadNum                int
}

func (c *Chat) Index() string {
	return fmt.Sprintf("%d_%d", c.ChatType, c.ChatId)
}

func NewChat(chatId int64, chatType int32, chatName string, msgs []*model.ChatMessage) *Chat {
	chat := &Chat{
		ChatId:   chatId,
		ChatType: chatType,
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

func (cc *ChatSvc) GetChat(chatId int64, chatType int32) *Chat {
	cc.mux.RLock()
	defer cc.mux.RUnlock()
	index := fmt.Sprintf("%d_%d", chatType, chatId)
	if i, exists := cc.chatIndex[index]; exists {
		chat := cc.chats[i]
		if chat != nil {
			return chat
		}
		return nil
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

func NewChatSvc(chats []*Chat) *ChatSvc {
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
