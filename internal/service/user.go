package service

import (
	"context"

	"github.com/hashicorp/golang-lru/v2"
	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

var cache *lru.Cache[int64, *sqllite.ImUser]

func init() {
	var err error
	cache, err = lru.New[int64, *sqllite.ImUser](500)
	if err != nil {
		panic("failed to create LRU cache: " + err.Error())
	}
}

func GetUserById(ctx context.Context, userId int64) (*sqllite.ImUser, error) {
	value, ok := cache.Get(userId)
	if ok {
		return value, nil
	}
	user, err := sqllite.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}
	cache.Add(userId, user)
	return user, nil
}

func UpdateUsers() {
	users, err := http.Users(context.Background())
	if err != nil {
		logger.Errorf("http.Users error: %v", err)
		return
	}
	if err := sqllite.BatchUpsertUsers(context.Background(), users); err != nil {
		logger.Errorf("BatchUpsertUsers error: %v", err)
		return
	}
	for _, user := range users {
		cache.Add(user.UserID, user)
	}
}
