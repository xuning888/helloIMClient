package im

type ChatMessage struct {
	ChatId   string   // 会话ID
	ChatType ChatType // 会话类型
	FromUid  string   // 消息发送者的ID
	FromName string   // 消息发送者的名称
	ToUid    string   // 消息接受者的ID
	
	ServerSeq int // 服务端的消息序号
}
