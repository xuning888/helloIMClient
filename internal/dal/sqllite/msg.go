package sqllite

import (
	"context"

	"gorm.io/gorm/clause"
)

// ChatMessage 映射到 chat_message 表
type ChatMessage struct {
	ChatID        int64  `gorm:"primaryKey;default:0;column:chat_id" json:"chatId"`
	MsgID         int64  `gorm:"primaryKey;default:0;column:msg_id" json:"msgId"`
	ChatType      int32  `gorm:"default:0;column:chat_type" json:"chatType"`
	MsgFrom       int64  `gorm:"default:0;column:msg_from" json:"msgFrom"`
	FromUserType  int32  `gorm:"default:0;column:from_user_type" json:"fromUserType"`
	MsgTo         int64  `gorm:"default:0;column:msg_to" json:"msgTo"`
	ToUserType    int32  `gorm:"default:0;column:to_user_type" json:"toUserType"`
	GroupID       int64  `gorm:"default:0;column:group_id" json:"groupId"`
	MsgSeq        int32  `gorm:"default:0;column:msg_seq" json:"msgSeq"`
	MsgContent    string `gorm:"type:text;column:msg_content" json:"msgContent"`
	ContentType   int32  `gorm:"default:0;column:content_type" json:"contentType"`
	CmdID         int32  `gorm:"default:0;column:cmd_id" json:"cmdId"`
	SendTime      int64  `gorm:"default:0;column:send_time" json:"sendTime"`
	ReceiptStatus int32  `gorm:"default:0;column:receipt_status" json:"receiptStatus"`
	ServerSeq     int64  `gorm:"default:0;column:server_seq" json:"serverSeq"`
}

func (ChatMessage) TableName() string {
	return "chat_message"
}

func NewMessage(chatType int32, chatId, msgId int64,
	msgFrom, msgTo int64,
	fromUserType, toUserType int32,
	msgSeq int32, msgContent string, contentType int32, cmdId int32, sendTime int64,
	receiptStatus int32, serverSeq int64) *ChatMessage {
	message := &ChatMessage{}
	message.ChatType = chatType
	message.ChatID = chatId
	message.MsgID = msgId
	message.MsgFrom = msgFrom
	message.MsgTo = msgTo
	message.FromUserType = fromUserType
	message.ToUserType = toUserType
	message.MsgSeq = msgSeq
	message.MsgContent = msgContent
	message.ContentType = contentType
	message.CmdID = cmdId
	message.SendTime = sendTime
	message.ReceiptStatus = receiptStatus
	message.ServerSeq = serverSeq
	if chatType == 2 {
		message.GroupID = message.ChatID
	}
	return message
}

func SaveOrUpdateMessage(ctx context.Context, message *ChatMessage) error {
	err := DB.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: "chat_id"}, {Name: "msg_id"},
			},
			UpdateAll: true,
		},
	).Create(message).Error
	if err != nil {
		return err
	}
	return nil
}

func GetMessage(ctx context.Context, chatId, msgId int64) (*ChatMessage, error) {
	msg := &ChatMessage{}
	err := DB.WithContext(ctx).Model(msg).
		Where("chat_id = ? and msg_id = ?", chatId, msgId).
		Find(msg).Error
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func GetLastMessage(ctx context.Context, chatId int64) (*ChatMessage, error) {
	msg := &ChatMessage{}
	err := DB.WithContext(ctx).Model(msg).
		Where("chat_id = ?", chatId).
		Order("server_seq desc").Limit(1).
		Find(msg).Error
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// GetMessagesWithOffset 分页获取消息
func GetMessagesWithOffset(ctx context.Context, chatId, offset int64, limit int) ([]*ChatMessage, error) {
	var msgs []*ChatMessage
	err := DB.WithContext(ctx).
		Model(&ChatMessage{}).
		Where("chat_id = ? and msg_id > ?", chatId, offset).
		Order("msg_id desc").
		Limit(limit).
		Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func GetRecentMessages(ctx context.Context, chatId int64, limit int) ([]*ChatMessage, error) {
	var msgs []*ChatMessage
	err := DB.WithContext(ctx).
		Model(&ChatMessage{}).
		Where("chat_id = ?", chatId).
		Order("server_seq desc").
		Limit(limit).
		Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
