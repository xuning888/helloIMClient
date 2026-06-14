package dbhelper

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func Init(userId int64, serverUrl string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	u, err := url.Parse(serverUrl)
	if err != nil {
		return err
	}
	absPath := filepath.Join(homeDir, ".helloIm", strings.ReplaceAll(u.Host, ":", "_"),
		fmt.Sprintf("%d", userId), "data.db")
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	return initDB(absPath)
}
