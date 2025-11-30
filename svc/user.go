package svc

import (
	"sync"

	"github.com/xuning888/helloIMClient/internal/model"
)

type UserSvc struct {
	users sync.Map
}

func (um *UserSvc) AddUser(user *model.ImUser) {
	um.users.Store(user.UserId, user)
}

func (um *UserSvc) GetUser(userId int64) *model.ImUser {
	if value, ok := um.users.Load(userId); ok {
		return value.(*model.ImUser)
	}
	// TODO 拉用户
	return nil
}

func NewUserSvc(users []*model.ImUser) *UserSvc {
	userSvc := &UserSvc{
		users: sync.Map{},
	}
	if len(users) != 0 {
		for _, user := range users {
			userSvc.AddUser(user)
		}
	}
	return userSvc
}
