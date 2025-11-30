package service

import (
	"context"
	"sync"

	"github.com/xuning888/helloIMClient/conf"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

func LastMessage(ctx context.Context, chatId, msgId int64, chatType int32) (*sqllite.ChatMessage, error) {
	lastMsg, err := sqllite.GetMessage(ctx, chatId, msgId)
	if err == nil {
		return lastMsg, nil
	}
	// 查询服务器
	if lastMsg, err2 := http.LastMessage(conf.UserId, chatId, chatType); err2 != nil {
		// 异步更新会话
		go updateChat()
		logger.Errorf("http.LastMessage error: %v", err)
		return nil, err2
	} else {
		// 保存消息到数据库
		if err3 := sqllite.SaveOrUpdateMessage(ctx, lastMsg); err3 != nil {
			logger.Errorf("SaveOrUpdateMessage error: %v", err3)
		}
		return lastMsg, nil
	}
}

func BatchLastMessage(ctx context.Context, chats []*sqllite.ImChat) map[string]*sqllite.ChatMessage {
	if len(chats) == 0 {
		return make(map[string]*sqllite.ChatMessage)
	}
	var wg sync.WaitGroup
	cache := sync.Map{}
	sem := make(chan struct{}, 5)
	for _, chat := range chats {
		wg.Add(1)
		sem <- struct{}{}
		go func(c *sqllite.ImChat) {
			defer func() {
				wg.Done()
				<-sem
			}()
			// 拉取最后一条消息
			lastMsg, err := LastMessage(ctx, c.ChatId, c.LastReadMsgId, c.ChatType)
			if err != nil {
				logger.Errorf("failed to get last message for chatId=%d: %v", c.ChatId, err)
				return
			}
			// 保存到缓存
			cache.Store(c.Key(), lastMsg)
		}(chat)
	}
	wg.Wait()
	result := make(map[string]*sqllite.ChatMessage)
	cache.Range(func(key, value any) bool {
		k := key.(string)
		msg := value.(*sqllite.ChatMessage)
		result[k] = msg
		return true
	})
	return result
}

func PullOfflineMsg(ctx context.Context,
	chatId int64, chatType int32, minServerSeq, maxServerSeq int64) ([]*sqllite.ChatMessage, error) {
	msgs, err := http.PullOfflineMsg(conf.UserId, chatId, chatType, minServerSeq, maxServerSeq)
	return msgs, err
}

func GetLatestOfflineMessages(ctx context.Context, chatId int64, chatType int32) ([]*sqllite.ChatMessage, error) {
	messages, err := http.GetLatestOfflineMessages(conf.UserId, chatId, chatType, 30)
	return messages, err
}
