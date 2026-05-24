package im

import (
	"context"

	sqllite "github.com/xuning888/helloIMClient/im/dal/sqllite"
)

// Store 存储管理器
type Store struct {
	Chats    ChatStore
	Messages MessageStore
	Users    UserStore
}

// MsgCache 消息缓存接口
type MsgCache interface {
	GetChat() *sqllite.ImChat
	GetMessages() []*sqllite.ChatMessage
	UpdateMessage(msgs []*sqllite.ChatMessage)
}

// ChatStore 会话存储接口
type ChatStore interface {
	List(ctx context.Context) ([]*sqllite.ImChat, error)
	ListFromRemote(ctx context.Context) ([]*sqllite.ImChat, error)
	GetOrCreate(ctx context.Context, chatID int64, chatType int32) (*sqllite.ImChat, error)
	SyncFromRemote(ctx context.Context) error
	UpdateVersion(ctx context.Context, chatID int64, chatType int32) error
}

// MessageStore 消息存储接口
type MessageStore interface {
	Recent(ctx context.Context, chatID int64, chatType int32, limit int) ([]*sqllite.ChatMessage, error)
	Save(ctx context.Context, msg *sqllite.ChatMessage) error
	GetByServerSeq(ctx context.Context, chatID int64, minSeq, maxSeq int64) ([]*sqllite.ChatMessage, error)
	LastMessage(ctx context.Context, chatID int64, chatType int32) (*sqllite.ChatMessage, error)
	BatchLastMessage(ctx context.Context, chats []*sqllite.ImChat) map[string]*sqllite.ChatMessage
	BatchLastMessageFromRemote(ctx context.Context, chats []*sqllite.ImChat) map[string]*sqllite.ChatMessage
	NewCache(chat *sqllite.ImChat) MsgCache
}

// UserStore 用户存储接口
type UserStore interface {
	Get(ctx context.Context, userID int64) (*sqllite.ImUser, error)
	Search(ctx context.Context, keyword string) ([]*sqllite.ImUser, error)
	Refresh(ctx context.Context) error
}
