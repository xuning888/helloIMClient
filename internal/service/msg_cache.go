package service

import (
	"context"
	"sort"
	"sync"

	"github.com/xuning888/helloIMClient/conf"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

const maxCachedMessages = 30 // 最大缓存消息数

type MsgCache struct {
	mux     sync.RWMutex
	chat    *sqllite.ImChat
	message []*sqllite.ChatMessage
	dup     map[int64]struct{}
}

func NewMsgCache(chat *sqllite.ImChat) *MsgCache {
	cache := &MsgCache{
		mux:     sync.RWMutex{},
		chat:    chat,
		message: make([]*sqllite.ChatMessage, 0, maxCachedMessages),
		dup:     map[int64]struct{}{},
	}
	cache.loadInitialMessages()
	return cache
}

func (m *MsgCache) loadInitialMessages() {
	m.mux.Lock()
	defer m.mux.Unlock()
	ctx := context.Background()
	chatId, chatType := m.chat.ChatId, m.chat.ChatType
	// 先从本地查询近期的消息, 如果查询不到就从远程拉
	messages, err := sqllite.GetRecentMessage(ctx, chatId, chatType, maxCachedMessages)
	if err != nil || len(messages) == 0 {
		if err != nil {
			logger.Errorf("GetRecentMessage chatId: %v, chatType: %v, error: %v", chatId, chatType, err)
		}
		messages, err = http.GetLatestOfflineMessages(conf.UserId, chatId, chatType, maxCachedMessages)
		if err != nil {
			logger.Errorf("GetLatestOfflineMessages chatId: %v, chatType: %v, error: %v", chatId, chatType, err)
		}
	}
	if len(messages) > 0 {
		// 保存到数据库并更新缓存
		sortMessages(messages)
		m.addMessages(messages)
		// 最后一条消息和chat上的最后一条消息做对比, 如果存在差异就获取最后一条消息
		lstMsg := messages[len(messages)-1]
		if lstMsg.MsgID != m.chat.LastReadMsgId {
			if lastMessage, err := LastMessageFromRemote(ctx, chatId, chatType); err != nil {
				logger.Errorf("loadInitialMessages.LastMessageFromRemote chatId: %d, chatType: %v  error: %v", chatId, chatType, err)
			} else {
				m.addMessages([]*sqllite.ChatMessage{lastMessage})
			}
		}
	}
	m.checkMissingMessageAndSort(ctx)
}

func (m *MsgCache) UpdateMessage(msgs []*sqllite.ChatMessage) {
	m.mux.Lock()
	defer m.mux.Unlock()
	ctx := context.Background()
	if len(msgs) == 0 {
		lastMessage, err := LastMessage(ctx, m.chat.ChatId, m.chat.ChatType)
		if err != nil {
			logger.Errorf("UpdateMessage.GetLastMessage error: %v", err)
		} else {
			m.addMessages([]*sqllite.ChatMessage{lastMessage})
		}
		m.checkMissingMessageAndSort(ctx)
		return
	}
	m.addMessages(msgs)
	m.checkMissingMessageAndSort(ctx)
}

func (m *MsgCache) checkMissingMessageAndSort(ctx context.Context) {
	m.sortMessage()
	minSeq, maxSeq := checkMissingMessage(m.message)
	if minSeq == maxSeq {
		return
	}
	logger.Infof("checkMessage missing, minSeq: %v, maxSeq: %v", minSeq, maxSeq)
	msgs, err := PullOfflineMsg(ctx, m.chat.ChatId, m.chat.ChatType, minSeq, maxSeq)
	if err != nil {
		logger.Errorf("checkMessage error: %v", err)
		return
	}
	if len(msgs) == 0 {
		return
	}
	m.addMessages(msgs)
	m.sortMessage()
}

func (m *MsgCache) sortMessage() {
	sortMessages(m.message)
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

func (m *MsgCache) addMessages(messages []*sqllite.ChatMessage) {
	if len(messages) == 0 {
		return
	}
	ctx := context.Background()
	for _, msg := range messages {
		if _, exists := m.dup[msg.MsgID]; exists {
			continue
		}
		m.dup[msg.MsgID] = struct{}{}
		if err := sqllite.SaveOrUpdateMessage(ctx, msg); err != nil {
			logger.Errorf("SaveOrUpdateMessage error: %v", err)
		}
		m.message = append(m.message, msg)
	}
}

func sortMessages(messages []*sqllite.ChatMessage) {
	if len(messages) == 0 {
		return
	}
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ServerSeq < messages[j].ServerSeq
	})
}
