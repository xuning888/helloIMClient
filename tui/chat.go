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

const gap = "\n\n"

type chatModel struct {
	imCli       *transport.ImClient
	chat        *sqllite.ImChat
	viewport    viewport.Model
	textarea    textarea.Model
	senderStyle lipgloss.Style
	messages    []*sqllite.ChatMessage
	chatList    chatListModel
	err         error
	width       int
	height      int
}

func initChatModel(chat *sqllite.ImChat, chatList chatListModel, imCli *transport.ImClient) tea.Model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.SetWidth(50)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(30, 5)

	// 获取最新离线消息
	messages, err := service.GetLatestOfflineMessages(context.Background(), chat.ChatId, chat.ChatType)
	if err != nil {
		logger.Errorf("GetLatestOfflineMessages error %v", err)
		messages = make([]*sqllite.ChatMessage, 0)
	}
	return &chatModel{
		imCli:       imCli,
		chat:        chat,
		viewport:    vp,
		textarea:    ta,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		messages:    messages,
		chatList:    chatList,
	}
}

func (m chatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tiCmd, vpCmd tea.Cmd
	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = m.width
		m.textarea.SetWidth(m.width)
		m.viewport.Height = m.height - m.textarea.Height() - lipgloss.Height(gap)
		if len(m.messages) > 0 {
			m.updateViewportContent()
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m.chatList, fetchUpdatedChatListCmd()
		case tea.KeyEnter:
			m.sendMessage()
			m.updateViewportContent()
			m.viewport.GotoBottom()
			m.textarea.Reset()
			m.chatList.Update(fetchUpdatedChatListCmd())
		}
	case errMsg:
		m.err = msg
		return m, nil
	}
	return m, tea.Batch(tiCmd, vpCmd)
}

func (m chatModel) View() string {
	return m.messageView()
}

func (m chatModel) sendMessage() {
	value := m.textarea.Value()
	if value == "" {
		return
	}
	request := c2csend.NewRequest(conf.UserId, m.chat.ChatId, value, 0, 0, 0)
	response, err := m.imCli.WriteMessage(context.Background(), request)
	if err != nil {
		logger.Errorf("消息发送失败, error: %v", err)
		m.textarea.SetValue("")
		return
	}
	m.saveC2CMessage(request, response)
}

func (m chatModel) saveC2CMessage(request *c2csend.Request, response protocol.Response) {
	chat := m.chat
	message := sqllite.NewMessage(1, chat.ChatId, response.MsgId(), conf.UserId, chat.ChatId,
		0, 0, 0, request.Content, request.ContentType, request.CmdId(),
		request.SendTimestamp, 0, response.ServerSeq())
	if err := sqllite.SaveOrUpdateMessage(context.Background(), message); err != nil {
		logger.Errorf("saveC2CMessage error: %v", err)
		return
	}
	m.updateMessage()
}

func (m chatModel) updateMessage() {
	messages, err := sqllite.GetRecentMessages(context.Background(), m.chat.ChatId, 20)
	if err != nil {
		return
	}
	m.messages = messages
}

func (m chatModel) updateViewportContent() {
	message := m.concatMessage()
	m.viewport.SetContent(message)
}

func (m chatModel) concatMessage() string {
	if len(m.messages) == 0 {
		return ""
	}
	var sbd strings.Builder
	halfWidth := m.width / 2
	fullWidth := m.width - 4

	for _, msg := range m.messages {
		timeStr := pkg.FormatTime(msg.SendTime)
		if msg.MsgFrom == conf.UserId {
			sbd.WriteString(m.renderMyMessage(msg.MsgContent, timeStr, halfWidth, fullWidth))
		} else {
			name := ""
			if user, err := sqllite.GetUserById(context.Background(), msg.MsgFrom); err == nil {
				name = user.UserName
			} else {
				logger.Errorf("GetUserById error: %v", err)
			}
			sbd.WriteString(m.renderYourMessage(name, msg.MsgContent, timeStr, halfWidth))
		}
	}
	return sbd.String()
}

func (m chatModel) renderMyMessage(content, time string, halfWidth, fullWidth int) string {
	msgContent := fmt.Sprintf("%s\n%s %s", content, "You", time)
	msgBlock := myMsgStyle.Width(halfWidth).Align(lipgloss.Right).Render(msgContent)
	return lipgloss.NewStyle().Align(lipgloss.Right).Width(fullWidth).Render(msgBlock) + "\n"
}

func (m chatModel) renderYourMessage(name, content, time string, halfWidth int) string {
	msgContent := fmt.Sprintf("%s %s\n%s", name, time, content)
	msgBlock := yourMsgStyle.Width(halfWidth).Render(msgContent)
	return msgBlock + "\n"
}

func (m chatModel) messageView() string {
	var content strings.Builder

	message := m.concatMessage()

	m.viewport.SetContent(message)
	content.WriteString(m.viewport.View() + "\n")

	input := inputStyle.Width(m.width).Render(m.textarea.View())
	content.WriteString(input)

	help := lipgloss.NewStyle().
		Foreground(subtextColor).
		Align(lipgloss.Center).
		Render("Enter 发送 • Esc 返回")
	content.WriteString("\n" + help)
	return content.String()
}
