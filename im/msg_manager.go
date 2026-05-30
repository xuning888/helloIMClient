package im

// MsgManager 消息管理器
type MsgManager struct {
	cli *Client
}

func newMsgManager(cli *Client) *MsgManager {
	return &MsgManager{cli: cli}
}

// AddNewMsgListener 注册新消息回调，返回取消函数
func (mm *MsgManager) AddNewMsgListener(cb EventCallback) func() {
	return mm.cli.events.subscribe(func(evt Event) {
		if evt.Type == EventMessageReceived {
			cb(evt)
		}
	})
}

// AddOnSendMsgListener 注册消息发送回调
func (mm *MsgManager) AddOnSendMsgListener(cb EventCallback) func() {
	return mm.cli.events.subscribe(func(evt Event) {
		if evt.Type == EventMessageSent {
			cb(evt)
		}
	})
}

func (mm *MsgManager) AddMessageInsertListener(cb EventCallback) func() {
	return mm.cli.events.subscribe(func(evt Event) {
		if evt.Type == EventMessageInsert {
			cb(evt)
		}
	})
}
