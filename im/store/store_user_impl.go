package store

import (
	"context"

	"github.com/xuning888/helloIMClient/im/dbhelper"
	"github.com/xuning888/helloIMClient/im/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

// ---- UserStore ----

func (s *StorageImpl) Get(ctx context.Context, userId int64) (*dbhelper.ImUser, error) {
	cache := s.cache
	value, ok := cache.Get(userId)
	if ok {
		return value, nil
	}
	user, err := dbhelper.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}
	cache.Add(userId, user)
	return user, nil
}

func (s *StorageImpl) Search(ctx context.Context, keyword string) ([]*dbhelper.ImUser, error) {
	return dbhelper.SearchUser(ctx, keyword)
}

func (s *StorageImpl) Refresh(ctx context.Context) error {
	users, err := http.Users(context.Background())
	if err != nil {
		logger.Errorf("http.Users error: %v", err)
		return err
	}
	if err := dbhelper.BatchUpsertUsers(context.Background(), users); err != nil {
		logger.Errorf("BatchUpsertUsers error: %v", err)
		return err
	}
	for _, user := range users {
		s.cache.Add(user.UserID, user)
	}
	return nil
}
