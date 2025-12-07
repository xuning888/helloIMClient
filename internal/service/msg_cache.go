package service

import (
	"context"
	"sync"

	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

const maxCachedMessages = 100 // 最大缓存消息数

type MsgCache struct {
	mux     sync.RWMutex
	chat    *sqllite.ImChat
	message []*sqllite.ChatMessage
}

func NewMsgCache(chat *sqllite.ImChat) *MsgCache {
	cache := &MsgCache{
		mux:     sync.RWMutex{},
		chat:    chat,
		message: make([]*sqllite.ChatMessage, 0, maxCachedMessages), // 预分配容量
	}

	cache.loadInitialMessages()
	return cache
}

func (m *MsgCache) loadInitialMessages() {
	ctx := context.Background()
	messages, err := GetLatestOfflineMessages(ctx, m.chat.ChatId, m.chat.ChatType)
	if err != nil {
		logger.Errorf("GetLatestOfflineMessages error: %v", err)
	}

	m.mux.Lock()
	defer m.mux.Unlock()

	// 保存到数据库并更新缓存
	for _, msg := range messages {
		if err = sqllite.SaveOrUpdateMessage(ctx, msg); err != nil {
			logger.Errorf("SaveOrUpdateMessage error: %v", err)
		}
	}

	if dbMessages, err2 := sqllite.GetRecentMessages(ctx, m.chat.ChatId, maxCachedMessages); err2 != nil {
		logger.Errorf("GetRecentMessages error: %v", err2)
	} else {
		messages = dbMessages
	}
	m.message = m.truncateMessages(messages)
}

func (m *MsgCache) UpdateMessage() {
	m.mux.Lock()
	defer m.mux.Unlock()

	ctx := context.Background()
	var newMessages []*sqllite.ChatMessage
	var err error

	if len(m.message) == 0 {
		// 如果没有缓存消息，加载最新消息
		newMessages, err = GetLatestOfflineMessages(ctx, m.chat.ChatId, m.chat.ChatType)
	} else {
		// 基于最后一条消息获取更新
		lastMsg := m.message[len(m.message)-1]
		newMessages, err = sqllite.GetMessagesWithOffset(ctx, m.chat.ChatId, lastMsg.MsgID, maxCachedMessages)
	}

	if err != nil {
		logger.Errorf("Get messages error: %v", err)
		return
	}

	for _, msg := range newMessages {
		if err = sqllite.SaveOrUpdateMessage(ctx, msg); err != nil {
			logger.Errorf("SaveOrUpdateMessage error: %v", err)
		}
	}
	allMessages := append(m.message, newMessages...)
	m.message = m.truncateMessages(allMessages)
}

// 截断消息，只保留最新的 maxCachedMessages 条
func (m *MsgCache) truncateMessages(messages []*sqllite.ChatMessage) []*sqllite.ChatMessage {
	if len(messages) <= maxCachedMessages {
		return messages
	}
	// 返回最新的 maxCachedMessages 条消息
	startIndex := len(messages) - maxCachedMessages
	return messages[startIndex:]
}

func (m *MsgCache) GetMessages() []*sqllite.ChatMessage {
	m.mux.RLock()
	defer m.mux.RUnlock()

	messages := make([]*sqllite.ChatMessage, len(m.message))
	copy(messages, m.message)
	return messages
}

func (m *MsgCache) GetChat() *sqllite.ImChat {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.chat
}
