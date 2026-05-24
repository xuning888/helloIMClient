package transport

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/xuning888/helloIMClient/conf"
	protocol2 "github.com/xuning888/helloIMClient/im/protocol"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

var (
	ErrClosed = errors.New("transport client closed")

	maxReconnectAttempts = 10
	baseReconnectDelay   = 500 // ms
)

// ConnState 连接状态
type ConnState int32

const (
	StateDisconnected ConnState = iota
	StateConnecting
	StateConnected
)

// AddrProvider 地址提供者（由上层注入）
type AddrProvider interface {
	GetAddr(ctx context.Context) ([]string, error)
}

// Client 传输层客户端，管理连接生命周期 + 消息收发
type Client struct {
	gnet.BuiltinEventEngine

	log logger.Logger

	// 连接
	gnetCli *gnet.Client
	conn    gnet.Conn
	connMu  sync.RWMutex

	// 状态
	state      atomic.Int32
	reconnect  atomic.Bool
	connecting atomic.Bool
	closing    atomic.Bool
	attempt    atomic.Int32
	closeOnce  sync.Once
	closed     atomic.Int32

	// 子组件
	sender   *sender
	dispatch func(protocol2.Message)

	// 地址
	addrProvider AddrProvider
	address      string

	// 生命周期
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient 创建传输客户端
func NewClient(dispatch func(protocol2.Message), addrProvider AddrProvider, getSeq GetSeq) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		log:          logger.Named("transport"),
		state:        atomic.Int32{},
		addrProvider: addrProvider,
		dispatch:     dispatch,
		ctx:          ctx,
		cancel:       cancel,
	}
	c.state.Store(int32(StateDisconnected))
	c.sender = newSender(getSeq, dispatch)
	return c
}

// Connect 建立连接
func (c *Client) Connect(ctx context.Context) error {
	if !c.connIsNil() {
		return nil
	}
	if !c.connecting.CompareAndSwap(false, true) {
		return nil
	}
	defer c.connecting.Store(false)

	c.setState(StateConnecting)
	c.closeConn()

	if c.address == "" {
		if err := c.fetchAddr(ctx); err != nil {
			c.setState(StateDisconnected)
			return err
		}
	}

	return c.dial(ctx, c.address)
}

// Send 发送消息
func (c *Client) Send(ctx context.Context, msg protocol2.Message) (protocol2.Message, error) {
	if c.State() != StateConnected {
		return nil, errors.New("transport: not connected")
	}
	conn := c.getConn()
	if conn == nil {
		return nil, errors.New("transport: no connection")
	}
	return c.sender.sendWithRetry(ctx, conn, msg, 200*time.Millisecond, 3)
}

// Close 关闭客户端
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.closed.Store(1)
		c.cancel()
		c.closeConn()
		c.sender.close()
	})
}

// State 当前连接状态
func (c *Client) State() ConnState {
	return ConnState(c.state.Load())
}

// ---- gnet.EventHandler ----

func (c *Client) OnTraffic(gconn gnet.Conn) gnet.Action {
	for {
		frame, _, action := readFrame(gconn)
		if frame == nil {
			return action
		}
		if frame.Header.Req == protocol2.RES {
			// ACK 响应：完成 sender 中的 promise
			c.sender.complete(frame)
		} else {
			// 推送消息：交接给 dispatch goroutine
			c.sender.dispatchFrame(frame, gconn)
		}
	}
}

func (c *Client) OnTick() (delay time.Duration, action gnet.Action) {
	if c.State() == StateConnected {
		conn := c.getConn()
		if conn != nil {
			if err := sendPing(conn); err != nil {
				c.log.Errorf("heartbeat error: %v", err)
			}
		}
	}
	return 10 * time.Second, gnet.None
}

func (c *Client) OnClose(gconn gnet.Conn, err error) gnet.Action {
	c.connMu.Lock()
	if c.conn == gconn {
		c.conn = nil
	}
	c.connMu.Unlock()

	// 主动关闭或正在关闭中，不触发重连
	if c.closed.Load() == 1 || c.closing.Load() {
		return gnet.None
	}
	c.forceReconnect()
	return gnet.None
}

// ---- 内部方法 ----

func (c *Client) dial(ctx context.Context, addr string) error {
	cli, err := gnet.NewClient(c, gnet.WithTicker(true),
		gnet.WithLogger(logger.Named("gnet")))
	if err != nil {
		return err
	}
	if err := cli.Start(); err != nil {
		cli.Stop()
		return err
	}

	type dialResult struct {
		conn gnet.Conn
		err  error
	}
	resultCh := make(chan dialResult, 1)
	go func() {
		conn, err := cli.Dial("tcp", addr)
		resultCh <- dialResult{conn: conn, err: err}
	}()

	var result dialResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		cli.Stop()
		return ctx.Err()
	}

	if result.err != nil {
		cli.Stop()
		return result.err
	}

	c.setConn(result.conn)
	c.gnetCli = cli

	// 认证
	if err := c.auth(ctx); err != nil {
		c.closeConn()
		// 尝试备用地址
		if slave := c.trySlave(); slave != "" {
			c.log.Infof("master failed, trying slave: %s", slave)
			return c.dial(ctx, slave)
		}
		c.forceReconnect()
		return err
	}

	c.setState(StateConnected)
	c.attempt.Store(0)
	c.log.Infof("connected to %s", addr)
	return nil
}

func (c *Client) trySlave() string {
	ips, err := c.addrProvider.GetAddr(context.Background())
	if err != nil || len(ips) < 2 {
		return ""
	}
	return ips[1]
}

func (c *Client) auth(ctx context.Context) error {
	msg := NewAuthRequest(conf.UserId, 0, "")
	conn := c.getConn()
	if conn == nil {
		return errors.New("no connection for auth")
	}
	resp, err := c.sender.send(ctx, conn, msg, 5*time.Second)
	if err != nil {
		return err
	}
	if authResp, ok := resp.(*AuthResponse); ok && authResp.AuthResponse.Success {
		return nil
	}
	return errors.New("auth failed")
}

func (c *Client) forceReconnect() {
	if c.closed.Load() == 1 {
		return
	}
	if !c.reconnect.CompareAndSwap(false, true) {
		return
	}
	defer c.reconnect.Store(false)

	attempt := int(c.attempt.Add(1))
	if attempt > maxReconnectAttempts {
		c.setState(StateDisconnected)
		c.log.Errorf("max reconnect attempts reached")
		return
	}

	delay := baseReconnectDelay * (1 << uint(attempt-1))
	if delay > 5000 {
		delay = 5000
	}
	c.log.Infof("reconnect attempt %d, delay %dms", attempt, delay)

	time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		c.address = ""
		c.setState(StateDisconnected)
		if err := c.fetchAddr(context.Background()); err != nil {
			c.log.Errorf("reconnect fetchAddr failed: %v", err)
			c.forceReconnect() // 取地址失败也重试
			return
		}
		if err := c.Connect(context.Background()); err != nil {
			c.log.Errorf("reconnect Connect failed: %v", err)
		}
	})
}

func (c *Client) fetchAddr(ctx context.Context) error {
	ips, err := c.addrProvider.GetAddr(ctx)
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		return errors.New("no available address")
	}
	c.address = ips[0]
	if len(ips) > 1 {
		c.log.Infof("got addresses: master=%s, slave=%s", ips[0], ips[1])
		c.address = ips[0]
	}
	return nil
}

func (c *Client) closeConn() {
	if !c.closing.CompareAndSwap(false, true) {
		return
	}
	defer c.closing.Store(false)

	c.connMu.Lock()
	conn := c.conn
	cli := c.gnetCli
	c.conn = nil
	c.gnetCli = nil
	c.connMu.Unlock()

	if conn != nil {
		conn.Close()
	}
	if cli != nil {
		cli.Stop()
	}
}

func (c *Client) setConn(conn gnet.Conn) {
	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()
}

func (c *Client) getConn() gnet.Conn {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.conn
}

func (c *Client) connIsNil() bool {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.conn == nil
}

func (c *Client) setState(state ConnState) {
	c.state.Store(int32(state))
}
