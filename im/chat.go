package im

type ChatType int

const (
	C2C ChatType = 1
	C2G ChatType = 2
)

type Chat struct {
	Id              int64          // 会话ID
	Type            ChatType       // 会话类型
	messages        []*ChatMessage // 会话中的消息
	lastChatMessage *ChatMessage   // 会话最后一条消息
	ChatName        string         // 会话名称
	Timestamp       int64          // 会话的更新时间
	UnReadNum       int            // 会话未读数
}
