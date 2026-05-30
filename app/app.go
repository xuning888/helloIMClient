package app

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xuning888/helloIMClient/im"
	"github.com/xuning888/helloIMClient/im/dbhelper"
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
	if err := i.sdk.ConnManager.Connect(ctx); err != nil {
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

	if err := i.sdk.ConnManager.Disconnect(ctx); err != nil {
		return nil
	}
	return nil
}

// registerEventCallbacks 将 SDK 事件转换为 TUI 命令
func (i *ImApp) registerEventCallbacks() {
	i.sdk.OnEvent(func(evt im.Event) {
		switch evt.Type {
		case im.EventMessageReceived:
			msgs, ok := evt.Data.([]*dbhelper.ChatMessage)
			if !ok || len(msgs) == 0 {
				return
			}
			if cmd := tui.FetchUpdatedChatListCmd(i.sdk); cmd != nil {
				i.program.Send(cmd())
			}
			if cmd := tui.FetchUpdateMessage(msgs[0].ChatID, msgs); cmd != nil {
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
