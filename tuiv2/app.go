package tuiv2

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xuning888/helloIMClient/im"
	"github.com/xuning888/helloIMClient/im/dbhelper"
	"github.com/xuning888/helloIMClient/im/payload"
	"github.com/xuning888/helloIMClient/im/protocol/send"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

// Screen 视图路由
type Screen uint8

const (
	ScreenChatList Screen = iota
	ScreenChat
	ScreenSearch
)

// AppModel 顶层模型，负责屏幕路由和布局
type AppModel struct {
	sdk      *im.Client
	screen   Screen
	chatList ChatListModel
	chat     *ChatModel
	search   *SearchModel
	width    int
	height   int
}

func InitMainModel(sdk *im.Client) tea.Model {
	chatList := NewChatListModel(sdk.Storage())
	return &AppModel{
		sdk:      sdk,
		screen:   ScreenChatList,
		chatList: chatList,
	}
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// 全局消息：尺寸变化
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case tea.KeyMsg:
		if msg.String() == KeyCtrlC {
			return m, tea.Quit
		}

	// 路由消息：切换视图
	case SelectChatMsg:
		if m.chat == nil || m.chat.cache.GetChat().ChatId != msg.Chat.ChatId {
			if m.chat != nil {
				m.chat.cache.Destroy()
			}
			storage := m.sdk.Storage()
			m.chat = NewChatModel(msg.Chat, storage, m.sdk.GetUID(), m.makeSendFn())
		}
		m.screen = ScreenChat
		m.updateLayout()

	case BackToListMsg:
		if m.chat != nil {
			m.chat.cache.Destroy()
			m.chat = nil
		}
		m.screen = ScreenChatList

	case StartSearchMsg:
		m.search = NewSearchModel(m.sdk.Storage())
		m.screen = ScreenSearch
		m.updateLayout()

	case ExitSearchMsg:
		m.search = nil
		m.screen = ScreenChatList

	case SearchSelectedUserMsg:
		user := msg.User
		if user != nil {
			chat, err := m.sdk.Storage().Chats.GetOrCreate(context.Background(), user.UserID, 1)
			if err != nil {
				logger.Errorf("创建聊天会话失败: %v", err)
				return m, nil
			}
			m.chat = NewChatModel(chat, m.sdk.Storage(), m.sdk.GetUID(), m.makeSendFn())
			m.screen = ScreenChat
			m.search = nil
		}
		logger.Infof("触发搜索结果事件, user: %v", user)
		cmds = append(cmds, FetchUpdatedChatListCmd(m.sdk.Storage()))
		m.updateLayout()
	}

	// 屏幕级路由：只转发到当前活动的视图
	switch m.screen {
	case ScreenChatList:
		updatedList, listCmd := m.chatList.Update(msg)
		m.chatList = updatedList.(ChatListModel)
		if listCmd != nil {
			cmds = append(cmds, listCmd)
		}
		// ChatListUpdatedMsg 即使在其他屏幕也要处理
		if _, ok := msg.(ChatListUpdatedMsg); !ok {
			if m.chat != nil {
				updatedChat, chatCmd := m.chat.Update(msg)
				if cm, ok := updatedChat.(*ChatModel); ok {
					m.chat = cm
				}
				if chatCmd != nil {
					cmds = append(cmds, chatCmd)
				}
			}
		}

	case ScreenChat:
		if m.chat != nil {
			updatedChat, chatCmd := m.chat.Update(msg)
			if cm, ok := updatedChat.(*ChatModel); ok {
				m.chat = cm
			}
			if chatCmd != nil {
				cmds = append(cmds, chatCmd)
			}
		}
		// 同时更新聊天列表（后台刷新）
		updatedList, listCmd := m.chatList.Update(msg)
		m.chatList = updatedList.(ChatListModel)
		if listCmd != nil {
			cmds = append(cmds, listCmd)
		}

	case ScreenSearch:
		if m.search != nil {
			updatedSearch, searchCmd := m.search.Update(msg)
			if sm, ok := updatedSearch.(*SearchModel); ok {
				m.search = sm
			}
			if searchCmd != nil {
				cmds = append(cmds, searchCmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// 搜索页面全屏
	if m.screen == ScreenSearch && m.search != nil {
		return m.search.View()
	}

	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth

	leftPanel := m.chatList.View()
	leftPanel = lipgloss.NewStyle().
		Width(leftWidth).
		Height(m.height).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(borderColor).
		Render(leftPanel)

	var rightPanel string
	if m.screen == ScreenChat && m.chat != nil {
		rightPanel = m.chat.View()
	} else {
		rightPanel = lipgloss.Place(rightWidth, m.height, lipgloss.Center, lipgloss.Center,
			"选择一个聊天开始对话")
	}

	rightPanel = lipgloss.NewStyle().
		Width(rightWidth).
		Height(m.height).
		Render(rightPanel)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	statusBar := m.statusBarView()

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

func (m AppModel) statusBarView() string {
	var hint string
	switch m.screen {
	case ScreenChatList:
		hint = "list: ↑↓ 选择 • Enter 打开 • F3 搜索 • ctrl+c 退出"
	case ScreenChat:
		hint = "chat: Enter 发送 • Tab 切换 • Esc 返回"
	case ScreenSearch:
		hint = "search: ↑↓ 选择 • Enter 创建会话 • Esc 返回"
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(1).
		Background(backgroundColor).
		Foreground(subtextColor).
		Align(lipgloss.Center).
		Render(hint)
}

func (m *AppModel) updateLayout() {
	if m.chat != nil {
		m.chat.UpdateSize(m.width/3*2, m.height-1)
	}
	if m.search != nil {
		m.search.UpdateSize(m.width/3*2, m.height-1)
	}
	m.chatList.UpdateSize(m.width/3, m.height-1)
}

// makeSendFn 构建发送消息的闭包，供 ChatModel 使用
func (m *AppModel) makeSendFn() SendMessageFunc {
	return func(ctx context.Context, chatID int64, chatType int32, content string) (*dbhelper.ChatMessage, error) {
		uid := m.sdk.GetUID()
		p := payload.NewTextMessage(content, false, nil)
		request := send.NewSendMsg(uid, chatID, chatType, p, 0, 0)

		ack, err := m.sdk.SendMessage(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("send message: %w", err)
		}

		sendAck, ok := ack.(*send.SendAck)
		if !ok {
			return nil, fmt.Errorf("unexpected ack type")
		}

		contentStr, contentType := payload.ExtractContent(request.Payload)
		msg := dbhelper.NewMessage(chatType, chatID, sendAck.MsgId(), uid, chatID,
			request.FromUserType, request.ToUserType, sendAck.MsgSeq(),
			contentStr, contentType, request.CmdId(),
			request.SendTimestamp, 0, sendAck.ServerSeq())

		if err := m.sdk.Storage().Messages.Save(ctx, msg); err != nil {
			logger.Errorf("saveSentMessage error: %v", err)
		}
		m.sdk.Storage().Chats.UpdateVersion(ctx, chatID, chatType)
		return msg, nil
	}
}
