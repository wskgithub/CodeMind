// Package timezone provides timezone utilities for Asia/Shanghai.
package timezone

import "time"

// Shanghai is the Asia/Shanghai timezone.
var Shanghai *time.Location

func loadShanghaiLocation(tzName string) *time.Location {
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return time.FixedZone("CST", 8*60*60) //nolint:mnd // 8 hours in seconds
	}
	return loc
}

func init() {
	Shanghai = loadShanghaiLocation("Asia/Shanghai")
}

// Now returns the current time in Asia/Shanghai timezone.
func Now() time.Time {
	return time.Now().In(Shanghai)
}

// Today returns today's midnight in Asia/Shanghai timezone (wrapped in UTC).
func Today() time.Time {
	now := Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

// TodayStr returns today's date string in Asia/Shanghai timezone (format: 2006-01-02).
func TodayStr() string {
	return Now().Format("2006-01-02")
}

// MonthStr returns the current month string in Asia/Shanghai timezone (format: 2006-01).
func MonthStr() string {
	return Now().Format("2006-01")
}
