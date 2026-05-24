package im

import (
	"context"

	sqllite2 "github.com/xuning888/helloIMClient/im/dal/sqllite"
)

// Store 存储管理器
type Store struct {
	Chats    ChatStore
	Messages MessageStore
	Users    UserStore
}

// MsgCache 消息缓存接口
type MsgCache interface {
	GetChat() *sqllite2.ImChat
	GetMessages() []*sqllite2.ChatMessage
	UpdateMessage(msgs []*sqllite2.ChatMessage)
}

// ChatStore 会话存储接口
type ChatStore interface {
	List(ctx context.Context) ([]*sqllite2.ImChat, error)
	ListFromRemote(ctx context.Context) ([]*sqllite2.ImChat, error)
	GetOrCreate(ctx context.Context, chatID int64, chatType int32) (*sqllite2.ImChat, error)
	SyncFromRemote(ctx context.Context) error
	UpdateVersion(ctx context.Context, chatID int64, chatType int32) error
}

// MessageStore 消息存储接口
type MessageStore interface {
	Recent(ctx context.Context, chatID int64, chatType int32, limit int) ([]*sqllite2.ChatMessage, error)
	Save(ctx context.Context, msg *sqllite2.ChatMessage) error
	GetByServerSeq(ctx context.Context, chatID int64, minSeq, maxSeq int64) ([]*sqllite2.ChatMessage, error)
	LastMessage(ctx context.Context, chatID int64, chatType int32) (*sqllite2.ChatMessage, error)
	BatchLastMessage(ctx context.Context, chats []*sqllite2.ImChat) map[string]*sqllite2.ChatMessage
	BatchLastMessageFromRemote(ctx context.Context, chats []*sqllite2.ImChat) map[string]*sqllite2.ChatMessage
	NewCache(chat *sqllite2.ImChat) MsgCache
}

// UserStore 用户存储接口
type UserStore interface {
	Get(ctx context.Context, userID int64) (*sqllite2.ImUser, error)
	Search(ctx context.Context, keyword string) ([]*sqllite2.ImUser, error)
	Refresh(ctx context.Context) error
}
