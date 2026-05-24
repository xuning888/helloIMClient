package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xuning888/helloIMClient/im"
	sqllite2 "github.com/xuning888/helloIMClient/im/dal/sqllite"
	"github.com/xuning888/helloIMClient/im/payload"
	"github.com/xuning888/helloIMClient/im/protocol/send"
	"github.com/xuning888/helloIMClient/pkg"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

var _ tea.Model = &chatModel{}

type chatModel struct {
	cache    im.MsgCache
	sdk      *im.Client
	viewport viewport.Model
	textarea textarea.Model
	width    int
	height   int
}

func initChatModel(chat *sqllite2.ImChat, sdk *im.Client) *chatModel {
	ta := textarea.New()
	ta.Placeholder = "输入消息..."
	ta.Focus()
	ta.Prompt = "│ "
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(50, 10)
	vp.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor)

	cache := sdk.Storage().Messages.NewCache(chat)
	return &chatModel{
		cache:    cache,
		sdk:      sdk,
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
			cmds = append(cmds, FetchBackToListMsg(), FetchUpdatedChatListCmd(m.sdk))
			return m, tea.Batch(cmds...)
		case tea.KeyEnter:
			var message *sqllite2.ChatMessage = nil
			if m.textarea.Focused() {
				message = m.sendMessage()
				m.textarea.Reset()
				cmds = append(cmds, viewport.Sync(m.viewport))
			}
			if message != nil {
				cmds = append(cmds, FetchUpdatedChatListCmd(m.sdk))
				chatId := m.cache.GetChat().ChatId
				cmds = append(cmds, FetchUpdateMessage(chatId, []*sqllite2.ChatMessage{message}))
			}
		}
	case updateMessage:
		if m.cache.GetChat().ChatId == msg.chatId {
			m.cache.UpdateMessage(msg.msgs)
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

func (m chatModel) View() string {
	var chatName string
	chat := m.cache.GetChat()
	if chat.ChatType == 1 {
		if user, err := m.sdk.Storage().Users.Get(context.Background(), chat.ChatId); err == nil {
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

func (m chatModel) sendMessage() *sqllite2.ChatMessage {
	value := m.textarea.Value()
	if value == "" {
		return nil
	}
	chat := m.cache.GetChat()
	p := payload.NewTextMessage(value, false, nil)
	request := send.NewSendMsg(m.sdk.GetUID(), chat.ChatId, chat.ChatType, p, 0, 0)
	ack, err := m.sdk.SendMessage(context.Background(), request)
	if err != nil {
		logger.Errorf("消息发送失败, error: %v", err)
		m.textarea.SetValue("")
		return nil
	}
	sendAck, ok := ack.(*send.SendAck)
	if !ok {
		return nil
	}
	msg := m.saveSentMessage(request, sendAck)
	return msg
}

func (m *chatModel) updateSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width - 4
	m.viewport.Height = height - 7
	m.textarea.SetWidth(width - 2)
}

func (m chatModel) saveSentMessage(req *send.SendMsg, ack *send.SendAck) *sqllite2.ChatMessage {
	chat := m.cache.GetChat()
	uid := m.sdk.GetUID()
	content, contentType := payload.ExtractContent(req.Payload)
	message := sqllite2.NewMessage(chat.ChatType, chat.ChatId, ack.MsgId(), uid, chat.ChatId,
		req.FromUserType, req.ToUserType, ack.MsgSeq(), content, contentType, req.CmdId(),
		req.SendTimestamp, 0, ack.ServerSeq())
	if err := m.sdk.Storage().Messages.Save(context.Background(), message); err != nil {
		logger.Errorf("saveSentMessage error: %v", err)
	}
	m.sdk.Storage().Chats.UpdateVersion(context.Background(), chat.ChatId, chat.ChatType)
	return message
}

func (m chatModel) viewMessage() string {
	chatMessages := m.cache.GetMessages()
	if len(chatMessages) == 0 {
		return lipgloss.Place(m.viewport.Width, m.viewport.Height, lipgloss.Center, lipgloss.Center,
			"暂无消息，开始对话吧！")
	}
	var messages strings.Builder
	uid := m.sdk.GetUID()
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
			if user, err := m.sdk.Storage().Users.Get(context.Background(), msg.MsgFrom); err == nil {
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
