package store

import (
	"context"
	"errors"

	"github.com/xuning888/helloIMClient/im/dbhelper"
	"github.com/xuning888/helloIMClient/im/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"gorm.io/gorm"
)

func (s *StorageImpl) List(ctx context.Context) ([]*dbhelper.ImChat, error) {
	chats, err := dbhelper.MultiGetChat(ctx, s.Uid())
	if err != nil {
		logger.Errorf("MultiGetChat error: %v", err)
		return nil, err
	}
	return chats, nil
}

func (s *StorageImpl) ListFromRemote(ctx context.Context) ([]*dbhelper.ImChat, error) {
	_ = s.SyncFromRemote(ctx)
	return s.List(ctx)
}

func (s *StorageImpl) GetOrCreate(ctx context.Context, chatID int64, chatType int32) (*dbhelper.ImChat, error) {
	uid := s.Uid()
	if chat, err := dbhelper.SelectChat(ctx, uid, chatID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			imChat := dbhelper.NewImChat(uid, chatID, chatType)
			if err2 := dbhelper.InsertChat(ctx, imChat); err2 != nil {
				return nil, err2
			}
			return imChat, nil
		}
		return nil, err
	} else {
		return chat, nil
	}
}

func (s *StorageImpl) SyncFromRemote(ctx context.Context) error {
	chats, err := http.GetAllChat(s.Uid())
	logger.Infof("UpdateChatsFromRemote")
	if err != nil {
		logger.Errorf("GetAllChat error: %v", err)
		return err
	}
	if err2 := dbhelper.BatchUpdate(context.Background(), chats); err2 != nil {
		logger.Errorf("updateChat error: %v", err)
	}
	return nil
}

func (s *StorageImpl) UpdateVersion(ctx context.Context, chatId int64, chatType int32) error {
	chat, err := s.GetOrCreate(ctx, chatId, chatType)
	if err != nil {
		logger.Errorf("UpdateChatVersion.SelectChat error: %v", err)
		return err
	}
	lastMsg, err := s.LastMessage(ctx, chatId, chat.ChatType)
	if err != nil {
		logger.Infof("UpdateChatVersion.LastMessage error: %v", err)
		return err
	}
	if lastMsg.SendTime > chat.UpdateTimestamp {
		chat.UpdateTimestamp = lastMsg.SendTime
		chat.LastReadMsgId = lastMsg.MsgID
		err = dbhelper.BatchUpdate(ctx, append([]*dbhelper.ImChat{}, chat))
		if err != nil {
			logger.Errorf("UpdateChatVersion.BatchUpdate error: %v", err)
		}
		logger.Infof("UpdateChatVersion success: %v", chat)
	}
	return nil
}
