package tui

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor    = lipgloss.Color("12") // 亮蓝色
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

var (
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

	headerStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(headerColor).
			Padding(0, 1).
			Bold(true).
			Border(lipgloss.DoubleBorder(), false, false, true, false).
			BorderForeground(headerColor)

	// 会话项
	chatItemStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(textColor).
			Padding(0, 2).
			Margin(0, 0, 0, 0).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(borderColor)

	selectedChatStyle = chatItemStyle.Copy().
				Background(backgroundColor).
				Foreground(selectedColor).
				Bold(true).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(selectedColor)

	// 分隔线样式
	separatorStyle = lipgloss.NewStyle().
			Foreground(borderColor).
			Bold(true)

	// 会话列表
	chatListStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(textColor).
			Padding(0, 1)

	// 输入框样式
	inputStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(textColor).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(borderColor)
)
