package sqllite

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/xuning888/helloIMClient/pkg/logger"
)

var DB *gorm.DB

func Init(DSN string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(DSN), &gorm.Config{
		Logger: logger.NewGormLogger(),
	})
	if err != nil {
		return err
	}
	if err = migrate(); err != nil {
		return err
	}
	return nil
}

func migrate() error {
	if err := DB.AutoMigrate(&ImUser{}); err != nil {
		return err
	}
	if err := DB.AutoMigrate(&ChatMessage{}); err != nil {
		return err
	}
	if err := DB.AutoMigrate(&ImChat{}); err != nil {
		return err
	}
	if err := initSequenceManager(DB); err != nil {
		return err
	}
	return nil
}
