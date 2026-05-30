package store

import (
	"context"
	"sort"

	"github.com/xuning888/helloIMClient/im/dbhelper"
	"github.com/xuning888/helloIMClient/im/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

func (s *StorageImpl) Recent(ctx context.Context, chatID int64, chatType int32, limit int) ([]*dbhelper.ChatMessage, error) {
	return dbhelper.GetRecentMessage(ctx, chatID, chatType, limit)
}

func (s *StorageImpl) Save(ctx context.Context, msg *dbhelper.ChatMessage) error {
	return dbhelper.SaveOrUpdateMessage(ctx, s.Uid(), msg)
}

func (s *StorageImpl) GetByServerSeq(ctx context.Context, chatID int64, minSeq, maxSeq int64) ([]*dbhelper.ChatMessage, error) {
	return dbhelper.GetMessagesBySeq(ctx, chatID, minSeq, maxSeq)
}

func (s *StorageImpl) LastMessage(ctx context.Context, chatId int64, chatType int32) (*dbhelper.ChatMessage, error) {
	lastMsg, err := dbhelper.GetLastMessage(ctx, chatId)
	if err == nil {
		logger.Infof("LastMessage from DB chatId: %v, chatType: %v", chatId, chatType)
		return lastMsg, nil
	}
	logger.Warnf("LastMessage, getLasetMessage from DB error: %v", err)
	// 查询服务器
	if lastMsg, err2 := http.LastMessage(s.Uid(), chatId, chatType); err2 != nil {
		// 异步更新会话
		logger.Errorf("http.LastMessage error: %v", err)
		return nil, err2
	} else {
		// 保存消息到数据库
		if err3 := dbhelper.SaveOrUpdateMessage(ctx, 0, lastMsg); err3 != nil {
			logger.Errorf("SaveOrUpdateMessage error: %v", err3)
		}
		return lastMsg, nil
	}
}

func (s *StorageImpl) BatchLastMessage(ctx context.Context, chats []*dbhelper.ImChat) map[string]*dbhelper.ChatMessage {
	if len(chats) == 0 {
		return make(map[string]*dbhelper.ChatMessage)
	}
	result := make(map[string]*dbhelper.ChatMessage)
	for _, chat := range chats {
		// 拉取最后一条消息
		lastMsg, err := s.LastMessage(ctx, chat.ChatId, chat.ChatType)
		if err != nil {
			logger.Errorf("failed to get last message for chatId=%d: %v", chat.ChatId, err)
		}
		result[chat.Key()] = lastMsg
	}
	return result
}

func (s *StorageImpl) BatchLastMessageFromRemote(ctx context.Context, chats []*dbhelper.ImChat) map[string]*dbhelper.ChatMessage {
	if len(chats) == 0 {
		return make(map[string]*dbhelper.ChatMessage)
	}
	result := make(map[string]*dbhelper.ChatMessage)
	for _, chat := range chats {
		// 拉取最后一条消息
		lastMsg, err := s.LastMessageFromRemote(ctx, chat.ChatId, chat.ChatType)
		if err != nil {
			logger.Errorf("failed to get last message for chatId=%d: %v", chat.ChatId, err)
		}
		result[chat.Key()] = lastMsg
	}
	return result
}

func (s *StorageImpl) GetLatestOfflineMessages(ctx context.Context, chatId int64, chatType int32, size int32) ([]*dbhelper.ChatMessage, error) {
	return http.GetLatestOfflineMessages(s.Uid(), chatId, chatType, size)
}

func (s *StorageImpl) LastMessageFromRemote(ctx context.Context, chatId int64, chatType int32) (*dbhelper.ChatMessage, error) {
	return http.LastMessage(s.Uid(), chatId, chatType)
}

func (s *StorageImpl) PullOfflineMsg(ctx context.Context, chatId int64, chatType int32, minServerSeq, maxServerSeq int64) ([]*dbhelper.ChatMessage, error) {
	userId := s.Uid()
	messages, err := dbhelper.GetMessagesBySeq(ctx, chatId, minServerSeq, maxServerSeq)
	if err != nil || len(messages) == 0 {
		logger.Errorf("PullOfflineMsg.GetMessages chatId: %v, minServerSeq: %d maxServerSeq: %d, error: %v",
			chatId, minServerSeq, maxServerSeq, err)
		if messages, err = http.PullOfflineMsg(userId, chatId, chatType, minServerSeq, maxServerSeq); err != nil {
			logger.Errorf("PullOfflineMsg.http chatId: %v, chatType: %d, minServerSeq: %v, maxServerSeq: %d, error: %v",
				chatId, chatType, minServerSeq, maxServerSeq, err)
			return nil, err
		} else {
			return messages, nil
		}
	}
	minSeq, maxSeq := checkMissingMessage(messages)
	if minSeq == maxSeq {
		return messages, nil
	}
	logger.Infof("PullOfflineMsg missing minSeq: %v, maxSeq: %v", minSeq, maxSeq)
	if msgs, err2 := http.PullOfflineMsg(userId, chatId, chatType, minSeq, maxSeq); err2 != nil {
		logger.Errorf("PullOfflineMsg.http chatId: %v, chatType: %d, minSeq: %v, maxSeq: %d, error: %v",
			chatId, chatType, minServerSeq, maxServerSeq, err)
		return messages, nil
	} else {
		for _, msg := range msgs {
			messages = append(messages, msg)
		}
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].ServerSeq < messages[j].ServerSeq
		})
		return messages, nil
	}
}

func (s *StorageImpl) NewCache(chat *dbhelper.ImChat) MsgCache {
	key := chat.Key()
	value, ok := s.msgss.Load(key)
	if ok {
		return value.(*MsgCacheImpl)
	}
	cache := NewMsgCache(s.store, chat)
	s.msgss.Store(key, cache)
	return cache
}

func (s *StorageImpl) DestroyCache(chat *dbhelper.ImChat) {
	key := chat.Key()
	s.msgss.LoadAndDelete(key)
}
