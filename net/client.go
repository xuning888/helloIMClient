package net

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/xuning888/helloIMClient/frame"
	pb "github.com/xuning888/helloIMClient/proto"
)

var (
	Seqs = atomic.Int32{}
)

func GetSeq() int32 {
	seq := Seqs.Add(1)
	if seq < 0 {
		Seqs.Store(0)
		seq = Seqs.Add(1)
	}
	return seq
}

type DownMsgHandler func(reply *frame.Frame)

var (
	ErrorsDupCli     = fmt.Errorf("不能重复建立连接")
	ErrorsAuthFailed = fmt.Errorf("长连接认证失败")
)

type CliState int

const (
	Init       CliState = 0 // 初始化
	Connecting CliState = 1 // 建立TCP连接中
	Connected  CliState = 2 // 建立TCP成功
	Authing    CliState = 3 // 长连接认证中
	Authed     CliState = 4 // 长连接认证成功
	Closed     CliState = 5 // cli关闭
)

type GateUser struct {
	Uid      string
	UserType int32
}

type ImClient struct {
	gnet.BuiltinEventEngine
	cli       *gnet.Client      // gnet客户端
	conn      gnet.Conn         // gnet长连接
	replies   chan *frame.Frame // 响应缓冲区
	stop      chan struct{}     // 传递stop信号
	inflightQ *inflightQueue    // 上行消息飞行队列
	handle    DownMsgHandler    // 业务处理器
	state     CliState          // cli的状态
	lock      sync.RWMutex      // 不可重入锁，保护临界资源
}

func (imCli *ImClient) Connect(addr string, user *GateUser) error {

	// 连接不是初始化过程, 查看状态, 使用读锁
	imCli.lock.RLock()
	if imCli.state != Init && imCli.state != Closed {
		return ErrorsDupCli
	}
	imCli.lock.RUnlock()

	// 向gnet注册handler, 写状态, 使用写锁
	imCli.lock.Lock()
	cli, err := gnet.NewClient(imCli, gnet.WithMulticore(true))
	if err != nil {
		imCli.state = Closed
		return err
	}
	imCli.cli = cli
	// 启动gnet
	err = cli.Start()
	if err != nil {
		imCli.state = Closed
		return err
	}

	imCli.state = Connecting // 连接建立中
	conn, err := cli.Dial("tcp", addr)
	if err != nil {
		imCli.state = Closed
		return err
	}
	imCli.state = Connected // 连接建立成功
	imCli.conn = conn
	imCli.lock.Unlock()

	// 开始认证
	imCli.state = Authing // 连接认证中
	authRequst := &pb.AuthRequest{
		Uid:      user.Uid,
		UserType: user.UserType,
	}
	if err2 := imCli.auth(authRequst); err2 != nil {
		log.Printf("认证失败, error: %v", err2)
		imCli.state = Closed
		return err2
	}
	imCli.state = Authed // 认证成功
	return nil
}

func (imCli *ImClient) Close() error {
	imCli.lock.Lock()
	defer imCli.lock.Unlock()
	if imCli.state == Closed {
		return nil
	}
	imCli.state = Closed
	close(imCli.stop)
	if imCli.conn != nil {
		imCli.conn.Close()
	}
	if imCli.cli != nil {
		imCli.cli.Stop()
	}
	close(imCli.replies)
	return nil
}

func NewImClient(downMsgHandler DownMsgHandler) *ImClient {
	imcli := &ImClient{
		replies: make(chan *frame.Frame, 100), // 接受消息的缓冲区
		handle:  downMsgHandler,               // 业务处理器
		stop:    make(chan struct{}),          // stopChan, 用于发送stop信号
		state:   Init,                         // 初始化
		lock:    sync.RWMutex{},               // 搞个锁
	}
	imcli.inflightQ = makeInflightQueue(time.Minute*2, 3, time.Second*5, imcli.sendFrame)
	go imcli.process() // 开始处理消息
	return imcli
}
