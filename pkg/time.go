package pkg

import "time"

var (
	DateTime = "2006/01/02 15:04:05"
	Date     = "2006-01-02"
)

func FormatTime(timestamp int64, pattern string) string {
	t := time.Unix(timestamp/1000, (timestamp%1000)*1e6)
	return t.Format(pattern)
}
