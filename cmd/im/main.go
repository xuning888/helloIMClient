// main.go
package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/xuning888/helloIMClient/svc"
	"log"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	Me                  = &svc.User{}
	chats               []*svc.Chat
	currentChatMessages []*svc.ChatMessage
)

func main() {
	// 初始化示例数据
	initSampleData()

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func initSampleData() {
	// 设置当前用户
	Me.UserId = 111
	Me.UserName = "我"
	Me.UserType = 1

	// 创建示例聊天会话
	chats = []*svc.Chat{
		{
			Id:        1,
			Type:      svc.C2C,
			ChatName:  "张三",
			Timestamp: time.Now().Unix() - 300,
			UnReadNum: 2,
			LastChatMessage: &svc.ChatMessage{
				MsgId:    1,
				ChatId:   1,
				ChatType: int32(svc.C2C),
				MsgFrom:  222,
				ToUid:    111,
				Content:  "你好，在忙什么呢？",
			},
		},
		{
			Id:        2,
			Type:      svc.C2C,
			ChatName:  "李四",
			Timestamp: time.Now().Unix() - 600,
			UnReadNum: 0,
			LastChatMessage: &svc.ChatMessage{
				MsgId:    2,
				ChatId:   "2",
				ChatType: svc.C2C,
				FromUid:  333,
				FromName: "我",
				ToUid:    111,
				Content:  "好的，明天见",
			},
		},
		{
			Id:        3,
			Type:      svc.C2G,
			ChatName:  "开发团队群",
			Timestamp: time.Now().Unix() - 1200,
			UnReadNum: 5,
			LastChatMessage: &svc.ChatMessage{
				MsgId:    3,
				ChatId:   "3",
				ChatType: svc.C2G,
				FromUid:  444,
				FromName: "王五",
				ToUid:    555,
				Content:  "大家记得参加下午的会议",
			},
		},
	}
}

type model struct {
	chatList     []*svc.Chat
	selectedChat int
	currentView  string // "list" or "chat"
	textInput    textinput.Model
	viewport     viewport.Model
	ready        bool
	width        int
	height       int
}

type viewMsg struct{}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "输入消息..."
	ti.Focus()
	return model{
		selectedChat: 0,
		currentView:  "list",
		textInput:    ti,
	}
}

// 颜色定义 - 控制台风格
var (
	backgroundColor = lipgloss.Color("0")  // 黑色背景
	borderColor     = lipgloss.Color("8")  // 灰色边框
	textColor       = lipgloss.Color("15") // 白色文本
	subtextColor    = lipgloss.Color("8")  // 暗灰色文本
	selectedColor   = lipgloss.Color("14") // 亮青色
	myMsgColor      = lipgloss.Color("10") // 亮绿色
	otherMsgColor   = lipgloss.Color("11") // 亮黄色
	unreadColor     = lipgloss.Color("9")  // 亮红色
	headerColor     = lipgloss.Color("13") // 亮紫色
)

// 样式定义
var (
	// 会话列表
	chatListStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(textColor).
			Padding(0, 1)

	// 会话项
	chatItemStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(textColor).
			Padding(0, 2).
			Margin(0, 0, 0, 0).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(borderColor)

	// 会话项选中后的样式
	selectedChatStyle = chatItemStyle.Copy().
				Background(backgroundColor).
				Foreground(selectedColor).
				Bold(true).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(selectedColor)

	// 我发送的消息的样式
	myMsgStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(myMsgColor).
			Padding(0, 1).
			Margin(0, 0, 1, 0).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(myMsgColor)

	// 别人发送的消息的样式
	yourMsgStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(otherMsgColor).
			Padding(0, 1).
			Margin(0, 0, 1, 0).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(otherMsgColor)

	// 输入框样式
	inputStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(textColor).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(borderColor)

	// 标题栏样式
	headerStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(headerColor).
			Padding(0, 1).
			Bold(true).
			Border(lipgloss.DoubleBorder(), false, false, true, false).
			BorderForeground(headerColor)

	// 未读消息数样式
	unreadStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(unreadColor).
			Bold(true)

	// 分隔线样式
	separatorStyle = lipgloss.NewStyle().
			Foreground(borderColor).
			Bold(true)

	// 帮助信息样式
	helpStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(subtextColor).
			Italic(true).
			Align(lipgloss.Center).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(borderColor)
)

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.currentView == "chat" {
				m.currentView = "list"
				return m, nil
			}

		case "enter":
			if m.currentView == "list" {
				// 进入聊天界面
				m.currentView = "chat"
				m.loadChatMessages()
				return m, nil
			} else if m.currentView == "chat" {
				// 发送消息
				if m.textInput.Value() != "" {
					m.sendMessage(m.textInput.Value())
					m.textInput.SetValue("")
				}
				return m, nil
			}
		case "up":
			if m.currentView == "list" && m.selectedChat > 0 {
				m.selectedChat--
			}

		case "down":
			if m.currentView == "list" && m.selectedChat < len(chats)-1 {
				m.selectedChat++
			}
		}
	}

	if m.currentView == "chat" {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *model) loadChatMessages() {
	// 模拟加载聊天消息
	if m.selectedChat < len(chats) {
		chat := chats[m.selectedChat]

		// 创建示例消息
		currentChatMessages = []*svc.ChatMessage{
			{
				MsgId:     1,
				ChatId:    strconv.FormatInt(chat.Id, 10),
				ChatType:  chat.Type,
				FromUid:   222,
				FromName:  chat.ChatName,
				ToUid:     111,
				Content:   "你好！",
				Timestamp: time.Now().Unix() - 3600,
			},
			{
				MsgId:     2,
				ChatId:    strconv.FormatInt(chat.Id, 10),
				ChatType:  chat.Type,
				FromUid:   Me.UserId,
				FromName:  Me.UserName,
				ToUid:     222,
				Content:   "你好，有什么事吗？",
				Timestamp: time.Now().Unix() - 3500,
			},
			{
				MsgId:     3,
				ChatId:    strconv.FormatInt(chat.Id, 10),
				ChatType:  chat.Type,
				FromUid:   222,
				FromName:  chat.ChatName,
				ToUid:     Me.UserId,
				Content:   chat.LastChatMessage.Content,
				Timestamp: time.Now().Unix() - 300,
			},
		}
	}
}

func (m *model) sendMessage(content string) {
	if m.selectedChat < len(chats) {
		chat := chats[m.selectedChat]

		newMsg := &svc.ChatMessage{
			MsgId:     int64(len(currentChatMessages) + 1),
			ChatId:    strconv.FormatInt(chat.Id, 10),
			ChatType:  chat.Type,
			FromUid:   Me.UserId,
			FromName:  Me.UserName,
			ToUid:     222,
			Content:   content,
			Timestamp: time.Now().Unix(),
		}

		currentChatMessages = append(currentChatMessages, newMsg)

		// 更新聊天列表中的最后一条消息
		chat.LastChatMessage = newMsg
		chat.Timestamp = newMsg.Timestamp
	}
}

func (m model) View() string {
	if !m.ready {
		return "初始化中..."
	}

	switch m.currentView {
	case "list":
		return m.chatListView()
	case "chat":
		return m.chatView()
	default:
		return m.chatListView()
	}
}

func (m model) chatListView() string {
	var content strings.Builder

	// 标题栏
	title := fmt.Sprintf("┌─ HelloIM Console ─ 会话列表 ─┐")
	header := headerStyle.Width(m.width).Render(title)
	content.WriteString(header + "\n")

	// 聊天列表
	for i, chat := range chats {
		var item string

		// 格式化时间
		timeStr := formatTime(chat.Timestamp)

		// 构建聊天项内容
		chatInfo := fmt.Sprintf("%-20s %s", chat.ChatName, timeStr)
		if chat.LastChatMessage != nil {
			chatInfo += "\n" + truncateText(chat.LastChatMessage.Content, 30)
		}

		// 添加未读消息数
		if chat.UnReadNum > 0 {
			unreadBadge := unreadStyle.Render(strconv.Itoa(chat.UnReadNum))
			chatInfo += " " + unreadBadge
		}

		// 应用样式
		if i == m.selectedChat {
			item = selectedChatStyle.Width(m.width - 4).Render(chatInfo)
		} else {
			item = chatItemStyle.Width(m.width - 4).Render(chatInfo)
		}
		content.WriteString(item + "\n")
	}

	// 底部提示
	help := lipgloss.NewStyle().
		Foreground(subtextColor).
		Align(lipgloss.Center).
		Render("↑/↓ 选择 • Enter 进入 • q 退出")

	content.WriteString("\n" + help)

	return chatListStyle.Render(content.String())
}

func (m model) chatView() string {
	if m.selectedChat >= len(chats) {
		return "聊天不存在"
	}

	chat := chats[m.selectedChat]
	var content strings.Builder

	// 标题栏
	header := headerStyle.Width(m.width).Render(fmt.Sprintf("← %s", chat.ChatName))
	content.WriteString(header + "\n")

	// 消息列表
	var messages strings.Builder
	for _, msg := range currentChatMessages {
		timeStr := formatTime(msg.Timestamp)

		if msg.FromUid == Me.UserId {
			// 我的消息（右对齐）
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
			// 其他人的消息（左对齐）
			msgContent := fmt.Sprintf("%s %s\n%s", msg.FromName, timeStr, msg.Content)
			msgBlock := yourMsgStyle.
				Width(m.width / 2).
				Render(msgContent)
			messages.WriteString(msgBlock + "\n")
		}
	}

	// 使用 viewport 显示消息
	m.viewport.SetContent(messages.String())
	content.WriteString(m.viewport.View() + "\n")

	// 输入框
	input := inputStyle.Width(m.width).Render(m.textInput.View())
	content.WriteString(input)

	// 底部提示
	help := lipgloss.NewStyle().
		Foreground(subtextColor).
		Align(lipgloss.Center).
		Render("Enter 发送 • Esc 返回")
	content.WriteString("\n" + help)

	return content.String()
}

// 辅助函数
func formatTime(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	now := time.Now()

	if t.Day() == now.Day() && t.Month() == now.Month() && t.Year() == now.Year() {
		return t.Format("15:04")
	} else if t.Year() == now.Year() {
		return t.Format("01/02")
	} else {
		return t.Format("2006/01/02")
	}
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
