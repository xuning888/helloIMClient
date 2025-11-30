package tui

import "github.com/charmbracelet/lipgloss"

// 样式优化
var (
	// 更现代的配色
	backgroundColor = lipgloss.Color("#1E1E1E") // 深灰背景
	borderColor     = lipgloss.Color("#404040") // 边框色
	textColor       = lipgloss.Color("#FFFFFF") // 白色文本
	subtextColor    = lipgloss.Color("#888888") // 灰色副文本
	selectedColor   = lipgloss.Color("#2A2A2A") // 选中项背景
	myMsgColor      = lipgloss.Color("#007AFF") // 自己消息颜色
	otherMsgColor   = lipgloss.Color("#404040") // 他人消息颜色
	headerColor     = lipgloss.Color("#2A2A2A") // 标题背景
)

var (
	// 消息样式
	myMsgStyle = lipgloss.NewStyle().
			Background(myMsgColor).
			Foreground(textColor).
			Padding(0, 1).
			Margin(0, 2, 1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(myMsgColor).
			MaxWidth(40)

	yourMsgStyle = lipgloss.NewStyle().
			Background(otherMsgColor).
			Foreground(textColor).
			Padding(0, 1).
			Margin(0, 2, 1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(otherMsgColor).
			MaxWidth(40)

	// 聊天列表样式
	chatItemStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(textColor).
			Padding(0, 1).
			Height(3)

	selectedChatStyle = chatItemStyle.Copy().
				Background(selectedColor).
				Foreground(textColor)
)
