package sqllite

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/xuning888/helloIMClient/pkg/logger"
	"gorm.io/gorm/clause"
)

type ImUser struct {
	UserID     int64  `gorm:"column:user_id;primaryKey;not null;default:0" json:"userId"`
	UserType   int    `gorm:"column:user_type;not null;default:0" json:"userType"`
	UserName   string `gorm:"column:user_name;not null;default:''" json:"userName"`
	Icon       string `gorm:"column:icon;not null;default:''" json:"icon"`
	Mobile     string `gorm:"column:mobile;not null;default:''" json:"mobile"`
	Extra      string `gorm:"column:extra;not null;default:''" json:"extra"`
	UserStatus int    `gorm:"column:user_status;not null;default:0" json:"userStatus"`
}

func (ImUser) TableName() string {
	return "im_user"
}

func (u *ImUser) String() string {
	if u == nil {
		return ""
	}
	marshal, err := json.Marshal(u)
	if err != nil {
		return ""
	}
	return string(marshal)
}

func SearchUser(ctx context.Context, key string) ([]*ImUser, error) {
	if len(strings.TrimSpace(key)) == 0 {
		return []*ImUser{}, nil
	}
	var users []*ImUser
	err := DB.WithContext(ctx).
		Where("user_name like ?", "%"+key+"%").
		Limit(20).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func GetUserById(ctx context.Context, userId int64) (*ImUser, error) {
	user := &ImUser{}
	logger.Infof("GetUserById userId: %v", userId)
	err := DB.WithContext(ctx).Where("user_id = ?", userId).First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetAllUsers(ctx context.Context) ([]*ImUser, error) {
	var users []*ImUser
	err := DB.WithContext(ctx).
		Order("user_id ASC").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func BatchUpsertUsers(ctx context.Context, users []*ImUser) error {
	if len(users) == 0 {
		return nil
	}
	return DB.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"user_type", "user_name", "icon", "mobile", "extra", "user_status",
			}),
		},
	).Create(&users).Error
}
