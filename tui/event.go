package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xuning888/helloIMClient/im"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
)

type chatListUpdatedMsg struct {
	lastMessages map[string]*sqllite.ChatMessage
	chats        []*sqllite.ImChat
	err          error
}

// FetchUpdatedChatListCmd 创建更新会话列表的命令
func FetchUpdatedChatListCmd(sdk *im.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		chats, err := sdk.Storage().Chats.List(ctx)
		if err != nil {
			return chatListUpdatedMsg{chats: nil, lastMessages: nil, err: err}
		}
		lastMessages := sdk.Storage().Messages.BatchLastMessage(ctx, chats)
		return chatListUpdatedMsg{chats: chats, lastMessages: lastMessages, err: nil}
	}
}

type selectChatMsg struct {
	chat *sqllite.ImChat
}

func fetchChatModel(chat *sqllite.ImChat) tea.Cmd {
	return func() tea.Msg {
		return selectChatMsg{
			chat: chat,
		}
	}
}

type backToListMsg struct{}

func FetchBackToListMsg() tea.Cmd {
	return func() tea.Msg {
		return backToListMsg{}
	}
}

type updateMessage struct {
	chatId int64
	msgs   []*sqllite.ChatMessage
}

// FetchUpdateMessage 创建更新消息的命令
func FetchUpdateMessage(chatId int64, msg []*sqllite.ChatMessage) tea.Cmd {
	return func() tea.Msg {
		return updateMessage{
			chatId: chatId,
			msgs:   msg,
		}
	}
}
