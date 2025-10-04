package transport

import (
	"context"
	"errors"
	"github.com/panjf2000/gnet/v2"
	"github.com/xuning888/helloIMClient/protocol"
	"github.com/xuning888/helloIMClient/transport/http"
	"sync"
	"time"
)

var (
	ErrEmptyServerUrl = errors.New("server url is required")
)

var (
	defaultHttpTimeout = time.Second * 3
)

type ImUser struct {
	UserId   int64  `json:"userId"`
	UserType int    `json:"userType"`
	Token    string `json:"token"`
}

type Result struct {
	conn *Conn
	resp protocol.Response
	err  error
}

type Dispatch func(result *Result)

type BaseInfo struct {
	User *ImUser
	// 长连接的主备地址
	IpList []string
}

type ImClient struct {
	mux          sync.Mutex   // 保护info
	Info         *BaseInfo    // Info 基础信息
	serverUrl    string       // IM 服务端WebApi地址
	ImHttpClient *http.Client // http 客户端
	cli          *gnet.Client
	sender       *Sender
	dispatch     Dispatch
}

func NewImClient(user *ImUser, dispatch Dispatch, opts ...Option) (*ImClient, error) {
	// 加载配置项
	options := loadOptions(opts...)
	if options.serverUrl == "" {
		return nil, ErrEmptyServerUrl
	}
	// 设置默认的http超时时间
	if options.httpTimeout == 0 {
		options.httpTimeout = defaultHttpTimeout
	}
	if options.maxAttempts == 0 {
		options.maxAttempts = 10
	}
	if options.lingerMs == 0 {
		options.lingerMs = time.Millisecond * 10
	}
	// 初始化imcli
	imCli, err := initImCli(user, dispatch, options)
	if err != nil {
		return nil, err
	}
	return imCli, nil
}

func (imCli *ImClient) WriteMessage(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	message, err := imCli.sender.writeMessage(ctx, request)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (imCli *ImClient) Close() {
	imCli.mux.Lock()
	defer imCli.mux.Unlock()
	imCli.sender.close()
}

func (imCli *ImClient) dial() Dial {
	return func(ctx context.Context) (*Conn, error) {
		imCli.mux.Lock()
		defer imCli.mux.Unlock()
		// 如果没有地址就拉一下最新的地址
		ips := imCli.Info.IpList
		if len(ips) == 0 {
			if err := imCli.fetchIpList(); err != nil {
				return nil, err
			}
		}
		ips = imCli.Info.IpList
		if len(ips) < 2 {
			// 只拉到一个地址
			return imCli.doDail(ips[0])
		}
		masterIP, slaveIP := ips[0], ips[1]
		// 先连接主, 连接不上就连接从
		conn, err := imCli.doDail(masterIP)
		if err == nil {
			return conn, nil
		}
		conn, err = imCli.doDail(slaveIP)
		if err == nil {
			return conn, nil
		}
		// 都无法建立连接
		imCli.Info.IpList = make([]string, 0)
		return nil, errors.New("")
	}
}

func (imCli *ImClient) doDail(addr string) (*Conn, error) {
	conn, err := imCli.cli.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	pconn := newConn(conn, imCli.dispatch)
	return pconn, nil
}

func (imCli *ImClient) auth(ctx context.Context, conn *Conn) error {
	err := conn.authReq(ctx, imCli.Info.User)
	if err != nil {
		return err
	}
	return nil
}

func (imCli *ImClient) fetchIpList() error {
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancelFunc()
	ips, err := imCli.ImHttpClient.IpList(timeout)
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		return errors.New("no available tcp gateway")
	}
	imCli.Info.IpList = ips
	return nil
}

func initImCli(user *ImUser, dispatch Dispatch, options *Options) (*ImClient, error) {
	imCli := &ImClient{
		Info: &BaseInfo{
			User: user,
		},
		dispatch: dispatch,
	}
	imCli.serverUrl = options.serverUrl
	// 初始化IM的http客户端
	imCli.ImHttpClient = http.NewClient(imCli.serverUrl, options.httpTimeout)
	if err := imCli.fetchIpList(); err != nil {
		return nil, err
	}
	// 构造transport
	transport := newTransport(imCli.dial(), imCli.auth)
	cli, err := gnet.NewClient(transport)
	if err != nil {
		return nil, err
	}
	imCli.cli = cli
	// 启动事件轮询
	err = imCli.cli.Start()
	if err != nil {
		return nil, err
	}
	// 构造发送器
	imCli.sender = newSender(transport, options.lingerMs, options.maxAttempts, options.initSeq)
	return imCli, nil
}
