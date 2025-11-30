package sqllite

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

type Seq struct {
	ID     int32 `gorm:"primaryKey"`
	Number int32 `gorm:"column:number;default:0"`
}

func (Seq) TableName() string {
	return "seq"
}

var globalSeqManager *GlobalSequenceManager

type GlobalSequenceManager struct {
	currentID atomic.Int32
	db        *gorm.DB
}

func initSequenceManager(database *gorm.DB) error {
	globalSeqManager = &GlobalSequenceManager{
		db: database,
	}
	err := database.AutoMigrate(&Seq{})
	if err != nil {
		return err
	}
	var seq Seq
	res := database.Where("id = ?", 1).First(&seq)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		seq.ID = 1
		seq.Number = 0
		database.Create(&seq)
	} else if res.Error != nil {
		return res.Error
	}

	globalSeqManager.currentID.Store(seq.Number)

	go globalSeqManager.startPersistenceTask()

	return nil
}

func (g *GlobalSequenceManager) startPersistenceTask() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		g.saveCurrentID(context.Background())
	}
}

// saveCurrentID 将内存值持久化到数据库
func (g *GlobalSequenceManager) saveCurrentID(ctx context.Context) {
	currentValue := g.currentID.Load()
	err := g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Seq{}).Where("id = ?", 1).Update("number", currentValue)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return tx.Create(&Seq{ID: 1, Number: currentValue}).Error
		}
		return nil
	})
	if err != nil {
		log.Printf("Error saving sequence ID: %v\n", err)
	}
}

func GetSeq() int32 {
	return globalSeqManager.currentID.Add(1)
}

func OnShutdown(ctx context.Context) {
	globalSeqManager.saveCurrentID(ctx)
}
