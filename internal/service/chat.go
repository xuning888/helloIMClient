package service

import (
	"context"
	"errors"

	"github.com/xuning888/helloIMClient/conf"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"gorm.io/gorm"
)

func GetAllChat(ctx context.Context) ([]*sqllite.ImChat, error) {
	chats, err := sqllite.MultiGetChat(ctx)
	if err != nil {
		logger.Errorf("MultiGetChat error: %v", err)
		return nil, err
	}
	return chats, nil
}

func GetAllChatFromRemote(ctx context.Context) ([]*sqllite.ImChat, error) {
	UpdateChatsFromRemote()
	return GetAllChat(ctx)
}

func GetOrCreateChat(ctx context.Context, chatId int64, chatType int32) (*sqllite.ImChat, error) {
	if chat, err := sqllite.SelectChat(ctx, conf.UserId, chatId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			imChat := sqllite.NewImChat(conf.UserId, chatId, chatType)
			if err2 := sqllite.InsertChat(ctx, imChat); err2 != nil {
				return nil, err2
			}
			return imChat, nil
		}
		return nil, err
	} else {
		return chat, nil
	}
}

func UpdateChatsFromRemote() {
	chats, err := http.GetAllChat(conf.UserId)
	logger.Infof("UpdateChatsFromRemote")
	if err != nil {
		logger.Errorf("GetAllChat error: %v", err)
		return
	}
	if err2 := sqllite.BatchUpdate(context.Background(), chats); err2 != nil {
		logger.Errorf("updateChat error: %v", err)
	}
}

func UpdateChatVersion(chatId int64, chatType int32) {
	ctx := context.Background()
	chat, err := GetOrCreateChat(ctx, chatId, chatType)
	if err != nil {
		logger.Errorf("UpdateChatVersion.SelectChat error: %v", err)
		return
	}
	lastMsg, err := LastMessage(ctx, chatId, chat.ChatType)
	if err != nil {
		logger.Infof("UpdateChatVersion.LastMessage error: %v", err)
		return
	}
	if lastMsg.SendTime > chat.UpdateTimestamp {
		chat.UpdateTimestamp = lastMsg.SendTime
		chat.LastReadMsgId = lastMsg.MsgID
		err = sqllite.BatchUpdate(ctx, append([]*sqllite.ImChat{}, chat))
		if err != nil {
			logger.Errorf("UpdateChatVersion.BatchUpdate error: %v", err)
		}
		logger.Infof("UpdateChatVersion success: %v", chat)
	}
}
