package dal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
)

func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	absPath := filepath.Join(homeDir, ".helloIm", "data.db")
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	err = sqllite.Init(absPath)
	if err != nil {
		return err
	}
	return nil
}
