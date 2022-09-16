package time

import (
	"fmt"
	"time"
)

// UnixNanoString creates a string containing the Unix nano
// timestamp that is valid up to the limit of Unix
func UnixNanoString(t time.Time) string {
	return fmt.Sprintf("%d%09d", t.Unix(), t.Nanosecond())
}

// Date creates a string representing the time in YYYY-MM-DD format
func Date(t time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
}
