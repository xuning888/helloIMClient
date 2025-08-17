package net

import (
	"errors"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/xuning888/helloIMClient/frame"
	"github.com/xuning888/helloIMClient/pkg"
	pb "github.com/xuning888/helloIMClient/proto"
	"google.golang.org/protobuf/proto"
)

var (
	ErrorsState = errors.New("长连接还未建立")
)

// sendFrame 同步发送消息
// 只关注消息发没发出去，不关注ACK
func (imCli *ImClient) sendFrame(frame *frame.Frame) error {
	if frame == nil {
		return errors.New("frame is nil")
	}
	// 检查状态，使用读锁
	imCli.lock.RLock()
	defer imCli.lock.RUnlock()
	state := imCli.state
	if state == Init || state == Closed || state == Connecting {
		return ErrorsState
	}
	// 使用promise把异步转换为同步, 作用于gnet和业务层的通信
	promise := pkg.NewPromise[interface{}]()
	msg := frame.ToBytes()
	err := imCli.conn.AsyncWrite(msg, func(c gnet.Conn, err error) error {
		if err != nil {
			promise.Complete(nil, err)
			return err
		}
		promise.Complete(struct{}{}, nil)
		return nil
	})
	if err != nil {
		promise.Complete(nil, err)
	}
	_, err2 := promise.Get(0)
	if err2 != nil {
		return err2
	}
	return nil
}

// SendFrameWithCallback 异步发送消息
// 可以设置回调函数, 回调函数用于关注消息的ACK
// 上行发送成功之后才会将frame注册到飞行队列
func (imCli *ImClient) SendFrameWithCallback(request *frame.Frame, ackCallBack WhenComplete) error {
	if request == nil {
		return errors.New("frame is nil")
	}
	// 检查状态，使用读锁
	imCli.lock.RLock()
	defer imCli.lock.RUnlock()
	state := imCli.state
	if state == Init || state == Closed || state == Connecting {
		return ErrorsState
	}
	msg := request.ToBytes()
	err := imCli.conn.AsyncWrite(msg, func(c gnet.Conn, err error) error {
		if err != nil {
			// 如果在这里触发，基本是网络或者框架异常, 消息没发出去, 取消ACK
			imCli.inflightQ.RemoveAndStop(request)
			return err
		}
		return nil
	})
	// 这个error是将发送行为提交到gnet的antsPool或者poller的队列失败, 消息也没有被发送出去, 直接返回error
	if err != nil {
		return err
	}
	// 将request注册到飞行队列
	imCli.inflightQ.Put(request, func(req, reply *frame.Frame, err error) {
		// 再包一层, 关注ACK的处理ACK，不关注ACK的就忽略
		if ackCallBack != nil {
			ackCallBack(req, reply, err)
		}
		// 回调到业务处理器
		imCli.handle(reply)
	})
	return nil
}

func (imCli *ImClient) auth(request *pb.AuthRequest) error {
	promise := imCli.asyncAuth(request)
	authResponse, err := promise.Get(time.Second * 10)
	if err != nil {
		return err
	}
	// auth成功, 设置cli的状态
	if authResponse.Success {
		return nil
	}
	return ErrorsAuthFailed
}

func (imCli *ImClient) asyncAuth(request *pb.AuthRequest) *pkg.Promise[*pb.AuthResponse] {
	var p = pkg.NewPromise[*pb.AuthResponse]()

	marshal, err := proto.Marshal(request)
	if err != nil {
		p.Complete(nil, err)
		return p
	}

	requestFrame := &frame.Frame{
		Header: frame.NewMsgHeader(1, GetSeq(), int(pb.CmdId_CMD_ID_AUTH), len(marshal)),
		Body:   marshal,
	}

	err3 := imCli.SendFrameWithCallback(requestFrame, func(req, reply *frame.Frame, err error) {
		if err != nil {
			p.Complete(nil, err)
			return
		}
		// 反序列化
		authResponse := &pb.AuthResponse{}
		if err2 := proto.Unmarshal(reply.Body, authResponse); err2 != nil {
			p.Complete(nil, err2)
			return
		}
		// 设置ACK结果
		p.Complete(authResponse, nil)
	})

	// 异步发送消息失败, 这里是这里是gnet框架范围的异常
	if err3 != nil {
		p.Complete(nil, err)
	}
	return p
}
