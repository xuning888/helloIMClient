package sqllite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/xuning888/helloIMClient/conf"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"gorm.io/gorm"
)

type ImChat struct {
	UserId             int64 `gorm:"column:user_id;default:0;primaryKey" json:"userId"`
	ChatId             int64 `gorm:"column:chat_id;default:0;primaryKey" json:"chatId"`
	ChatType           int32 `gorm:"column:chat_type;default:1" json:"chatType"`
	ChatTop            bool  `gorm:"column:chat_top;default:0" json:"chatTop"`
	ChatMute           bool  `gorm:"column:chat_mute;default:0" json:"chatMute"`
	ChatDel            bool  `gorm:"column:chat_del;default:0" json:"chatDel"`
	UpdateTimestamp    int64 `gorm:"column:update_timestamp;default:0" json:"updateTimestamp"`
	DelTimestamp       int64 `gorm:"column:del_timestamp;default:0" json:"delTimestamp"`
	LastReadMsgId      int64 `gorm:"column:last_read_msg_id;default:0" json:"lastReadMsgId"`
	SubStatus          int   `gorm:"column:sub_status;default:0" json:"subStatus"`
	JoinGroupTimestamp int64 `gorm:"column:join_group_timestamp;default:0" json:"joinGroupTimestamp"`
}

func (ImChat) TableName() string {
	return "im_chat"
}

func (chat ImChat) String() string {
	marshal, err := json.Marshal(&chat)
	if err != nil {
		return ""
	}
	return string(marshal)
}

func (c ImChat) Key() string {
	return fmt.Sprintf("%d_%d_%d", c.UserId, c.ChatId, c.ChatType)
}

func NewImChat(userId, chatId int64, chatType int32) *ImChat {
	chat := &ImChat{
		UserId:             userId,
		ChatId:             chatId,
		ChatType:           chatType,
		ChatTop:            false,
		ChatMute:           false,
		ChatDel:            false,
		UpdateTimestamp:    time.Now().UnixMilli(),
		DelTimestamp:       0,
		LastReadMsgId:      0,
		SubStatus:          0,
		JoinGroupTimestamp: 0,
	}
	return chat
}

// SortChatList 会话列表排序
func SortChatList(chats []*ImChat) {
	sort.Slice(chats, func(i, j int) bool {
		if chats[i].ChatTop != chats[j].ChatTop {
			return chats[i].ChatTop
		}
		return chats[i].UpdateTimestamp > chats[j].UpdateTimestamp
	})
}

func BatchUpdate(ctx context.Context, chats []*ImChat) error {
	if len(chats) == 0 {
		return nil
	}
	var updates, inserts = make([]*ImChat, 0), make([]*ImChat, 0)
	for _, chat := range chats {
		_, err := SelectChat(ctx, chat.UserId, chat.ChatId)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			inserts = append(inserts, chat)
		} else if err != nil {
			return err
		} else {
			// 存在的记录
			updates = append(updates, chat)
		}
	}
	if len(inserts) > 0 {
		if err := DB.WithContext(ctx).Model(&ImChat{}).Create(&inserts).Error; err != nil {
			return err
		}
	}
	for _, chat := range updates {
		if err := DB.WithContext(ctx).Model(&ImChat{}).
			Where("user_id = ? AND chat_id = ?", chat.UserId, chat.ChatId).
			Updates(chat).Error; err != nil {
			return err
		}
	}
	return nil
}

func SelectChat(ctx context.Context, userId, chatId int64) (*ImChat, error) {
	chat := &ImChat{}
	err := DB.WithContext(ctx).Model(&ImChat{}).
		Where("user_id = ? and chat_id = ?", userId, chatId).
		First(chat).Error
	if err != nil {
		return nil, err
	}
	return chat, nil
}

// MultiGetChat
// Note: 查询100条会话
func MultiGetChat(ctx context.Context) ([]*ImChat, error) {
	var chats = make([]*ImChat, 0)
	res := DB.WithContext(ctx).Model(&ImChat{}).
		Where("user_id = ?", conf.UserId).
		Order("chat_top desc").
		Order("update_timestamp desc").
		Limit(100).Find(&chats)
	err := res.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	SortChatList(chats)
	logger.Infof("MultiGetChat chats: %v", chats)
	return chats, nil
}

func InsertChat(ctx context.Context, chat *ImChat) error {
	if chat == nil {
		return nil
	}
	if err := DB.WithContext(ctx).Model(chat).Create(chat).Error; err != nil {
		return err
	}
	return nil
}
