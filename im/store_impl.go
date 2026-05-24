package im

import (
	"context"

	sqllite2 "github.com/xuning888/helloIMClient/im/dal/sqllite"
	service2 "github.com/xuning888/helloIMClient/im/service"
)

func newStore() *Store {
	return &Store{
		Chats:    &chatStoreImpl{},
		Messages: &messageStoreImpl{},
		Users:    &userStoreImpl{},
	}
}

// ---- ChatStore ----

type chatStoreImpl struct{}

func (s *chatStoreImpl) List(ctx context.Context) ([]*sqllite2.ImChat, error) {
	return service2.GetAllChat(ctx)
}

func (s *chatStoreImpl) ListFromRemote(ctx context.Context) ([]*sqllite2.ImChat, error) {
	return service2.GetAllChatFromRemote(ctx)
}

func (s *chatStoreImpl) GetOrCreate(ctx context.Context, chatID int64, chatType int32) (*sqllite2.ImChat, error) {
	return service2.GetOrCreateChat(ctx, chatID, chatType)
}

func (s *chatStoreImpl) SyncFromRemote(ctx context.Context) error {
	service2.UpdateChatsFromRemote()
	return nil
}

func (s *chatStoreImpl) UpdateVersion(ctx context.Context, chatID int64, chatType int32) error {
	service2.UpdateChatVersion(chatID, chatType)
	return nil
}

// ---- MessageStore ----

type messageStoreImpl struct{}

func (s *messageStoreImpl) Recent(ctx context.Context, chatID int64, chatType int32, limit int) ([]*sqllite2.ChatMessage, error) {
	return sqllite2.GetRecentMessage(ctx, chatID, chatType, limit)
}

func (s *messageStoreImpl) Save(ctx context.Context, msg *sqllite2.ChatMessage) error {
	return sqllite2.SaveOrUpdateMessage(ctx, msg)
}

func (s *messageStoreImpl) GetByServerSeq(ctx context.Context, chatID int64, minSeq, maxSeq int64) ([]*sqllite2.ChatMessage, error) {
	return sqllite2.GetMessagesBySeq(ctx, chatID, minSeq, maxSeq)
}

func (s *messageStoreImpl) LastMessage(ctx context.Context, chatID int64, chatType int32) (*sqllite2.ChatMessage, error) {
	return service2.LastMessage(ctx, chatID, chatType)
}

func (s *messageStoreImpl) BatchLastMessage(ctx context.Context, chats []*sqllite2.ImChat) map[string]*sqllite2.ChatMessage {
	return service2.BatchLastMessage(ctx, chats)
}

func (s *messageStoreImpl) BatchLastMessageFromRemote(ctx context.Context, chats []*sqllite2.ImChat) map[string]*sqllite2.ChatMessage {
	return service2.BatchLastMessageFromRemote(ctx, chats)
}

func (s *messageStoreImpl) NewCache(chat *sqllite2.ImChat) MsgCache {
	return service2.NewMsgCache(chat)
}

// ---- UserStore ----

type userStoreImpl struct{}

func (s *userStoreImpl) Get(ctx context.Context, userID int64) (*sqllite2.ImUser, error) {
	return service2.GetUserById(ctx, userID)
}

func (s *userStoreImpl) Search(ctx context.Context, keyword string) ([]*sqllite2.ImUser, error) {
	return sqllite2.SearchUser(ctx, keyword)
}

func (s *userStoreImpl) Refresh(ctx context.Context) error {
	service2.UpdateUsers()
	return nil
}
