package app

import (
	"context"
	"github.com/xuning888/helloIMClient/option"
	"github.com/xuning888/helloIMClient/svc"
	"github.com/xuning888/helloIMClient/transport"
)

type ImApp struct {
	*router
}

func (i *ImApp) Start() error {
	if err := i.imCli.Start(); err != nil {
		return err
	}
	// 拉取用户信息
	users, err := i.imCli.ImHttpClient.Users(context.Background())
	if err != nil {
		return err
	}

	chats := make([]*svc.Chat, 0)
	for _, u := range users {
		//
		svc.NewChat()
	}
	i.commonSvc = svc.NewCommonSvc(users, nil)
	return nil
}

func NewApp(imUser *transport.ImUser, opts ...option.Option) (*ImApp, error) {
	imApp := &ImApp{
		router: &router{
			handlers: make(map[int32]Handler),
		},
	}
	imClient, err := transport.NewImClient(imUser, imApp.dispatch, opts...)
	if err != nil {
		return nil, err
	}
	imApp.imCli = imClient
	return imApp, nil
}
