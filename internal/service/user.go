package service

import (
	"context"

	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/internal/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

func GetAllUser(ctx context.Context) ([]*sqllite.ImUser, error) {
	updateUsers()
	users, err := sqllite.GetAllUsers(ctx)
	return users, err
}

func updateUsers() {
	users, err := http.Users(context.Background())
	if err != nil {
		logger.Errorf("http.Users error: %v", err)
		return
	}
	if err := sqllite.BatchUpsertUsers(context.Background(), users); err != nil {
		logger.Errorf("BatchUpsertUsers error: %v", err)
		return
	}
}
