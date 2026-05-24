package im

import (
	"context"

	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/service"
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

func (s *chatStoreImpl) List(ctx context.Context) ([]*sqllite.ImChat, error) {
	return service.GetAllChat(ctx)
}

func (s *chatStoreImpl) ListFromRemote(ctx context.Context) ([]*sqllite.ImChat, error) {
	return service.GetAllChatFromRemote(ctx)
}

func (s *chatStoreImpl) GetOrCreate(ctx context.Context, chatID int64, chatType int32) (*sqllite.ImChat, error) {
	return service.GetOrCreateChat(ctx, chatID, chatType)
}

func (s *chatStoreImpl) SyncFromRemote(ctx context.Context) error {
	service.UpdateChatsFromRemote()
	return nil
}

func (s *chatStoreImpl) UpdateVersion(ctx context.Context, chatID int64, chatType int32) error {
	service.UpdateChatVersion(chatID, chatType)
	return nil
}

// ---- MessageStore ----

type messageStoreImpl struct{}

func (s *messageStoreImpl) Recent(ctx context.Context, chatID int64, chatType int32, limit int) ([]*sqllite.ChatMessage, error) {
	return sqllite.GetRecentMessage(ctx, chatID, chatType, limit)
}

func (s *messageStoreImpl) Save(ctx context.Context, msg *sqllite.ChatMessage) error {
	return sqllite.SaveOrUpdateMessage(ctx, msg)
}

func (s *messageStoreImpl) GetByServerSeq(ctx context.Context, chatID int64, minSeq, maxSeq int64) ([]*sqllite.ChatMessage, error) {
	return sqllite.GetMessagesBySeq(ctx, chatID, minSeq, maxSeq)
}

func (s *messageStoreImpl) LastMessage(ctx context.Context, chatID int64, chatType int32) (*sqllite.ChatMessage, error) {
	return service.LastMessage(ctx, chatID, chatType)
}

func (s *messageStoreImpl) BatchLastMessage(ctx context.Context, chats []*sqllite.ImChat) map[string]*sqllite.ChatMessage {
	return service.BatchLastMessage(ctx, chats)
}

func (s *messageStoreImpl) BatchLastMessageFromRemote(ctx context.Context, chats []*sqllite.ImChat) map[string]*sqllite.ChatMessage {
	return service.BatchLastMessageFromRemote(ctx, chats)
}

func (s *messageStoreImpl) NewCache(chat *sqllite.ImChat) MsgCache {
	return service.NewMsgCache(chat)
}

// ---- UserStore ----

type userStoreImpl struct{}

func (s *userStoreImpl) Get(ctx context.Context, userID int64) (*sqllite.ImUser, error) {
	return service.GetUserById(ctx, userID)
}

func (s *userStoreImpl) Search(ctx context.Context, keyword string) ([]*sqllite.ImUser, error) {
	return sqllite.SearchUser(ctx, keyword)
}

func (s *userStoreImpl) Refresh(ctx context.Context) error {
	service.UpdateUsers()
	return nil
}
