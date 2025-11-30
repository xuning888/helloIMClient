package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xuning888/helloIMClient/transport"
)

type commonModel struct {
	imCli    *transport.ImClient
	chatList chatListModel
	chat     *chatModel
	focus    string
	width    int
	height   int
}

func InitMainModel(imCli *transport.ImClient) tea.Model {
	chatList := initChatListModel(imCli)
	return &commonModel{
		imCli:    imCli,
		chatList: chatList,
		chat:     nil,
		focus:    "list",
	}
}

func (m commonModel) Init() tea.Cmd {
	return nil
}

func (m commonModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyTab.String():
			if m.focus == "list" {
				m.focus = "chat"
			} else {
				m.focus = "list"
			}
		case tea.KeyCtrlC.String():
			return m, tea.Quit
		}
	case selectChatMsg:
		// 选择聊天
		chatModel := initChatModel(msg.chat, m.chatList, m.imCli)
		m.chat = chatModel
		m.focus = "chat"
		m.updateLayout()
	case backToListMsg:
		// 返回聊天列表
		m.focus = "list"
		m.chat = nil
	}
	// 根据焦点分发事件
	var cmd tea.Cmd
	if m.focus == "list" {
		updatedList, listCmd := m.chatList.Update(msg)
		m.chatList = updatedList.(chatListModel)
		if listCmd != nil {
			cmd = listCmd
		}
	} else if m.chat != nil {
		updatedChat, chatCmd := m.chat.Update(msg)
		if cm, ok := updatedChat.(*chatModel); ok {
			m.chat = cm
		}
		if chatCmd != nil {
			cmd = chatCmd
		}
	}
	return m, cmd
}

func (m commonModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// 微信风格布局：左侧会话列表，右侧聊天窗口
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth

	// 左侧会话列表
	leftPanel := m.chatList.View()
	leftPanel = lipgloss.NewStyle().
		Width(leftWidth).
		Height(m.height).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(borderColor).
		Render(leftPanel)

	// 右侧内容
	var rightPanel string
	if m.chat == nil {
		rightPanel = lipgloss.Place(rightWidth, m.height, lipgloss.Center, lipgloss.Center,
			"选择一个聊天开始对话")
	} else {
		rightPanel = m.chat.View()
	}

	rightPanel = lipgloss.NewStyle().
		Width(rightWidth).
		Height(m.height).
		Render(rightPanel)

	// 组合左右面板
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// 底部状态栏
	statusBar := m.statusBarView()

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

func (m commonModel) statusBarView() string {
	focusInfo := fmt.Sprintf("焦点: %s", m.focus)
	if m.focus == "list" {
		focusInfo = "↑↓ 选择 • Space 打开 • Tab 切换 • Q 退出"
	} else {
		focusInfo = "Enter 发送 • Tab 切换 • Esc 返回 • Q 退出"
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(1).
		Background(backgroundColor).
		Foreground(subtextColor).
		Align(lipgloss.Center).
		Render(focusInfo)
}

func (m *commonModel) updateLayout() {
	if m.chat != nil {
		m.chat.updateSize(m.width/3*2, m.height-1) // 减去状态栏高度
	}
	m.chatList.updateSize(m.width/3, m.height-1)
}
