package dbhelper

import (
	"fmt"
	"os"
	"path/filepath"
)

func Init(userId int64) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	absPath := filepath.Join(homeDir, ".helloIm", fmt.Sprintf("%d", userId), "data.db")
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	return initDB(absPath)
}
