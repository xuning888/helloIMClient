package pkg

import "time"

func FormatTime(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	now := time.Now()

	if t.Day() == now.Day() && t.Month() == now.Month() && t.Year() == now.Year() {
		return t.Format("15:04")
	} else if t.Year() == now.Year() {
		return t.Format("01/02")
	} else {
		return t.Format("2006/01/02")
	}
}
