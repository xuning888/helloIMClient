package transport

import (
	"context"
	"github.com/panjf2000/gnet/v2"
	"sync"
	"sync/atomic"
	"time"
)

type Auth func(ctx context.Context, conn *Conn) error

type Dial func(ctx context.Context) (*Conn, error)

// Transport TCP 网络传输的实现细节
type Transport struct {
	gnet.BuiltinEventEngine              // 实现gnet的eventHandler, 事件回调
	connManager             *connManager // 连接管理器
	closed                  int32
	reconn                  chan struct{}
}

func (t *Transport) roundTrip(ctx context.Context, item *syncItem) error {
	if atomic.LoadInt32(&t.closed) == 1 {
		return ErrClosed
	}
	// 获取连接
	conn, err := t.connManager.Dial()
	if err != nil {
		return err
	}
	if !conn.Authed() {
		if err2 := t.connManager.auth(ctx, conn); err2 != nil {
			return err2
		}
	}
	// 发送数据
	if err = conn.asyncWrite(item); err != nil {
		return err
	}
	// 同步等待100ms
	timeout, cancelFunc := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancelFunc()
	err = item.await(timeout)
	if err != nil {
		return err
	}
	return nil
}

func (t *Transport) OnClose(conn gnet.Conn, err error) gnet.Action {
	t.connManager.delete(conn)
	return gnet.None
}

// OnTick 定时任务
func (t *Transport) OnTick() (delay time.Duration, action gnet.Action) {
	return time.Second * 10, gnet.None
}

func (t *Transport) OnTraffic(c gnet.Conn) (action gnet.Action) {
	pconn := t.connManager.get(c)
	if pconn == nil {
		return gnet.Close
	}
	return pconn.Decode()
}

type connManager struct {
	mux   sync.Mutex
	dial  Dial
	auth  Auth
	conns sync.Map
}

func (cm *connManager) Dial() (*Conn, error) {
	cm.mux.Lock()
	defer cm.mux.Unlock()
	var conn *Conn = nil
	cm.conns.Range(func(key, value any) bool {
		if c, ok := value.(*Conn); ok {
			conn = c
			return false
		}
		return true
	})
	if conn != nil {
		return conn, nil
	}
	ctx := context.Background()
	conn, err := cm.dial(ctx)
	if err != nil {
		return nil, err
	}
	cm.conns.Store(conn.conn, conn)
	return conn, nil
}

func (cm *connManager) get(conn gnet.Conn) *Conn {
	cm.mux.Lock()
	defer cm.mux.Unlock()
	if value, ok := cm.conns.Load(conn); ok {
		return value.(*Conn)
	}
	return nil
}

func (cm *connManager) delete(c gnet.Conn) {
	cm.mux.Lock()
	defer cm.mux.Unlock()
	if value, loaded := cm.conns.LoadAndDelete(c); loaded {
		if conn, ok := value.(*Conn); ok {
			conn.Close()
		}
	}
}

func newTransport(dial Dial, auth Auth) *Transport {
	t := &Transport{
		connManager: &connManager{
			mux:  sync.Mutex{},
			dial: dial,
			auth: auth,
		},
		reconn: make(chan struct{}, 1),
	}
	return t
}
