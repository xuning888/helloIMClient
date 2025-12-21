package service

import (
	"context"
	"sort"

	"github.com/xuning888/helloIMClient/conf"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

func LastMessage(ctx context.Context, chatId int64, chatType int32) (*sqllite.ChatMessage, error) {
	lastMsg, err := sqllite.GetLastMessage(ctx, chatId)
	if err == nil {
		logger.Infof("LastMessage from DB chatId: %v, chatType: %v", chatId, chatType)
		return lastMsg, nil
	}
	logger.Warnf("LastMessage, getLasetMessage from DB error: %v", err)
	// 查询服务器
	if lastMsg, err2 := http.LastMessage(conf.UserId, chatId, chatType); err2 != nil {
		// 异步更新会话
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

func LastMessageFromRemote(ctx context.Context, chatId int64, chatType int32) (*sqllite.ChatMessage, error) {
	lastMessage, err := http.LastMessage(conf.UserId, chatId, chatType)
	if err != nil {
		return nil, err
	}
	return lastMessage, nil
}

func BatchLastMessage(ctx context.Context, chats []*sqllite.ImChat) map[string]*sqllite.ChatMessage {
	if len(chats) == 0 {
		return make(map[string]*sqllite.ChatMessage)
	}
	result := make(map[string]*sqllite.ChatMessage)
	for _, chat := range chats {
		// 拉取最后一条消息
		lastMsg, err := LastMessage(ctx, chat.ChatId, chat.ChatType)
		if err != nil {
			logger.Errorf("failed to get last message for chatId=%d: %v", chat.ChatId, err)
		}
		result[chat.Key()] = lastMsg
	}
	return result
}

func BatchLastMessageFromRemote(ctx context.Context, chats []*sqllite.ImChat) map[string]*sqllite.ChatMessage {
	if len(chats) == 0 {
		return make(map[string]*sqllite.ChatMessage)
	}
	result := make(map[string]*sqllite.ChatMessage)
	for _, chat := range chats {
		// 拉取最后一条消息
		lastMsg, err := LastMessageFromRemote(ctx, chat.ChatId, chat.ChatType)
		if err != nil {
			logger.Errorf("failed to get last message for chatId=%d: %v", chat.ChatId, err)
		}
		result[chat.Key()] = lastMsg
	}
	return result
}

func PullOfflineMsg(ctx context.Context,
	chatId int64, chatType int32, minServerSeq, maxServerSeq int64) ([]*sqllite.ChatMessage, error) {
	messages, err := sqllite.GetMessagesBySeq(ctx, chatId, minServerSeq, maxServerSeq)
	if err != nil {
		logger.Errorf("PullOfflineMsg.GetMessages chatId: %v, minServerSeq: %d maxServerSeq: %d, error: %v",
			chatId, minServerSeq, maxServerSeq, err)
		if messages, err = http.PullOfflineMsg(conf.UserId, chatId, chatType, minServerSeq, maxServerSeq); err != nil {
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
	if msgs, err2 := http.PullOfflineMsg(conf.UserId, chatId, chatType, minSeq, maxSeq); err2 != nil {
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

func checkMissingMessage(sortedMessage []*sqllite.ChatMessage) (minServerSeq, maxServerSeq int64) {
	if len(sortedMessage) == 0 {
		return
	}
	var hasMissing = false
	minServerSeq, maxServerSeq = sortedMessage[0].ServerSeq, sortedMessage[0].ServerSeq
	for i := 1; i < len(sortedMessage); i++ {
		prevSeq := sortedMessage[i-1].ServerSeq
		curSeq := sortedMessage[i].ServerSeq
		if curSeq-prevSeq > 1 {
			if !hasMissing {
				minServerSeq = prevSeq
				maxServerSeq = curSeq
				hasMissing = true
			} else {
				if prevSeq < minServerSeq {
					minServerSeq = prevSeq
				}
				if curSeq > maxServerSeq {
					maxServerSeq = curSeq
				}
			}
		}
	}
	if !hasMissing {
		minServerSeq, maxServerSeq = 0, 0
	}
	return
}
