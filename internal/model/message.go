package model

import (
	"encoding/json"
	"fmt"
)

type ChatMessage struct {
	ChatType      int32  `json:"chatType,omitempty"`
	ChatId        int64  `json:"chatId,omitempty"`
	ChatIdStr     string `json:"chatIdStr,omitempty"`
	MsgId         int64  `json:"msgId,omitempty"`
	MsgIdStr      string `json:"msgIdStr,omitempty"`
	MsgFrom       int64  `json:"msgFrom,omitempty"`
	MsgFromStr    string `json:"msgFromStr,omitempty"`
	FromUserType  int32  `json:"fromUserType,omitempty"`
	MsgTo         int64  `json:"msgTo,omitempty"`
	MsgToStr      string `json:"msgToStr,omitempty"`
	ToUserType    int32  `json:"toUserType,omitempty"`
	GroupId       int64  `json:"groupId,omitempty"`
	GroupIdStr    string `json:"groupIdStr,omitempty"`
	MsgSeq        int32  `json:"msgSeq,omitempty"`
	MsgContent    string `json:"msgContent,omitempty"`
	ContentType   int32  `json:"contentType,omitempty"`
	CmdId         int32  `json:"cmdId,omitempty"`
	SendTime      int64  `json:"sendTime,omitempty"`
	ReceiptStatus int32  `json:"receiptStatus,omitempty"`
	ServerSeq     int64  `json:"serverSeq,omitempty"`
}

func (m *ChatMessage) String() string {
	if m == nil {
		return "{}"
	}
	marshal, _ := json.Marshal(m)
	return string(marshal)
}

func NewMessage(chatType int32, chatId, msgId int64,
	msgFrom, msgTo int64,
	fromUserType, toUserType int32,
	msgSeq int32, msgContent string, contentType int32, cmdId int32, sendTime int64,
	receiptStatus int32, serverSeq int64) *ChatMessage {
	message := &ChatMessage{}
	message.ChatType = chatType
	message.ChatId = chatId
	message.ChatIdStr = fmt.Sprintf("%d", chatId)
	message.MsgId = msgId
	message.MsgIdStr = fmt.Sprintf("%d", msgId)
	message.MsgFrom = msgFrom
	message.MsgFromStr = fmt.Sprintf("%d", msgFrom)
	message.MsgTo = msgTo
	message.MsgToStr = fmt.Sprintf("%d", msgTo)
	message.FromUserType = fromUserType
	message.ToUserType = toUserType
	message.MsgSeq = msgSeq
	message.MsgContent = msgContent
	message.ContentType = contentType
	message.CmdId = cmdId
	message.SendTime = sendTime
	message.ReceiptStatus = receiptStatus
	message.ServerSeq = serverSeq
	if chatType == 2 {
		message.GroupId = message.ChatId
		message.GroupIdStr = message.ChatIdStr
	}
	return message
}
