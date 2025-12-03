package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

var _ tea.Model = &searchModel{}

type searchModel struct {
	searchInput   textarea.Model
	searchResults []*sqllite.ImUser
	width         int
	height        int
	cursor        int
	searching     bool
}

func initSearchModel() *searchModel {
	searchTa := textarea.New()
	searchTa.Placeholder = "输入用户名搜索..."
	searchTa.Focus()
	searchTa.ShowLineNumbers = false
	searchTa.KeyMap.InsertNewline.SetEnabled(false)
	return &searchModel{
		searchInput:   searchTa,
		searchResults: make([]*sqllite.ImUser, 0),
		cursor:        0,
		searching:     false,
	}
}

func (m searchModel) Init() tea.Cmd {
	return nil
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyEsc.String(): // 退出搜索
			return m, fetchExitSearchMsg()
		case tea.KeyEnter.String():
			var user *sqllite.ImUser = nil
			if len(m.searchResults) != 0 {
				if m.cursor >= 0 && m.cursor < len(m.searchResults) {
					user = m.searchResults[m.cursor]
				}
			}
			cmds = append(cmds, fetchSearchSelectedUserMsg(user), FetchUpdatedChatListCmd())
			return m, tea.Batch(cmds...)
		case tea.KeyUp.String():
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown.String():
			if m.cursor < len(m.searchResults)-1 {
				m.cursor++
			}
		default: // 默认搜索用户
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			cmds = append(cmds, cmd)
			searchKey := strings.TrimSpace(m.searchInput.Value())
			if searchKey != "" {
				m.searching = true
				cmds = append(cmds, fetchSearchUserMsg(searchKey))
			} else {
				m.searchResults = make([]*sqllite.ImUser, 0)
				m.searching = false
			}
		}
	case searchUserMsg:
		m.searching = false
		if msg.err == nil {
			m.searchResults = msg.users
		} else {
			logger.Errorf("搜索用户失败: %v", msg.err)
			m.searchResults = make([]*sqllite.ImUser, 0)
		}
	}
	return &m, tea.Batch(cmds...)
}

func (m searchModel) View() string {
	var content strings.Builder

	// 标题
	title := lipgloss.NewStyle().
		Width(m.width).
		Height(2).
		Background(headerColor).
		Foreground(textColor).
		Bold(true).
		Align(lipgloss.Center).
		Render("搜索用户")
	content.WriteString(title + "\n")

	// 搜索框
	searchBox := m.renderSearchBox()
	content.WriteString(searchBox + "\n")

	// 分隔线
	separator := lipgloss.NewStyle().
		Width(m.width).
		Foreground(borderColor).
		Render(strings.Repeat("─", m.width))
	content.WriteString(separator + "\n")

	// 搜索结果
	content.WriteString(m.renderSearchResults())

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(content.String())
}

func (m searchModel) renderSearchBox() string {
	searchLabel := "搜索: "
	searchInput := m.searchInput.View()

	searchBox := lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Width(len(searchLabel)).Render(searchLabel),
		searchInput,
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Padding(1, 2).
		Render(searchBox)
}

func (m searchModel) renderSearchResults() string {
	var results strings.Builder
	if m.searching {
		results.WriteString(lipgloss.NewStyle().Padding(1, 2).Render("搜索中...\n"))
	} else if len(m.searchResults) == 0 {
		if strings.TrimSpace(m.searchInput.Value()) != "" {
			results.WriteString(lipgloss.NewStyle().Padding(1, 2).Render("未找到用户\n"))
		} else {
			results.WriteString(lipgloss.NewStyle().Padding(1, 2).Render("输入用户名进行搜索\n"))
		}
	} else {
		for i, user := range m.searchResults {
			var userStyle lipgloss.Style
			if i == m.cursor {
				userStyle = selectedChatStyle
			} else {
				userStyle = chatItemStyle
			}
			userInfo := fmt.Sprintf("%s (ID: %d)", user.UserName, user.UserID)
			resultItem := userStyle.Render(userInfo)
			results.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(resultItem) + "\n")
			if i < len(m.searchResults)-1 {
				separator := lipgloss.NewStyle().
					Width(m.width - 4). // 减去padding
					Foreground(borderColor).
					Render(strings.Repeat("─", m.width-4))
				results.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(separator) + "\n")
			}
		}
	}
	return results.String()
}

func (m *searchModel) updateSize(w, h int) {
	m.width = w
	m.height = h
}

type exitSearch struct {
}

func fetchExitSearchMsg() tea.Cmd {
	return func() tea.Msg {
		return exitSearch{}
	}
}

type searchSelectedUserMsg struct {
	user *sqllite.ImUser
}

func fetchSearchSelectedUserMsg(user *sqllite.ImUser) tea.Cmd {
	return func() tea.Msg {
		return searchSelectedUserMsg{
			user: user,
		}
	}
}

type searchUserMsg struct {
	key   string
	users []*sqllite.ImUser
	err   error
}

func fetchSearchUserMsg(key string) tea.Cmd {
	return func() tea.Msg {
		users, err := sqllite.SearchUser(context.Background(), key)
		return searchUserMsg{
			key:   key,
			users: users,
			err:   err,
		}
	}
}

type startSearchMsg struct{}

func fetchStartSearchCmd() tea.Cmd {
	return func() tea.Msg {
		return startSearchMsg{}
	}
}
