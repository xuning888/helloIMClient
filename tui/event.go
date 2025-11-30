package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/service"
)

type (
	errMsg error
)

type chatListUpdatedMsg struct {
	lastMessages map[string]*sqllite.ChatMessage
	chats        []*sqllite.ImChat
	err          error
}

func fetchUpdatedChatListCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		chats, err := service.GetAllChat(ctx)
		if err != nil {
			return chatListUpdatedMsg{chats: nil, lastMessages: nil, err: err}
		}
		lastMessages := service.BatchLastMessage(ctx, chats)
		return chatListUpdatedMsg{chats: chats, lastMessages: lastMessages, err: nil}
	}
}
