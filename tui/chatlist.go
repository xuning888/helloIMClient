package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/service"
	"github.com/xuning888/helloIMClient/pkg"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/transport"
)

var _ tea.Model = &chatListModel{}

type chatListModel struct {
	imCli        *transport.ImClient
	cursor       int // 游遍, 选择的会话
	chats        []*sqllite.ImChat
	lastMessages map[string]*sqllite.ChatMessage
	width        int
	height       int
}

func initChatListModel(imCli *transport.ImClient) chatListModel {
	// 拉去全部会话
	ctx := context.Background()
	chats, err := service.GetAllChat(ctx)
	if err != nil {
		logger.Errorf("Error MultiGetChat: %v", err)
		chats = make([]*sqllite.ImChat, 0)
	}
	// 获取会话的最后一条消息
	lastMessages := service.BatchLastMessage(ctx, chats)
	return chatListModel{
		imCli:        imCli,
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
		case tea.KeySpace.String(): // 选中会话
			if m.cursor >= 0 && m.cursor < len(m.chats) {
				chat := m.chats[m.cursor] // 获取会话
				return m, fetchChatModel(chat)
			}
			return m, nil
		case tea.KeyF3.String(): // 进入搜索
			return m, fetchStartSearchCmd()
		case tea.KeyCtrlC.String(): // 退出程序
			return m, tea.Quit
		}
	case chatListUpdatedMsg: // 处理会话更新事件
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
		logger.Info("更新会话列表")
	}
	return m, nil
}

func (m chatListModel) View() string {
	return m.chatListView()
}

func (m chatListModel) chatListView() string {
	var content strings.Builder

	// 列表标题
	title := lipgloss.NewStyle().
		Width(m.width).
		Height(3).
		Background(headerColor).
		Foreground(textColor).
		Bold(true).
		Align(lipgloss.Center).
		Render(" 会话列表 ")
	content.WriteString(title + "\n")

	// 会话项
	for i, chat := range m.chats {
		var name string
		if chat.ChatType == 1 {
			if user, err := sqllite.GetUserById(context.Background(), chat.ChatId); err == nil {
				name = user.UserName
			}
		}
		lastMsg := m.lastMessages[chat.Key()]
		lastMsgText := ""
		if lastMsg != nil {
			lastMsgText = truncateText(lastMsg.MsgContent, 20)
		}

		timeStr := pkg.FormatTime(chat.UpdateTimestamp)

		// 会话项内容
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

		// 分隔线
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
