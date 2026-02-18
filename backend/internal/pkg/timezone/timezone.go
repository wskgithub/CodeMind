package timezone

import "time"

// Shanghai 东八区时区
var Shanghai *time.Location

func init() {
	var err error
	Shanghai, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 若时区数据不可用，手动创建 UTC+8
		Shanghai = time.FixedZone("CST", 8*60*60)
	}
}

// Now 获取 Asia/Shanghai 时区的当前时间
func Now() time.Time {
	return time.Now().In(Shanghai)
}

// Today 获取 Asia/Shanghai 时区的今日零点时间（UTC 包装）
// 使用 UTC 时区包装日期值，避免 pgx 驱动在写入 PostgreSQL DATE 字段时
// 因时区转换导致日期偏移（如 +0800 的 2月17日 00:00 被转为 UTC 2月16日 16:00，截取后变成 2月16日）
func Today() time.Time {
	now := Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

// TodayStr 获取 Asia/Shanghai 时区的今日日期字符串（格式：2006-01-02）
func TodayStr() string {
	return Now().Format("2006-01-02")
}

// MonthStr 获取 Asia/Shanghai 时区的当月字符串（格式：2006-01）
func MonthStr() string {
	return Now().Format("2006-01")
}
