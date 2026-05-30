package tuiv2

import "github.com/charmbracelet/lipgloss"

var (
	backgroundColor = lipgloss.Color("#1E1E1E")
	borderColor     = lipgloss.Color("#404040")
	textColor       = lipgloss.Color("#FFFFFF")
	subtextColor    = lipgloss.Color("#888888")
	selectedColor   = lipgloss.Color("#2A2A2A")
	myMsgColor      = lipgloss.Color("#007AFF")
	otherMsgColor   = lipgloss.Color("#404040")
	headerColor     = lipgloss.Color("#2A2A2A")
)

var (
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

	chatItemStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1).
			Height(3)

	selectedChatStyle = chatItemStyle.
				Background(selectedColor).
				Foreground(textColor)
)
