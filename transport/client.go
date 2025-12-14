package transport

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/xuning888/helloIMClient/internal/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/protocol"
	"github.com/xuning888/helloIMClient/protocol/echo"
	"github.com/xuning888/helloIMClient/protocol/heartbeat"
)

var (
	defaultHttpTimeout = time.Second * 3
)

type Result struct {
	resp protocol.Response
	err  error
}

func (r *Result) GetResp() protocol.Response {
	return r.resp
}

func (r *Result) GetErr() error {
	return r.err
}

type Dispatch func(result *Result)

type baseInfo struct {
	// 长连接的主备地址
	IpList []string
}

type ImClient struct {
	mux       sync.Mutex // 保护info
	Info      *baseInfo  // Info 基础信息
	serverUrl string     // IM 服务端WebApi地址
	cli       *gnet.Client
	sender    *sender
	dispatch  Dispatch
}

func NewImClient(dispatch Dispatch) (*ImClient, error) {
	// 初始化imcli
	imCli, err := initImCli(dispatch)
	if err != nil {
		return nil, err
	}
	return imCli, nil
}

func (imCli *ImClient) Start() error {
	// 启动事件轮询
	if err := imCli.cli.Start(); err != nil {
		return err
	}
	// 发一个echo, 确认网络通畅
	if _, err := imCli.WriteMessage(context.Background(), echo.NewRequest()); err != nil {
		return err
	}
	return nil
}

func (imCli *ImClient) WriteMessage(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	message, err := imCli.sender.writeMessage(ctx, request)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (imCli *ImClient) WriteMessageWithSeq(ctx context.Context, seq int32, request protocol.Request) error {
	return imCli.sender.writeMessageWihSeq(ctx, seq, request)
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
	err := conn.authReq(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (imCli *ImClient) fetchIpList() error {
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancelFunc()
	ips, err := http.IpList(timeout)
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		return errors.New("no available tcp gateway")
	}
	imCli.Info.IpList = ips
	return nil
}

func (imCli *ImClient) heartBeat() error {
	request := heartbeat.NewRequest()
	if err := imCli.WriteMessageWithSeq(context.Background(), 0, request); err != nil {
		logger.Errorf("heartBeat error: %v\n", err)
		return err
	}
	return nil
}

func initImCli(dispatch Dispatch) (*ImClient, error) {
	imCli := &ImClient{
		Info:     &baseInfo{},
		dispatch: dispatch,
	}
	if err := imCli.fetchIpList(); err != nil {
		return nil, err
	}
	// 构造transport
	transport := newTransport(imCli.dial(), imCli.auth, imCli.heartBeat)
	cli, err := gnet.NewClient(transport, gnet.WithTicker(true))
	if err != nil {
		return nil, err
	}
	imCli.cli = cli
	// 构造发送器
	imCli.sender = newSender(transport, 10, 3)
	return imCli, nil
}
