package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xuning888/helloIMClient/pkg"
	"github.com/xuning888/helloIMClient/svc"
	"strconv"
	"strings"
)

var (
	viewList = "list"
	viewChat = "chat"
)

var Me = svc.User{}

type Model struct {
	commonSvc   *svc.CommonSvc  // service
	selectChat  int             // 会话列表游标: 选择的会话
	currentView string          // 当前展示的模式: list 或者 chat
	textInput   textinput.Model // 输入框
	viewport    viewport.Model  // viewport
	ready       bool            // 是否准备好了
	width       int             // 宽度
	height      int             // 高度
}

func InitModel(commonSvc *svc.CommonSvc) Model {
	ti := textinput.New()
	ti.Placeholder = "请输入消息..."
	ti.Focus()
	return Model{
		commonSvc:   commonSvc,
		selectChat:  0,        // 默认选中第一个会话
		currentView: viewList, // 默认进入会话列表
		textInput:   ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-8)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 8
		}

		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.currentView == viewChat {
				m.currentView = viewList
				return m, nil
			}
		case "enter":
			if m.currentView == viewList {
				m.currentView = viewChat
				return m, nil
			} else if m.currentView == viewChat {
				value := m.textInput.Value()
				if value != "" {
					// TODO sendMessage
					fmt.Printf("sendMessage")
				}
				return m, nil
			}
		case "up":
			if m.currentView == viewList && m.selectChat > 0 {
				m.selectChat--
			}
		case "down":
			if m.currentView == viewList && m.selectChat < m.commonSvc.ChatLen()-1 {
				m.selectChat++
			}
		}
	}
	if m.currentView == viewChat {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "初始化中..."
	}
	switch m.currentView {
	case viewChat:
		return m.chatView()
	case viewList:
		return m.chatListView()
	default:
		return m.chatListView()
	}
}

func (m Model) chatListView() string {
	var content strings.Builder
	header := headerStyle.Width(m.width).Render("helloIm")
	content.WriteString(header + "\n")
	m.commonSvc.ChatRange(func(i int, chat *svc.Chat) bool {
		// 格式化最后一条消息的时间戳
		timeStr := pkg.FormatTime(chat.LastChatMessageTimestamp)

		// 基础展示信息
		chatInfo := fmt.Sprintf("%-20s %s", chat.ChatName, timeStr)
		if chat.LastChatMessage != nil {
			// 如果会话的最后一条消息不为空, 也展示进去
			chatInfo += "\n" + processLastMessage(chat.LastChatMessage)
		}
		// 处理未读数量
		if chat.UnReadNum > 0 {
			unreadBadge := unreadStyle.Render(strconv.Itoa(chat.UnReadNum))
			chatInfo += " " + unreadBadge
		}

		var item string
		if i == m.selectChat {
			item = selectedChatStyle.Width(m.width - 4).Render(chatInfo)
		} else {
			item = chatItemStyle.Width(m.width - 4).Render(chatInfo)
		}
		content.WriteString(item + "\n")
		return true
	})

	// 分割线
	separator := separatorStyle.Width(m.width).Render(strings.Repeat("─", m.width))
	content.WriteString(separator + "\n")

	// 底部的提示
	help := helpStyle.Width(m.width).Render("↑/↓ 选择 • Enter 进入 • q 退出")

	content.WriteString("\n" + help)
	return chatListStyle.Render(content.String())
}

func (m Model) chatView() string {
	if m.selectChat >= m.commonSvc.ChatLen() {
		return "会话不存在"
	}

	chat := m.commonSvc.GetChatByIndex(m.selectChat)
	var content strings.Builder
	header := headerStyle.Width(m.width).Render(fmt.Sprintf("%s", chat.ChatName))
	content.WriteString(header)

	var messages strings.Builder
	chat.Msgs.Range(func(i int, msg *svc.ChatMessage) bool {
		timeStr := pkg.FormatTime(msg.Timestamp)

		// 我发送的消息
		if msg.FromUid == Me.UserId {
			msgContent := fmt.Sprintf("%s\n%s %s", msg.Content, Me.UserName, timeStr)
			msgBlock := myMsgStyle.
				Width(m.width / 2).
				Align(lipgloss.Right).
				Render(msgContent)

			messages.WriteString(lipgloss.NewStyle().
				Align(lipgloss.Right).
				Width(m.width-4).
				Render(msgBlock) + "\n")
		} else {
			msgContent := fmt.Sprintf("%s %s\n%s", msg.FromName, timeStr, msg.Content)
			msgBlock := yourMsgStyle.
				Width(m.width / 2).
				Render(msgContent)
			messages.WriteString(msgBlock + "\n")
		}
		return true
	})

	m.viewport.SetContent(messages.String())
	content.WriteString(m.viewport.View() + "\n")

	input := inputStyle.Width(m.width).Render(m.textInput.View())
	content.WriteString(input)

	help := lipgloss.NewStyle().
		Foreground(subtextColor).
		Align(lipgloss.Center).
		Render("Enter 发送 • Esc 返回")
	content.WriteString("\n" + help)
	return content.String()
}

// processLastMessage 处理下最后一条消息
// Note: 如果是群聊会话的最后一条消息，可能出现@我的消息
func processLastMessage(lastMessage *svc.ChatMessage) string {
	if lastMessage == nil {
		return ""
	}
	content := lastMessage.Content
	return truncateText(content, 30)
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
