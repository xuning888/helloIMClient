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
	updateChat()
	chats, err := sqllite.MultiGetChat(ctx)
	if err != nil {
		logger.Errorf("MultiGetChat error: %v", err)
		return nil, err
	}
	return chats, nil
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

func updateChat() {
	chats, err := http.GetAllChat(conf.UserId)
	if err != nil {
		logger.Errorf("GetAllChat error: %v", err)
		return
	}
	if err2 := sqllite.BatchUpdate(context.Background(), chats); err2 != nil {
		logger.Errorf("updateChat error: %v", err)
	}
}
