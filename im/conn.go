package im

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/antlabs/timer"
	"github.com/panjf2000/gnet/v2"
	"github.com/xuning888/helloIMClient/internal/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/protocol"
	"github.com/xuning888/helloIMClient/protocol/auth"
)

var (
	maxReconnectTimes  = 10
	baseReconnectDelay = 500
)

type GetSeq func() int32

type HiConn struct {
	log logger.Logger

	// 连接状态
	isReconnecting atomic.Bool // 是否在重连中
	isConnecting   atomic.Bool // 是否正在连接
	isClosing      atomic.Bool // 连接是否正在关闭
	connCount      int         // 重连次数

	// 网络配置
	address string // 连接地址

	// 事件驱动和网络相关对象
	cli          *gnet.Client  // gnet客户端
	eventHandler *eventHandler // gnet事件驱动
	conn         gnet.Conn     // gnet长连接
	ms           *MsgSender    // 发送器

	// 控制参数
	ctx      context.Context
	cancel   context.CancelFunc
	connLock sync.RWMutex
	tm       timer.Timer

	// 获取客户端序号
	getSeq GetSeq
}

func NewHiConn(ms *MsgSender, getSeq GetSeq) *HiConn {
	ctx, cancelFunc := context.WithCancel(context.Background())
	hc := &HiConn{
		log:            logger.Named("hiconn"),
		isReconnecting: atomic.Bool{},
		isConnecting:   atomic.Bool{},
		isClosing:      atomic.Bool{},
		connCount:      0,
		address:        "",
		cli:            nil,
		eventHandler:   nil,
		conn:           nil,
		ms:             ms,
		ctx:            ctx,
		cancel:         cancelFunc,
		connLock:       sync.RWMutex{},
		tm:             timer.NewTimer(),
		getSeq:         getSeq,
	}
	go hc.fetchConnAddress()
	return hc
}

// Connect 建立连接
func (hc *HiConn) Connect() {
	if !hc.connIsNil() {
		return
	}
	// 检查连接状态, 如果连接建立中就返回
	if !hc.isConnecting.CompareAndSwap(false, true) {
		hc.log.Warnf("连接已经建立, 忽略重复连接的请求")
		return
	}
	defer hc.isConnecting.Store(false)

	// 关闭现有的连接
	hc.CloseConnect()

	// 开始建立连接
	go hc.connectAsync(hc.address)
}

func (hc *HiConn) connectAsync(addr string) {
	ctx, cancel := context.WithTimeout(hc.ctx, 5*time.Second)
	defer cancel()

	connChan := make(chan gnet.Conn, 1)
	errChan := make(chan error, 1)

	// 创建新的gnet客户端
	cli, err := gnet.NewClient(newEventHandler(func(conn gnet.Conn) {
		connChan <- conn
	}))
	if err != nil {
		hc.log.Errorf("创建gnet客户端失败, error: %v", err)
		hc.forceReconnect() // 强制重连
		return
	}
	go func() {
		// 启动客户端，其实就是启动事件驱动器
		if err := cli.Start(); err != nil {
			errChan <- err
			return
		}
		// 建立tcp连接
		if _, err := cli.Dial("tcp", addr); err != nil {
			errChan <- err
			return
		}
	}()

	select {
	case conn := <-connChan:
		hc.connLock.Lock()
		hc.conn = conn
		hc.cli = cli
		conn.SetContext(hc) // 关联起来
		hc.connLock.Unlock()
		hc.log.Infof("长连接建立成功")
		// 发送认证消息与服务端建立session
		if err := hc.sendAuthMessage(); err != nil {
			// 认证失败, 关闭连接, 然后重连
			hc.CloseConnect()
			hc.forceReconnect()
			return
		}
		hc.isConnecting.Store(false)
	case err := <-errChan:
		hc.log.Errorf("连接建立失败, error: %v", err)
		if err := cli.Stop(); err != nil {
			// 停止事件驱动器
			hc.log.Errorf("关闭事件驱动异常, error: %v", err)
		}
		hc.forceReconnect() // 强制重连
	case <-ctx.Done():
		hc.log.Errorf("连接建立超时")
		if err := cli.Stop(); err != nil {
			// 停止事件驱动器
			hc.log.Errorf("关闭事件驱动异常, error: %v", err)
		}
		hc.forceReconnect()
	}
}

func (hc *HiConn) CloseConnect() {
	if !hc.isClosing.CompareAndSwap(false, true) {
		hc.log.Infof("关闭连接进行中")
		return
	}
	defer hc.isClosing.Store(false)

	hc.connLock.Lock()
	if hc.conn == nil {
		hc.connLock.Unlock()
		hc.log.Infof("连接已经为空")
		return
	}
	conn, cli := hc.conn, hc.cli
	hc.conn = nil
	hc.cli = nil
	hc.connLock.Unlock()
	go func() {
		timeout, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancelFunc()
		closeChan := make(chan error, 1)
		go func() {
			if conn != nil {
				conn.SetContext(nil) // 清空关联关系
				if err := conn.Close(); err != nil {
					closeChan <- err
				}
			}
			if cli != nil {
				if err := cli.Stop(); err != nil {
					closeChan <- err
				}
			}
		}()
		select {
		case <-timeout.Done():
			hc.log.Errorf("连接关闭超时, conn: %v", conn.RemoteAddr())
		case err := <-closeChan:
			if err != nil {
				hc.log.Errorf("连接关闭异常, error: %v", err)
			} else {
				hc.log.Infof("连接关闭成功: %v", conn.RemoteAddr())
			}
		}
	}()
}

func (hc *HiConn) forceReconnect() {
	// 连接正在重连中, 就退出
	if !hc.isReconnecting.CompareAndSwap(false, true) {
		hc.log.Warnf("连接正在重连中.....")
		return
	}
	defer hc.isReconnecting.Store(false)

	hc.connCount++
	if hc.connCount > maxReconnectTimes {
		hc.log.Errorf("达到最大连接次数, 停止重连")
		return
	}
	delay := baseReconnectDelay * (1 << uint(hc.connCount-1))
	if delay > 5000 {
		delay = 5000
	}
	hc.log.Infof("重连延迟: %dms", delay)

	// 异步执行
	go func() {
		select {
		case <-time.After(time.Millisecond * time.Duration(delay)):
			hc.reconnection()
			return
		case <-hc.ctx.Done():
			return
		}
	}()
}

func (hc *HiConn) reconnection() {
	if hc.isClosing.Load() {
		hc.log.Infof("等待连接关闭完成后再重连")
		// 500ms之后执行重连
		hc.tm.AfterFunc(time.Millisecond*500, hc.reconnection)
		return
	}
	// 把地址清空, 重新连接到原因可能就是地址不可用了
	hc.address = ""
	// 获取远程地址, 然后重新建立连接
	hc.fetchConnAddress()
	if hc.connIsNil() {
		hc.Connect()
	}
}

func (hc *HiConn) fetchConnAddress() {
	// 从远程获取地址
	timeout, cancelFunc := context.WithTimeout(hc.ctx, 5*time.Second)
	defer cancelFunc()
	type Result struct {
		ipList []string
		err    error
	}
	addrChan := make(chan *Result, 1)
	go func() {
		ipList, err := http.IpList(timeout)
		addrChan <- &Result{
			ipList: ipList, err: err,
		}
	}()

	select {
	case res := <-addrChan:
		if res.err != nil {
			hc.log.Errorf("获取无效的连接地址, error: %v", res.err)
			hc.isReconnecting.Store(false)
			hc.forceReconnect()
			return
		}
		address := hc.decodeIpList(res.ipList)
		if address == "" {
			hc.log.Errorf("获取无效的连接地址")
			hc.isReconnecting.Store(false)
			hc.forceReconnect()
			return
		}
		hc.address = address
	case <-timeout.Done():
		hc.log.Errorf("获取连接地址超时")
		hc.isReconnecting.Store(false)
		hc.forceReconnect()
	}
}

func (hc *HiConn) decodeIpList(ipList []string) string {
	var n = len(ipList)
	if n < 1 {
		return ""
	}
	if n == 2 {
		m, s := ipList[0], ipList[1]
		hc.log.Infof("获取到远程地址, m: %v, s: %v", m, s)
		return m
	}
	return ipList[0]
}

func (hc *HiConn) connIsNil() bool {
	hc.connLock.RLock()
	defer hc.connLock.RUnlock()
	return hc.conn == nil
}

func (hc *HiConn) getConn() gnet.Conn {
	hc.connLock.RLock()
	defer hc.connLock.RUnlock()
	return hc.conn
}

func (hc *HiConn) sendAuthMessage() error {
	request := auth.NewRequest(1, 0, "")
	response, err := hc.SyncRequest(context.Background(), request)
	if err != nil {
		return err
	}
	if resp, ok := response.(*auth.Response); ok {
		hc.log.Infof("auth auth responose: %v", resp)
		return nil
	} else {
		err := fmt.Errorf("发送认证消息, 未知的response: %v", resp)
		hc.log.Errorf("error: %v", err)
		return err
	}
}

func (hc *HiConn) SyncRequest(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	seq := hc.getSeq()
	frame, err := protocol.EncodeMessageToFrame(seq, protocol.REQ, request)
	if err != nil {
		hc.log.Errorf("auth消息序列化失败, error: %v", err)
		return nil, err
	}
	resp, err := hc.Sync(ctx, frame)
	if err != nil {
		return nil, err
	}
	decodeResp, err := protocol.DecodeResp(resp)
	if err != nil {
		return nil, err
	}
	return decodeResp, nil
}

func (hc *HiConn) Sync(ctx context.Context, request *protocol.Frame) (*protocol.Frame, error) {
	resp, err := hc.ms.syncWithRetry(ctx, hc, request, time.Millisecond*200, 3)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (hc *HiConn) OneWay(ctx context.Context, request *protocol.Frame) error {
	if err := hc.ms.oneway(ctx, hc, request); err != nil {
		return err
	}
	return nil
}

func (hc *HiConn) onMessage() gnet.Action {
	return hc.ms.onMessage(hc)
}

type asyncWrite func(err error)

func (hc *HiConn) asyncMessage(msg []byte, callback asyncWrite) error {
	err := hc.conn.AsyncWrite(msg, func(c gnet.Conn, err error) error {
		callback(err)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
