package app

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xuning888/helloIMClient/im"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/tui"
)

// ImApp 应用程序，负责连接 SDK 和 TUI
type ImApp struct {
	sdk     *im.Client
	program *tea.Program
}

// New 创建应用实例
func New(sdk *im.Client) *ImApp {
	return &ImApp{
		sdk: sdk,
	}
}

// Start 启动应用
func (i *ImApp) Start() error {
	ctx := context.Background()
	if err := i.sdk.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	// 拉取用户信息
	i.sdk.Storage().Users.Refresh(ctx)

	// 注册 SDK 事件回调，桥接到 TUI
	i.registerEventCallbacks()

	// 创建 Bubble Tea 程序
	program := tea.NewProgram(tui.InitMainModel(i.sdk), tea.WithAltScreen())
	i.program = program

	if _, err := program.Run(); err != nil {
		return err
	}

	i.sdk.Disconnect(ctx)
	return nil
}

// registerEventCallbacks 将 SDK 事件转换为 TUI 命令
func (i *ImApp) registerEventCallbacks() {
	i.sdk.OnEvent(func(evt im.Event) {
		switch evt.Type {
		case im.EventMessageReceived:
			msg, ok := evt.Data.(*sqllite.ChatMessage)
			if !ok {
				return
			}
			// 更新 TUI：执行 tea.Cmd 得到 tea.Msg 后发送
			if cmd := tui.FetchUpdatedChatListCmd(i.sdk); cmd != nil {
				i.program.Send(cmd())
			}
			if cmd := tui.FetchUpdateMessage(msg.ChatID, []*sqllite.ChatMessage{msg}); cmd != nil {
				i.program.Send(cmd())
			}
		case im.EventConnected:
			logger.Infof("app: SDK connected")

		case im.EventDisconnected:
			logger.Infof("app: SDK disconnected")

		case im.EventError:
			if err, ok := evt.Data.(error); ok {
				logger.Errorf("app: SDK error: %v", err)
			}
		}
	})
}
