package svc

import (
	"encoding/json"
	"sync"
)

var Me = &User{}

type User struct {
	UserId   int64  `json:"userId"`   // UserId
	UserName string `json:"userName"` // 用户名称
	UserType int    `json:"userType"` // 用户类型
}

func (u *User) String() string {
	marshal, _ := json.Marshal(u)
	return string(marshal)
}

type UserSvc struct {
	users sync.Map
}

func (um *UserSvc) AddUser(user *User) {
	um.users.Store(user.UserId, user)
}

func (um *UserSvc) GetUser(userId int64) *User {
	if value, ok := um.users.Load(userId); ok {
		return value.(*User)
	}
	// TODO 拉用户
	return nil
}

func newUserSvc(users []*User) *UserSvc {
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
