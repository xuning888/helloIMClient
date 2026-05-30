package tuiv2

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xuning888/helloIMClient/im/dbhelper"
	"github.com/xuning888/helloIMClient/im/store"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

type SearchModel struct {
	storage       *store.Store
	searchInput   textarea.Model
	searchResults []*dbhelper.ImUser
	width         int
	height        int
	cursor        int
	searching     bool
}

func NewSearchModel(storage *store.Store) *SearchModel {
	searchTa := textarea.New()
	searchTa.Placeholder = "输入用户名搜索..."
	searchTa.Focus()
	searchTa.ShowLineNumbers = false
	searchTa.KeyMap.InsertNewline.SetEnabled(false)
	return &SearchModel{
		storage:       storage,
		searchInput:   searchTa,
		searchResults: make([]*dbhelper.ImUser, 0),
		cursor:        0,
		searching:     false,
	}
}

func (m SearchModel) Init() tea.Cmd {
	return nil
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case KeySearchBack:
			return m, FetchExitSearchCmd()
		case KeySearchSelect:
			var user *dbhelper.ImUser
			if len(m.searchResults) != 0 && m.cursor >= 0 && m.cursor < len(m.searchResults) {
				user = m.searchResults[m.cursor]
			}
			cmds = append(cmds, FetchSearchSelectedUserCmd(user))
			return m, tea.Batch(cmds...)
		case KeySearchUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case KeySearchDown:
			if m.cursor < len(m.searchResults)-1 {
				m.cursor++
			}
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			cmds = append(cmds, cmd)
			searchKey := strings.TrimSpace(m.searchInput.Value())
			if searchKey != "" {
				m.searching = true
				cmds = append(cmds, FetchSearchUserCmd(m.storage, searchKey))
			} else {
				m.searchResults = make([]*dbhelper.ImUser, 0)
				m.searching = false
			}
		}

	case SearchUserMsg:
		m.searching = false
		if msg.Err == nil {
			m.searchResults = msg.Users
		} else {
			logger.Errorf("搜索用户失败: %v", msg.Err)
			m.searchResults = make([]*dbhelper.ImUser, 0)
		}
	}

	return &m, tea.Batch(cmds...)
}

func (m SearchModel) View() string {
	var content strings.Builder

	title := lipgloss.NewStyle().
		Width(m.width).
		Height(2).
		Background(headerColor).
		Foreground(textColor).
		Bold(true).
		Align(lipgloss.Center).
		Render("搜索用户")
	content.WriteString(title + "\n")

	searchBox := m.renderSearchBox()
	content.WriteString(searchBox + "\n")

	separator := lipgloss.NewStyle().
		Width(m.width).
		Foreground(borderColor).
		Render(strings.Repeat("─", m.width))
	content.WriteString(separator + "\n")

	content.WriteString(m.renderSearchResults())

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(content.String())
}

func (m SearchModel) renderSearchBox() string {
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

func (m SearchModel) renderSearchResults() string {
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
				sep := lipgloss.NewStyle().
					Width(m.width - 4).
					Foreground(borderColor).
					Render(strings.Repeat("─", m.width-4))
				results.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(sep) + "\n")
			}
		}
	}
	return results.String()
}

func (m *SearchModel) UpdateSize(width, height int) {
	m.width = width
	m.height = height
}
