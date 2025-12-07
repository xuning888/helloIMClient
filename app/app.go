package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/service"
	"github.com/xuning888/helloIMClient/transport"
	"github.com/xuning888/helloIMClient/tui"
)

type ImApp struct {
	user *sqllite.ImUser
	*router
}

func (i *ImApp) Start() error {
	if err := i.imCli.Start(); err != nil {
		return err
	}
	service.UpdateUsers()
	// 拉取用户信息
	program := tea.NewProgram(tui.InitMainModel(i.imCli), tea.WithAltScreen())
	i.program = program
	if _, err := program.Run(); err != nil {
		return err
	}
	i.imCli.Close()
	return nil
}

func NewApp() (*ImApp, error) {
	imApp := &ImApp{
		router: &router{
			handlers: make(map[int32]Handler),
		},
	}
	imClient, err := transport.NewImClient(imApp.dispatch)
	if err != nil {
		return nil, err
	}
	imApp.imCli = imClient
	return imApp, nil
}
