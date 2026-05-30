package im

import (
	"context"

	"github.com/xuning888/helloIMClient/im/dbhelper"
	"github.com/xuning888/helloIMClient/im/http"
	"github.com/xuning888/helloIMClient/im/protocol"
	"github.com/xuning888/helloIMClient/im/store"
	"github.com/xuning888/helloIMClient/im/transport"
)

// Client IM SDK 客户端，SDK 的唯一入口
type Client struct {
	addr   string
	opts   *Options
	store  *store.Store
	events *callbackRegistry
	*msgManager
	*connManager
}

// New 创建 SDK 客户端实例
func New(addr string, opts ...Option) (*Client, error) {
	options := NewOptions()
	for _, o := range opts {
		o(options)
	}

	// 初始化 HTTP 客户端
	http.Init(addr, options.ConnectTimeout)

	// 初始化 SQLite
	if err := dbhelper.Init(options.UID); err != nil {
		return nil, err
	}

	// 创建事件总线
	events := newCallbackRegistry()

	// 创建存储
	storage := store.NewStore(options.UID)
	cli := &Client{
		addr:   addr,
		opts:   options,
		events: events,
		store:  storage,
	}
	// 创建分发器
	dispatcher := newDispatcher(storage, events)

	// 创建 transport
	tr := transport.NewClient(options.UID, dispatcher.dispatch, &defaultAddrProvider{}, dbhelper.GetSeq)

	// 创建子管理器
	cli.msgManager = newMsgManager(cli)
	cli.connManager = newConnManager(tr, events)

	return cli, nil
}

// Connect 建立连接
func (c *Client) Connect(ctx context.Context) error {
	return c.connManager.Connect(ctx)
}

// Disconnect 断开连接
func (c *Client) Disconnect(ctx context.Context) error {
	return c.connManager.Disconnect(ctx)
}

// State 获取当前连接状态
func (c *Client) State() ConnState {
	return c.connManager.State()
}

// SendMessage 发送上行消息并同步等待 ACK
func (c *Client) SendMessage(ctx context.Context, msg protocol.Message) (protocol.Message, error) {
	ack, err := c.connManager.transport.Send(ctx, msg)
	if err != nil {
		return nil, err
	}
	c.events.fire(Event{Type: EventMessageSent, Data: ack})
	return ack, nil
}

// Storage 获取存储管理器
func (c *Client) Storage() *store.Store {
	return c.store
}

// OnEvent 注册事件回调，返回取消订阅函数
func (c *Client) OnEvent(cb EventCallback) func() {
	return c.events.subscribe(cb)
}

// GetUID 获取当前用户 ID
func (c *Client) GetUID() int64 {
	return c.opts.UID
}

// defaultAddrProvider 默认地址提供者
type defaultAddrProvider struct{}

func (p *defaultAddrProvider) GetAddr(ctx context.Context) ([]string, error) {
	return http.IpList(ctx)
}
