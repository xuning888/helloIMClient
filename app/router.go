package app

import (
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/svc"
	"github.com/xuning888/helloIMClient/transport"
	"sync"
)

var contextPool = sync.Pool{
	New: func() interface{} {
		return &ImContext{}
	},
}

type Handler func(ctx *ImContext) error

type router struct {
	imCli     *transport.ImClient
	commonSvc *svc.CommonSvc
	handlers  map[int32]Handler
}

func (h *router) route(ctx *ImContext) error {
	cmdId := ctx.CmdId
	if handler, exits := h.handlers[cmdId]; exits {
		return handler(ctx)
	} else {
		return defaultHandler(ctx)
	}
}

func (h *router) dispatch(result *transport.Result) {
	context := contextPool.Get().(*ImContext)
	context.imCli = h.imCli                  // 设置imCli
	context.CmdId = result.GetResp().CmdId() // 设置指令号
	context.response = result.GetResp()      // 设置下行消息
	defer func() {
		context.reset()
		contextPool.Put(context)
	}()
	if err := h.route(context); err != nil {
		logger.Errorf("指令处理失败, cmdId: %d, err: %v", context.CmdId, err)
	}
}

// Register 注册处理器
func (h *router) Register(cmdId int32, handler Handler) {
	h.handlers[cmdId] = handler
}

func defaultHandler(ctx *ImContext) error {
	logger.Infof("defaultHandler 未知消息类型, cmdId: %d", ctx.CmdId)
	return nil
}
