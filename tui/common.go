package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xuning888/helloIMClient/internal/service"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/transport"
)

type commonModel struct {
	imCli    *transport.ImClient
	chatList chatListModel
	chat     *chatModel
	search   *searchModel
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
	var cmds []tea.Cmd = make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String():
			return m, tea.Quit
		}
	case selectChatMsg:
		// 选择聊天
		m.chat = initChatModel(msg.chat, m.imCli)
		m.focus = "chat"
		m.updateLayout()
	case backToListMsg, exitSearch: // 返回聊天列表
		m.focus = "list"
		m.chat = nil
	case startSearchMsg:
		m.search = initSearchModel()
		m.focus = "search"
		m.updateLayout()
	case searchSelectedUserMsg:
		user := msg.user
		if user != nil {
			chat, err := service.GetOrCreateChat(context.Background(), user.UserID, 1)
			if err != nil {
				logger.Errorf("创建聊天会话失败: %v", err)
				return m, nil
			}
			m.chat = initChatModel(chat, m.imCli)
			m.focus = "chat"
			m.search = nil
		}
		logger.Infof("触发搜索结果事件, user: %v", user)
		cmds = append(cmds, FetchUpdatedChatListCmd())
		m.updateLayout()
	}
	updatedList, listCmd := m.chatList.Update(msg)
	m.chatList = updatedList.(chatListModel)
	if listCmd != nil {
		cmds = append(cmds, listCmd)
	}
	if m.chat != nil {
		updatedChat, chatCmd := m.chat.Update(msg)
		if cm, ok := updatedChat.(*chatModel); ok {
			m.chat = cm
		}
		if chatCmd != nil {
			cmds = append(cmds, chatCmd)
		}
	}
	if m.search != nil {
		updatedSearch, searchCmd := m.search.Update(msg)
		if sm, ok := updatedSearch.(*searchModel); ok {
			m.search = sm
		}
		if searchCmd != nil {
			cmds = append(cmds, searchCmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m commonModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}
	if m.focus == "search" {
		if m.search != nil {
			return m.search.View()
		}
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
	if m.focus == "chat" && m.chat != nil {
		rightPanel = m.chat.View()
	} else {
		rightPanel = lipgloss.Place(rightWidth, m.height, lipgloss.Center, lipgloss.Center,
			"选择一个聊天开始对话")
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
		focusInfo = "list: ↑↓ 选择 • Space 打开 • Tab 切换 • ctrl+c 退出"
	} else if m.focus == "chat" {
		focusInfo = "chat: Enter 发送 • Tab 切换 • Esc 返回"
	} else {
		focusInfo = "search: ↑↓ 选择 • Enter 创建会话 • Esc 返回"
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
		m.chat.updateSize(m.width/3*2, m.height-1)
	}
	if m.search != nil {
		m.search.updateSize(m.width/3*2, m.height-1)
	}
	m.chatList.updateSize(m.width/3, m.height-1)
}
