package tuiv2

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xuning888/helloIMClient/im/dbhelper"
	"github.com/xuning888/helloIMClient/im/store"
)

// ---- Navigation ----

type SelectChatMsg struct {
	Chat *dbhelper.ImChat
}

func FetchSelectChatCmd(chat *dbhelper.ImChat) tea.Cmd {
	return func() tea.Msg {
		return SelectChatMsg{Chat: chat}
	}
}

type BackToListMsg struct{}

func FetchBackToListCmd() tea.Cmd {
	return func() tea.Msg {
		return BackToListMsg{}
	}
}

// ---- Chat list ----

type ChatListUpdatedMsg struct {
	Chats        []*dbhelper.ImChat
	LastMessages map[string]*dbhelper.ChatMessage
	Err          error
}

func FetchUpdatedChatListCmd(storage *store.Store) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		chats, err := storage.Chats.List(ctx)
		if err != nil {
			return ChatListUpdatedMsg{Err: err}
		}
		lastMessages := storage.Messages.BatchLastMessage(ctx, chats)
		return ChatListUpdatedMsg{Chats: chats, LastMessages: lastMessages}
	}
}

// ---- Chat messages ----

type UpdateMessageMsg struct {
	ChatID int64
	Msgs   []*dbhelper.ChatMessage
}

func FetchUpdateMessageCmd(chatID int64, msgs []*dbhelper.ChatMessage) tea.Cmd {
	return func() tea.Msg {
		return UpdateMessageMsg{ChatID: chatID, Msgs: msgs}
	}
}

// ---- Search ----

type StartSearchMsg struct{}

func FetchStartSearchCmd() tea.Cmd {
	return func() tea.Msg {
		return StartSearchMsg{}
	}
}

type ExitSearchMsg struct{}

func FetchExitSearchCmd() tea.Cmd {
	return func() tea.Msg {
		return ExitSearchMsg{}
	}
}

type SearchUserMsg struct {
	Key   string
	Users []*dbhelper.ImUser
	Err   error
}

func FetchSearchUserCmd(storage *store.Store, key string) tea.Cmd {
	return func() tea.Msg {
		users, err := storage.Users.Search(context.Background(), key)
		return SearchUserMsg{Key: key, Users: users, Err: err}
	}
}

type SearchSelectedUserMsg struct {
	User *dbhelper.ImUser
}

func FetchSearchSelectedUserCmd(user *dbhelper.ImUser) tea.Cmd {
	return func() tea.Msg {
		return SearchSelectedUserMsg{User: user}
	}
}
