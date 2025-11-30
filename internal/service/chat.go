package service

import (
	"context"

	"github.com/xuning888/helloIMClient/conf"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
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
