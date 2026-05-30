package tuiv2

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xuning888/helloIMClient/im/dbhelper"
	"github.com/xuning888/helloIMClient/im/store"
	"github.com/xuning888/helloIMClient/pkg"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

type ChatListModel struct {
	storage      *store.Store
	cursor       int
	chats        []*dbhelper.ImChat
	lastMessages map[string]*dbhelper.ChatMessage
	width        int
	height       int
}

func NewChatListModel(storage *store.Store) ChatListModel {
	ctx := context.Background()
	chats, err := storage.Chats.ListFromRemote(ctx)
	if err != nil {
		logger.Errorf("Error loading chats: %v", err)
		chats = make([]*dbhelper.ImChat, 0)
	}
	lastMessages := storage.Messages.BatchLastMessageFromRemote(ctx, chats)
	return ChatListModel{
		storage:      storage,
		cursor:       0,
		chats:        chats,
		lastMessages: lastMessages,
	}
}

func (m ChatListModel) Init() tea.Cmd {
	return nil
}

func (m ChatListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case KeyListUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case KeyListDown:
			if m.cursor < len(m.chats)-1 {
				m.cursor++
			}
		case KeyListEnter:
			if m.cursor >= 0 && m.cursor < len(m.chats) {
				return m, FetchSelectChatCmd(m.chats[m.cursor])
			}
		case KeyListSearch:
			return m, FetchStartSearchCmd()
		}

	case ChatListUpdatedMsg:
		if msg.Err != nil {
			return m, nil
		}
		newCursor := 0
		selected := m.cursor
		if selected >= 0 && selected < len(m.chats) {
			selectedChat := m.chats[selected]
			for i, chat := range msg.Chats {
				if chat.ChatId == selectedChat.ChatId && chat.ChatType == selectedChat.ChatType {
					newCursor = i
					break
				}
			}
		}
		m.chats = msg.Chats
		m.lastMessages = msg.LastMessages
		m.cursor = newCursor
		logger.Infof("触发更新会话列表事件")
	}

	return m, nil
}

func (m ChatListModel) View() string {
	var content strings.Builder

	title := lipgloss.NewStyle().
		Width(m.width).
		Height(4).
		Background(headerColor).
		Foreground(textColor).
		Bold(true).
		Align(lipgloss.Center).
		PaddingTop(1).
		Render(" 会话列表 ")
	content.WriteString(title + "\n")

	for i, chat := range m.chats {
		var name string
		if chat.ChatType == 1 {
			if user, err := m.storage.Users.Get(context.Background(), chat.ChatId); err == nil {
				name = user.UserName
			}
		}
		lastMsg := m.lastMessages[chat.Key()]
		lastMsgText := ""
		if lastMsg != nil {
			lastMsgText = truncateText(lastMsg.MsgContent, 20)
		}

		timeStr := pkg.FormatTime(chat.UpdateTimestamp, pkg.DateTime)

		chatContent := fmt.Sprintf("%s\n%s", name, lastMsgText)
		timeContent := timeStr

		var itemStyle lipgloss.Style
		if i == m.cursor {
			itemStyle = selectedChatStyle
		} else {
			itemStyle = chatItemStyle
		}

		item := lipgloss.JoinHorizontal(lipgloss.Top,
			itemStyle.Width(m.width-10).Render(chatContent),
			lipgloss.NewStyle().Width(8).Align(lipgloss.Right).Render(timeContent),
		)

		content.WriteString(item + "\n")

		if i < len(m.chats)-1 {
			separator := lipgloss.NewStyle().
				Width(m.width).
				Foreground(borderColor).
				Render(strings.Repeat("─", m.width))
			content.WriteString(separator + "\n")
		}
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(content.String())
}

func (m *ChatListModel) UpdateSize(width, height int) {
	m.width = width
	m.height = height
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
