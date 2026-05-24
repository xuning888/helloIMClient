package im

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xuning888/helloIMClient/im/transport"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

// ConnState SDK 层连接状态
type ConnState int32

const (
	StateDisconnected ConnState = iota
	StateConnecting
	StateConnected
	StateDisconnecting
)

type connManager struct {
	transport *transport.Client
	events    *callbackRegistry
	state     atomic.Int32
	closeOnce sync.Once
}

func newConnManager(tr *transport.Client, events *callbackRegistry) *connManager {
	return &connManager{
		transport: tr,
		events:    events,
	}
}

func (c *connManager) Connect(ctx context.Context) error {
	c.state.Store(int32(StateConnecting))
	c.events.fire(Event{Type: EventConnecting})

	if err := c.transport.Connect(ctx); err != nil {
		c.state.Store(int32(StateDisconnected))
		c.events.fire(Event{Type: EventDisconnected})
		return err
	}

	c.state.Store(int32(StateConnected))
	c.events.fire(Event{Type: EventConnected})
	return nil
}

func (c *connManager) Disconnect(ctx context.Context) error {
	var err error
	c.closeOnce.Do(func() {
		c.state.Store(int32(StateDisconnecting))
		done := make(chan struct{})
		go func() {
			c.transport.Close()
			close(done)
		}()
		select {
		case <-done:
		case <-ctx.Done():
			err = ctx.Err()
		case <-time.After(5 * time.Second):
			logger.Errorf("connManager: disconnect timeout")
		}
		c.state.Store(int32(StateDisconnected))
		c.events.fire(Event{Type: EventDisconnected})
	})
	return err
}

func (c *connManager) State() ConnState {
	return ConnState(c.state.Load())
}
