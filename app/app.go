package app

import (
	"github.com/xuning888/helloIMClient/option"
	"github.com/xuning888/helloIMClient/transport"
)

type ImApp struct {
	*router
}

func (i *ImApp) Start() error {
	return i.imCli.Start()
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
