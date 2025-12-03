package pkg

import "time"

func FormatTime(timestamp int64) string {
	t := time.Unix(timestamp/1000, (timestamp%1000)*1e6)
	return t.Format("2006-01-02 15:04:05")
}
