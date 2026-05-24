package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xuning888/helloIMClient/im"
	sqllite2 "github.com/xuning888/helloIMClient/im/dal/sqllite"
	"github.com/xuning888/helloIMClient/pkg"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

var _ tea.Model = &chatListModel{}

type chatListModel struct {
	sdk          *im.Client
	cursor       int
	chats        []*sqllite2.ImChat
	lastMessages map[string]*sqllite2.ChatMessage
	width        int
	height       int
}

func initChatListModel(sdk *im.Client) chatListModel {
	ctx := context.Background()
	chats, err := sdk.Storage().Chats.ListFromRemote(ctx)
	if err != nil {
		logger.Errorf("Error loading chats: %v", err)
		chats = make([]*sqllite2.ImChat, 0)
	}
	lastMessages := sdk.Storage().Messages.BatchLastMessageFromRemote(ctx, chats)
	return chatListModel{
		sdk:          sdk,
		cursor:       0,
		chats:        chats,
		lastMessages: lastMessages,
	}
}

func (m chatListModel) Init() tea.Cmd {
	return nil
}

func (m chatListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyUp.String():
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown.String():
			if m.cursor < len(m.chats)-1 {
				m.cursor++
			}
		case tea.KeyEnter.String():
			if m.cursor >= 0 && m.cursor < len(m.chats) {
				chat := m.chats[m.cursor]
				return m, fetchChatModel(chat)
			}
			return m, nil
		case tea.KeyF3.String():
			return m, fetchStartSearchCmd()
		case tea.KeyCtrlC.String():
			return m, tea.Quit
		}
	case chatListUpdatedMsg:
		if msg.err != nil {
			return m, nil
		}
		newSelected := 0
		selected := m.cursor
		if selected >= 0 && selected < len(m.chats) {
			selectedChat := m.chats[selected]
			for i, chat := range msg.chats {
				if chat.ChatId == selectedChat.ChatId && chat.ChatType == selectedChat.ChatType {
					newSelected = i
					break
				}
			}
		}
		m.chats = msg.chats
		m.lastMessages = msg.lastMessages
		m.cursor = newSelected
		logger.Infof("触发更新会话列表事件")
	}
	return m, nil
}

func (m chatListModel) View() string {
	return m.chatListView()
}

func (m chatListModel) chatListView() string {
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
			if user, err := m.sdk.Storage().Users.Get(context.Background(), chat.ChatId); err == nil {
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
		timeContent := fmt.Sprintf("%s", timeStr)

		var itemStyle lipgloss.Style
		if i == m.cursor {
			itemStyle = selectedChatStyle
		} else {
			itemStyle = chatItemStyle
		}

		item := lipgloss.JoinHorizontal(lipgloss.Top,
			itemStyle.Copy().Width(m.width-10).Render(chatContent),
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

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}

func (m *chatListModel) updateSize(width, height int) {
	m.width = width
	m.height = height
}
