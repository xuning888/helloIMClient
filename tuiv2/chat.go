package tuiv2

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xuning888/helloIMClient/im/dbhelper"
	"github.com/xuning888/helloIMClient/im/store"
	"github.com/xuning888/helloIMClient/pkg"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

// SendMessageFunc 由 AppModel 注入，包装 SDK 发送逻辑
type SendMessageFunc func(ctx context.Context, chatID int64, chatType int32, content string) (*dbhelper.ChatMessage, error)

type ChatModel struct {
	cache    store.MsgCache
	storage  *store.Store
	uid      int64
	sendMsg  SendMessageFunc
	viewport viewport.Model
	textarea textarea.Model
	width    int
	height   int
}

func NewChatModel(chat *dbhelper.ImChat, storage *store.Store, uid int64, sendFn SendMessageFunc) *ChatModel {
	ta := textarea.New()
	ta.Placeholder = "输入消息..."
	ta.Focus()
	ta.Prompt = "│ "
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(50, 10)
	vp.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor)

	cache := storage.Messages.NewCache(chat)
	return &ChatModel{
		cache:    cache,
		storage:  storage,
		uid:      uid,
		sendMsg:  sendFn,
		viewport: vp,
		textarea: ta,
	}
}

func (m ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case KeyChatBack:
			return m, tea.Batch(FetchBackToListCmd(), FetchUpdatedChatListCmd(m.storage))
		case KeyChatSend:
			var message *dbhelper.ChatMessage
			if m.textarea.Focused() {
				message = m.sendMessage()
				m.textarea.Reset()
				cmds = append(cmds, viewport.Sync(m.viewport))
			}
			if message != nil {
				cmds = append(cmds, FetchUpdatedChatListCmd(m.storage))
				chatID := m.cache.GetChat().ChatId
				cmds = append(cmds, FetchUpdateMessageCmd(chatID, []*dbhelper.ChatMessage{message}))
			}
		}

	case UpdateMessageMsg:
		if m.cache.GetChat().ChatId == msg.ChatID {
			m.cache.UpdateMessage(msg.Msgs)
		}
	}

	var taCmd, vpCmd tea.Cmd
	m.textarea, taCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	if taCmd != nil {
		cmds = append(cmds, taCmd)
	}
	if vpCmd != nil {
		cmds = append(cmds, vpCmd)
	}

	return &m, tea.Batch(cmds...)
}

func (m ChatModel) View() string {
	var chatName string
	chat := m.cache.GetChat()
	if chat.ChatType == 1 {
		if user, err := m.storage.Users.Get(context.Background(), chat.ChatId); err == nil {
			chatName = user.UserName
		}
	}
	title := lipgloss.NewStyle().
		Width(m.width).
		Height(2).
		Background(headerColor).
		Foreground(textColor).
		Bold(true).
		Align(lipgloss.Center).
		Render(fmt.Sprintf("与 %s 聊天中", chatName))

	messageArea := m.viewMessage()
	messageArea = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height - 5).
		Render(messageArea)

	inputArea := lipgloss.NewStyle().
		Width(m.width).
		Height(3).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(borderColor).
		Render(m.textarea.View())

	return lipgloss.JoinVertical(lipgloss.Left, title, messageArea, inputArea)
}

func (m ChatModel) sendMessage() *dbhelper.ChatMessage {
	value := m.textarea.Value()
	if value == "" {
		return nil
	}
	chat := m.cache.GetChat()
	msg, err := m.sendMsg(context.Background(), chat.ChatId, chat.ChatType, value)
	if err != nil {
		logger.Errorf("消息发送失败, error: %v", err)
		return nil
	}
	return msg
}

func (m *ChatModel) UpdateSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width - 4
	m.viewport.Height = height - 7
	m.textarea.SetWidth(width - 2)
}

func (m ChatModel) viewMessage() string {
	chatMessages := m.cache.GetMessages()
	if len(chatMessages) == 0 {
		return lipgloss.Place(m.viewport.Width, m.viewport.Height, lipgloss.Center, lipgloss.Center,
			"暂无消息，开始对话吧！")
	}
	var messages strings.Builder
	uid := m.uid
	for _, msg := range chatMessages {
		timeStr := pkg.FormatTime(msg.SendTime, pkg.DateTime)
		if msg.MsgFrom == uid {
			content := lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.NewStyle().Foreground(subtextColor).Render(timeStr),
				msg.MsgContent,
			)
			message := myMsgStyle.Render(content)
			message = lipgloss.NewStyle().Width(m.viewport.Width).Align(lipgloss.Right).Render(message)
			messages.WriteString(message + "\n")
		} else {
			var name string
			if user, err := m.storage.Users.Get(context.Background(), msg.MsgFrom); err == nil {
				name = user.UserName
			}
			content := lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.NewStyle().Foreground(subtextColor).Render(fmt.Sprintf("%s %s", name, timeStr)),
				msg.MsgContent,
			)
			message := yourMsgStyle.Render(content)
			messages.WriteString(message + "\n")
		}
	}
	m.viewport.SetContent(messages.String())
	m.viewport.GotoBottom()
	return m.viewport.View()
}
