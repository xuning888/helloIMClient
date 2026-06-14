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
	*MessageManager
	*ConnManager
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
	if err := dbhelper.Init(options.UID, addr); err != nil {
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
	mh := newMsgHandler(cli, events)

	// 创建 transport
	sender := transport.NewSender(dbhelper.GetSeq, mh.dispatch)
	tr := transport.NewConn(options.UID, &defaultAddrProvider{}, sender)

	// 创建子管理器
	cli.MessageManager = newMsgManager(cli)
	cli.ConnManager = newConnManager(tr, events)

	return cli, nil
}

// SendMessage 发送上行消息并同步等待 ACK
func (c *Client) SendMessage(ctx context.Context, msg protocol.Message) (protocol.Message, error) {
	ack, err := c.ConnManager.transport.Send(ctx, msg)
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

// LoadMessages 加载会话消息（初始打开会话时调用），返回已排序的消息列表
func (c *Client) LoadMessages(chat *dbhelper.ImChat) []*dbhelper.ChatMessage {
	cache := c.store.Messages.NewCache(chat)
	return cache.GetMessages()
}

// SaveSentMessage 保存已发送消息：入库 + 更新缓存 + 通知 UI
func (c *Client) SaveSentMessage(ctx context.Context, msg *dbhelper.ChatMessage) {
	if err := c.store.Messages.Save(ctx, msg); err != nil {
		c.events.fire(Event{Type: EventError, Data: err})
		return
	}
	c.store.Chats.UpdateVersion(ctx, msg.ChatID, msg.ChatType)

	sortedMessages := c.store.Messages.ProcessIncomingMessage(msg)
	c.MessageManager.pushNewMsg(sortedMessages)
}

// GetHistory 获取历史消息（游标分页），beforeSeq=0 表示获取最新消息
func (c *Client) GetHistory(ctx context.Context, chatID int64, chatType int32, beforeSeq int64, limit int) ([]*dbhelper.ChatMessage, error) {
	return c.store.Messages.GetHistory(ctx, chatID, chatType, beforeSeq, limit)
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
