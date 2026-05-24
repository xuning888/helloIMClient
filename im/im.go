package im

import (
	"context"

	"github.com/xuning888/helloIMClient/conf"
	"github.com/xuning888/helloIMClient/internal/dal"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/http"
	"github.com/xuning888/helloIMClient/protocol"
	"github.com/xuning888/helloIMClient/transport"
)

// Client IM SDK 客户端，SDK 的唯一入口
type Client struct {
	addr   string
	opts   *Options
	store  *Store
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

	// 设置全局配置（过渡方案，后续将移除全局状态）
	conf.UserId = options.UID
	conf.ServerUrl = addr

	// 初始化 HTTP 客户端
	http.Init(addr, options.ConnectTimeout)

	// 初始化 SQLite
	if err := dal.Init(); err != nil {
		return nil, err
	}

	// 创建事件总线
	events := newCallbackRegistry()

	// 创建存储
	store := newStore()

	cli := &Client{
		addr:   addr,
		opts:   options,
		store:  store,
		events: events,
	}

	// 创建分发器
	dispatcher := newDispatcher(store, events)

	// 创建 transport
	tr, err := transport.NewImClient(sqllite.GetSeq, dispatcher.dispatch)
	if err != nil {
		return nil, err
	}

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

// SendMessage 发送上行消息并同步等待 ACK 响应
func (c *Client) SendMessage(ctx context.Context, req protocol.Message) (protocol.Message, error) {
	resp, err := c.connManager.transport.WriteMessage(ctx, req)
	if err != nil {
		return nil, err
	}
	c.events.fire(Event{Type: EventMessageSent, Data: resp})
	return resp, nil
}

// SendMessageWithSeq 发送单向消息（不等待 ACK）
func (c *Client) SendMessageWithSeq(ctx context.Context, seq int32, req protocol.Message) error {
	return c.connManager.transport.WriteMessageWithSeq(ctx, seq, req)
}

// Storage 获取存储管理器
func (c *Client) Storage() *Store {
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
