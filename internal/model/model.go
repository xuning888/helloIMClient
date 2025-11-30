package model

import (
	"encoding/json"
)

type ChatInfo struct {
	Id                 string `json:"id"`
	UserID             int64  `json:"userId"`
	ChatID             int64  `json:"chatId"`
	ChatType           int32  `json:"chatType"`
	ChatTop            bool   `json:"chatTop"`
	ChatMute           bool   `json:"chatMute"`
	ChatDel            bool   `json:"chatDel"`
	UpdateTimestamp    int64  `json:"updateTimestamp"`
	DelTimestamp       int64  `json:"delTimestamp"`
	LastReadMsgID      int64  `json:"lastReadMsgId"`
	SubStatus          int32  `json:"subStatus"`
	JoinGroupTimestamp int64  `json:"joinGroupTimestamp"`
}

type ImUser struct {
	UserId     int64  `json:"userId"`
	UserType   int32  `json:"userType"`
	UserName   string `json:"userName"`
	Icon       string `json:"icon"`
	Mobile     string `json:"mobile"`
	DeviceId   string `json:"deviceId"`
	Extra      string `json:"extra"`
	UserStatus int32  `json:"userStatus"`
	Token      string `json:"token"`
}

func (m *ImUser) String() string {
	if m == nil {
		return "{}"
	}
	marshal, _ := json.Marshal(m)
	return string(marshal)
}
