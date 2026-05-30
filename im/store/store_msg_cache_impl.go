package store

import (
	"context"
	"sort"
	"sync"

	"github.com/xuning888/helloIMClient/im/dbhelper"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

const maxCachedMessages = 30 // 最大缓存消息数

type MsgCacheImpl struct {
	store   *Store
	mux     sync.RWMutex
	chat    *dbhelper.ImChat
	message []*dbhelper.ChatMessage
	dup     map[int64]struct{}
}

func NewMsgCache(adaptor *Store, chat *dbhelper.ImChat) *MsgCacheImpl {
	cache := &MsgCacheImpl{
		store:   adaptor,
		mux:     sync.RWMutex{},
		chat:    chat,
		message: make([]*dbhelper.ChatMessage, 0, maxCachedMessages),
		dup:     map[int64]struct{}{},
	}
	cache.Refresh()
	return cache
}

func (m *MsgCacheImpl) Refresh() {
	m.clean()
	m.loadInitialMessages()
}

func (m *MsgCacheImpl) loadInitialMessages() {
	m.mux.Lock()
	defer m.mux.Unlock()
	ctx := context.Background()
	chatId, chatType := m.chat.ChatId, m.chat.ChatType
	// 先从本地查询近期的消息, 如果查询不到就从远程拉
	messages, err := m.store.Messages.Recent(ctx, chatId, chatType, maxCachedMessages)
	if err != nil || len(messages) == 0 {
		if err != nil {
			logger.Errorf("GetRecentMessage chatId: %v, chatType: %v, error: %v", chatId, chatType, err)
		}
		messages, err = m.store.Messages.GetLatestOfflineMessages(ctx, chatId, chatType, maxCachedMessages)
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
			if lastMessage, err := m.store.Messages.LastMessageFromRemote(ctx, chatId, chatType); err != nil {
				logger.Errorf("loadInitialMessages.LastMessageFromRemote chatId: %d, chatType: %v  error: %v", chatId, chatType, err)
			} else {
				m.addMessages([]*dbhelper.ChatMessage{lastMessage})
			}
		}
	}
	m.checkMissingMessageAndSort(ctx)
}

func (m *MsgCacheImpl) updateMessage(msgs []*dbhelper.ChatMessage) {
	m.mux.Lock()
	defer m.mux.Unlock()
	ctx := context.Background()
	if len(msgs) == 0 {
		lastMessage, err := m.store.Messages.LastMessage(ctx, m.chat.ChatId, m.chat.ChatType)
		if err != nil {
			logger.Errorf("UpdateMessage.GetLastMessage error: %v", err)
		} else {
			m.addMessages([]*dbhelper.ChatMessage{lastMessage})
		}
		m.checkMissingMessageAndSort(ctx)
		return
	}
	m.addMessages(msgs)
	m.checkMissingMessageAndSort(ctx)
}

func (m *MsgCacheImpl) checkMissingMessageAndSort(ctx context.Context) {
	m.sortMessage()
	minSeq, maxSeq := checkMissingMessage(m.message)
	if minSeq == maxSeq {
		return
	}
	// 如果差距超过了50条，那么就拉去近线的50条
	var maxPull int64 = 50
	diff := maxSeq - minSeq
	if diff > maxPull {
		minSeq = maxSeq - maxPull
	}
	logger.Infof("checkMessage missing, minSeq: %v, maxSeq: %v", minSeq, maxSeq)

	msgs, err := m.store.Messages.PullOfflineMsg(ctx, m.chat.ChatId, m.chat.ChatType, minSeq, maxSeq)
	if err != nil {
		logger.Errorf("checkMessage error: %v", err)
		return
	}
	if len(msgs) == 0 {
		return
	}
	m.addMessages(msgs)
	m.sortMessage()
	m.message = m.truncateMessages(m.message)
}

func (m *MsgCacheImpl) sortMessage() {
	sortMessages(m.message)
}

func (m *MsgCacheImpl) GetMessages() []*dbhelper.ChatMessage {
	m.mux.RLock()
	defer m.mux.RUnlock()

	messages := make([]*dbhelper.ChatMessage, len(m.message))
	copy(messages, m.message)
	return messages
}

func (m *MsgCacheImpl) GetChat() *dbhelper.ImChat {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.chat
}

func (m *MsgCacheImpl) addMessages(messages []*dbhelper.ChatMessage) {
	if len(messages) == 0 {
		return
	}
	ctx := context.Background()
	for _, msg := range messages {
		if _, exists := m.dup[msg.MsgID]; exists {
			continue
		}
		m.dup[msg.MsgID] = struct{}{}
		if err := m.store.Messages.Save(ctx, msg); err != nil {
			logger.Errorf("SaveOrUpdateMessage error: %v", err)
		}
		m.message = append(m.message, msg)
	}
}

func sortMessages(messages []*dbhelper.ChatMessage) {
	if len(messages) == 0 {
		return
	}
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ServerSeq < messages[j].ServerSeq
	})
}

func (m *MsgCacheImpl) truncateMessages(messages []*dbhelper.ChatMessage) []*dbhelper.ChatMessage {
	if len(messages) <= maxCachedMessages {
		return messages
	}
	// 返回最新的 maxCachedMessages 条消息
	startIndex := len(messages) - maxCachedMessages
	msgs := make([]*dbhelper.ChatMessage, maxCachedMessages)
	copy(msgs, messages[startIndex:])
	for _, msg := range messages[:startIndex] {
		delete(m.dup, msg.MsgID)
	}
	return msgs
}

func (s *StorageImpl) ProcessIncomingMessage(msg *dbhelper.ChatMessage) []*dbhelper.ChatMessage {
	chat := dbhelper.NewImChat(s.Uid(), msg.ChatID, msg.ChatType)
	cache := s.NewCache(chat).(*MsgCacheImpl)
	cache.updateMessage([]*dbhelper.ChatMessage{msg})
	return cache.GetMessages()
}

func (m *MsgCacheImpl) Destroy() {
	m.clean()
	m.store.Messages.DestroyCache(m.chat)
}

func (m *MsgCacheImpl) clean() {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.message = make([]*dbhelper.ChatMessage, 0)
	m.dup = make(map[int64]struct{})
}

func checkMissingMessage(sortedMessage []*dbhelper.ChatMessage) (minServerSeq, maxServerSeq int64) {
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
