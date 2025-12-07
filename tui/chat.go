package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xuning888/helloIMClient/conf"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/service"
	"github.com/xuning888/helloIMClient/pkg"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/protocol"
	"github.com/xuning888/helloIMClient/protocol/c2csend"
	"github.com/xuning888/helloIMClient/transport"
)

var _ tea.Model = &chatModel{}

type chatModel struct {
	cache    *service.MsgCache
	imCli    *transport.ImClient
	viewport viewport.Model
	textarea textarea.Model
	width    int
	height   int
}

func initChatModel(chat *sqllite.ImChat, imCli *transport.ImClient) *chatModel {
	ta := textarea.New()
	ta.Placeholder = "输入消息..."
	ta.Focus()
	ta.Prompt = "│ "
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(50, 10)
	vp.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor)

	cache := service.NewMsgCache(chat)
	return &chatModel{
		cache:    cache,
		imCli:    imCli,
		viewport: vp,
		textarea: ta,
	}
}

func (m chatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			cmds = append(cmds, FetchBackToListMsg())
			cmds = append(cmds, FetchUpdatedChatListCmd())
			return m, tea.Batch(cmds...)
		case tea.KeyEnter:
			if m.textarea.Focused() {
				m.sendMessage()
				m.textarea.Reset()
				cmds = append(cmds, viewport.Sync(m.viewport))
			}
			cmds = append(cmds, FetchUpdatedChatListCmd())
		}
	case updateMessage:
		if m.cache.GetChat().ChatId == msg.chatId {
			m.updateMessage()
		}
	}
	// 更新子组件
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

func (m chatModel) View() string {
	var chatName string
	chat := m.cache.GetChat()
	if chat.ChatType == 1 {
		if user, err := sqllite.GetUserById(context.Background(), chat.ChatId); err == nil {
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

	// 消息区域
	messageArea := m.viewMessage()
	messageArea = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height - 5). // 减去标题和输入框高度
		Render(messageArea)

	// 输入区域
	inputArea := lipgloss.NewStyle().
		Width(m.width).
		Height(3).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(borderColor).
		Render(m.textarea.View())

	return lipgloss.JoinVertical(lipgloss.Left, title, messageArea, inputArea)
}

func (m chatModel) sendMessage() {
	value := m.textarea.Value()
	if value == "" {
		return
	}
	chat := m.cache.GetChat()
	request := c2csend.NewRequest(conf.UserId, chat.ChatId, value, 0, 0, 0)
	response, err := m.imCli.WriteMessage(context.Background(), request)
	if err != nil {
		logger.Errorf("消息发送失败, error: %v", err)
		m.textarea.SetValue("")
		return
	}
	m.saveC2CMessage(request, response)
}

func (m *chatModel) updateSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width - 4
	m.viewport.Height = height - 7 // 调整视口大小
	m.textarea.SetWidth(width - 2)
}

func (m chatModel) saveC2CMessage(request *c2csend.Request, response protocol.Response) {
	chat := m.cache.GetChat()
	message := sqllite.NewMessage(1, chat.ChatId, response.MsgId(), conf.UserId, chat.ChatId,
		0, 0, 0, request.Content, request.ContentType, request.CmdId(),
		request.SendTimestamp, 0, response.ServerSeq())
	if err := sqllite.SaveOrUpdateMessage(context.Background(), message); err != nil {
		logger.Errorf("saveC2CMessage error: %v", err)
		return
	}
	pm := &m
	pm.updateMessage()
}

func (m *chatModel) updateMessage() {
	m.cache.UpdateMessage()
}

func (m chatModel) viewMessage() string {
	chatMessages := m.cache.GetMessages()
	if len(chatMessages) == 0 {
		return lipgloss.Place(m.viewport.Width, m.viewport.Height, lipgloss.Center, lipgloss.Center,
			"暂无消息，开始对话吧！")
	}
	var messages strings.Builder
	for _, msg := range chatMessages {
		timeStr := pkg.FormatTime(msg.SendTime, pkg.DateTime)
		if msg.MsgFrom == conf.UserId {
			// 自己发送的消息，靠右显示
			content := lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.NewStyle().Foreground(subtextColor).Render(timeStr),
				msg.MsgContent,
			)
			message := myMsgStyle.Render(content)
			message = lipgloss.NewStyle().Width(m.viewport.Width).Align(lipgloss.Right).Render(message)
			messages.WriteString(message + "\n")
		} else {
			// 对方发送的消息，靠左显示
			var name string
			if user, err := sqllite.GetUserById(context.Background(), msg.MsgFrom); err == nil {
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
