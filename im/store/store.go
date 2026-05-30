package store

import (
	"context"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/xuning888/helloIMClient/im/dbhelper"
)

type Store struct {
	Users    UserStore
	Messages MessageStore
	Chats    ChatStore
}

func NewStore(uid int64) *Store {
	store := &Store{}
	storeImpl := newStoreImpl(store, uid)
	store.Chats = storeImpl
	store.Messages = storeImpl
	store.Users = storeImpl
	return store
}

var _ MessageStore = &StorageImpl{}
var _ ChatStore = &StorageImpl{}
var _ UserStore = &StorageImpl{}

type StorageImpl struct {
	store *Store
	uid   int64

	// cache 用户信息内存缓存
	cache *lru.Cache[int64, *dbhelper.ImUser]

	// msgCache
	msgss sync.Map
}

func newStoreImpl(adapter *Store, uid int64) *StorageImpl {
	store := &StorageImpl{uid: uid, store: adapter}
	cache, err := lru.New[int64, *dbhelper.ImUser](500)
	if err != nil {
		panic("failed to create LRU cache: " + err.Error())
	}
	store.cache = cache
	return store
}

func (s *StorageImpl) Uid() int64 {
	return s.uid
}

// MsgCache 消息缓存接口
type MsgCache interface {
	GetChat() *dbhelper.ImChat
	GetMessages() []*dbhelper.ChatMessage
	Refresh()
	Destroy()
}

// ChatStore 会话存储接口
type ChatStore interface {
	List(ctx context.Context) ([]*dbhelper.ImChat, error)
	ListFromRemote(ctx context.Context) ([]*dbhelper.ImChat, error)
	GetOrCreate(ctx context.Context, chatID int64, chatType int32) (*dbhelper.ImChat, error)
	SyncFromRemote(ctx context.Context) error
	UpdateVersion(ctx context.Context, chatID int64, chatType int32) error
}

// MessageStore 消息存储接口
type MessageStore interface {
	Recent(ctx context.Context, chatID int64, chatType int32, limit int) ([]*dbhelper.ChatMessage, error)
	Save(ctx context.Context, msg *dbhelper.ChatMessage) error
	GetByServerSeq(ctx context.Context, chatID int64, minSeq, maxSeq int64) ([]*dbhelper.ChatMessage, error)
	LastMessage(ctx context.Context, chatID int64, chatType int32) (*dbhelper.ChatMessage, error)
	BatchLastMessage(ctx context.Context, chats []*dbhelper.ImChat) map[string]*dbhelper.ChatMessage
	LastMessageFromRemote(ctx context.Context, chatId int64, chatType int32) (*dbhelper.ChatMessage, error)
	BatchLastMessageFromRemote(ctx context.Context, chats []*dbhelper.ImChat) map[string]*dbhelper.ChatMessage
	GetLatestOfflineMessages(ctx context.Context, chatId int64, chatType int32, size int32) ([]*dbhelper.ChatMessage, error)
	PullOfflineMsg(ctx context.Context, chatId int64, chatType int32, minServerSeq, maxServerSeq int64) ([]*dbhelper.ChatMessage, error)
	NewCache(chat *dbhelper.ImChat) MsgCache
	DestroyCache(chat *dbhelper.ImChat)
	GetHistory(ctx context.Context, chatID int64, chatType int32, beforeSeq int64, limit int) ([]*dbhelper.ChatMessage, error)
	ProcessIncomingMessage(msg *dbhelper.ChatMessage) []*dbhelper.ChatMessage
}

// UserStore 用户存储接口
type UserStore interface {
	Get(ctx context.Context, userID int64) (*dbhelper.ImUser, error)
	Search(ctx context.Context, keyword string) ([]*dbhelper.ImUser, error)
	Refresh(ctx context.Context) error
}
