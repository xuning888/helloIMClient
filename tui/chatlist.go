package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

func InitChatListModel(imCli *transport.ImClient) tea.Model {
	// 拉去全部会话
	ctx := context.Background()
	chats, err := service.GetAllChat(ctx)
	if err != nil {
		logger.Errorf("Error MultiGetChat: %v", err)
		chats = make([]*sqllite.ImChat, 0)
	}
	// 获取会话的最后一条消息
	lastMessages := service.BatchLastMessage(ctx, chats)
	return &chatListModel{
		imCli:        imCli,
		cursor:       0,
		chats:        chats,
		lastMessages: lastMessages,
		width:        500,
		height:       500,
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
			chat := m.chats[m.cursor]                // 获取会话
			model := initChatModel(chat, m, m.imCli) // 进入会话
			return model, nil
		case tea.KeyCtrlC.String(): // 退出程序
			return m, tea.Quit
		}
	case chatListUpdatedMsg: // 处理会话更新事件
		if msg.err != nil {
			return m, nil
		}
		selected := m.cursor
		selectedChat := m.chats[selected]
		newSelected := 0
		for i, chat := range msg.chats {
			if chat.ChatId == selectedChat.ChatId && chat.ChatType == selectedChat.ChatType {
				newSelected = i
				break
			}
		}
		m.chats = msg.chats
		m.lastMessages = msg.lastMessages
		m.cursor = newSelected
		fmt.Println("更新会话列表")
	}
	return m, nil
}

func (m chatListModel) View() string {
	return m.chatListView()
}

func (m chatListModel) chatListView() string {
	var content strings.Builder
	header := headerStyle.Width(m.width).Render("helloIm")
	content.WriteString(header + "\n")
	for i, chat := range m.chats {
		timeStr := pkg.FormatTime(chat.UpdateTimestamp)
		var name = ""
		if chat.ChatType == 1 {
			if user, err := sqllite.GetUserById(context.Background(), chat.ChatId); err == nil {
				name = user.UserName
			}
		}
		chatInfo := fmt.Sprintf("%-20s %s", name, timeStr)
		lastMessage := m.lastMessages[chat.Key()]
		if lastMessage != nil {
			// 如果会话的最后一条消息不为空, 也展示进去
			chatInfo += "\n" + processLastMessage(lastMessage)
		}
		var item string
		if i == m.cursor {
			item = selectedChatStyle.Width(m.width - 4).Render(chatInfo)
		} else {
			item = chatItemStyle.Width(m.width - 4).Render(chatInfo)
		}
		content.WriteString(item + "\n")
	}
	// 分割线
	separator := separatorStyle.Width(m.width).Render(strings.Repeat("─", m.width))
	content.WriteString(separator + "\n")

	return chatListStyle.Render(content.String())
}

func processLastMessage(lastMessage *sqllite.ChatMessage) string {
	if lastMessage == nil {
		return ""
	}
	content := lastMessage.MsgContent
	return truncateText(content, 30)
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
