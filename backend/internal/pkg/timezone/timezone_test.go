package timezone

import (
	"strings"
	"testing"
	"time"
)

// TestShanghaiInitialization 测试 Shanghai 时区是否正确初始化.
func TestShanghaiInitialization(t *testing.T) {
	tests := []struct {
		validate func(*time.Location) bool
		name     string
	}{
		{
			name: "Shanghai 时区不应为 nil",
			validate: func(loc *time.Location) bool {
				return loc != nil
			},
		},
		{
			name: "Shanghai 时区名称应为 Asia/Shanghai 或 CST",
			validate: func(loc *time.Location) bool {
				// 可能是 "Asia/Shanghai" 或 "CST"（当手动创建时）
				name := loc.String()
				return name == "Asia/Shanghai" || name == "CST"
			},
		},
		{
			name: "Shanghai 时区偏移应为 +0800",
			validate: func(loc *time.Location) bool {
				// 使用参考时间检查偏移量
				ref := time.Date(2024, 1, 1, 12, 0, 0, 0, loc)
				_, offset := ref.Zone()
				return offset == 8*60*60 // +8小时 = 28800秒
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.validate(Shanghai) {
				t.Errorf("Shanghai 时区验证失败: %s", tt.name)
			}
		})
	}
}

// TestNow 测试 Now() 返回上海时区时间.
func TestNow(t *testing.T) {
	tests := []struct {
		validate func(time.Time) bool
		name     string
	}{
		{
			name: "返回的时间时区应为 Shanghai",
			validate: func(tm time.Time) bool {
				return tm.Location().String() == Shanghai.String()
			},
		},
		{
			name: "返回的时间不应为零值",
			validate: func(tm time.Time) bool {
				return !tm.IsZero()
			},
		},
		{
			name: "返回的时间应在合理范围内（前后1分钟）",
			validate: func(tm time.Time) bool {
				// 获取当前 UTC 时间并转换为上海时区
				nowUTC := time.Now().UTC()
				diff := tm.Sub(nowUTC)
				// 时间差应在 -1分钟到+1分钟之间
				return diff > -time.Minute && diff < time.Minute
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := Now()
			if !tt.validate(now) {
				t.Errorf("Now() 验证失败: %s, got: %v", tt.name, now)
			}
		})
	}
}

// TestToday 测试 Today() 返回今日零点且时区为 UTC.
func TestToday(t *testing.T) {
	tests := []struct {
		validate func(time.Time) bool
		name     string
	}{
		{
			name: "返回的时间时区应为 UTC",
			validate: func(tm time.Time) bool {
				return tm.Location().String() == "UTC"
			},
		},
		{
			name: "时间应为零点 (00:00:00)",
			validate: func(tm time.Time) bool {
				return tm.Hour() == 0 && tm.Minute() == 0 && tm.Second() == 0 && tm.Nanosecond() == 0
			},
		},
		{
			name: "日期应与上海时区当前日期一致",
			validate: func(tm time.Time) bool {
				now := Now()
				return tm.Year() == now.Year() && tm.Month() == now.Month() && tm.Day() == now.Day()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			today := Today()
			if !tt.validate(today) {
				t.Errorf("Today() 验证失败: %s, got: %v", tt.name, today)
			}
		})
	}
}

// TestTodayStr 测试 TodayStr() 格式正确.
func TestTodayStr(t *testing.T) {
	tests := []struct {
		validate func(string) bool
		name     string
	}{
		{
			name: "格式应为 YYYY-MM-DD",
			validate: func(s string) bool {
				// 检查格式：4位数字-2位数字-2位数字
				parts := strings.Split(s, "-")
				if len(parts) != 3 {
					return false
				}
				return len(parts[0]) == 4 && len(parts[1]) == 2 && len(parts[2]) == 2
			},
		},
		{
			name: "日期应与上海时区当前日期一致",
			validate: func(s string) bool {
				expected := Now().Format("2006-01-02")
				return s == expected
			},
		},
		{
			name: "应能正确解析为时间",
			validate: func(s string) bool {
				_, err := time.Parse("2006-01-02", s)
				return err == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todayStr := TodayStr()
			if !tt.validate(todayStr) {
				t.Errorf("TodayStr() 验证失败: %s, got: %s", tt.name, todayStr)
			}
		})
	}
}

// TestMonthStr 测试 MonthStr() 格式正确.
func TestMonthStr(t *testing.T) {
	tests := []struct {
		validate func(string) bool
		name     string
	}{
		{
			name: "格式应为 YYYY-MM",
			validate: func(s string) bool {
				// 检查格式：4位数字-2位数字
				parts := strings.Split(s, "-")
				if len(parts) != 2 {
					return false
				}
				return len(parts[0]) == 4 && len(parts[1]) == 2
			},
		},
		{
			name: "月份应与上海时区当前月份一致",
			validate: func(s string) bool {
				expected := Now().Format("2006-01")
				return s == expected
			},
		},
		{
			name: "应能正确解析为时间",
			validate: func(s string) bool {
				_, err := time.Parse("2006-01", s)
				return err == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monthStr := MonthStr()
			if !tt.validate(monthStr) {
				t.Errorf("MonthStr() 验证失败: %s, got: %s", tt.name, monthStr)
			}
		})
	}
}

// TestCrossDayTimezoneHandling 测试跨天时区处理.
func TestCrossDayTimezoneHandling(t *testing.T) {
	tests := []struct {
		inputTime      time.Time
		name           string
		expectedHour   int
		expectedOffset int
	}{
		{
			name:           "UTC 零点转换为上海时区",
			inputTime:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedHour:   8, // UTC 00:00 = Shanghai 08:00
			expectedOffset: 8 * 60 * 60,
		},
		{
			name:           "UTC 16点转换为上海时区",
			inputTime:      time.Date(2024, 1, 15, 16, 0, 0, 0, time.UTC),
			expectedHour:   0, // UTC 16:00 = Shanghai 00:00 (次日)
			expectedOffset: 8 * 60 * 60,
		},
		{
			name:           "纽约时间转换为上海时区",
			inputTime:      time.Date(2024, 1, 15, 12, 0, 0, 0, time.FixedZone("EST", -5*60*60)),
			expectedHour:   1, // EST 12:00 = Shanghai 01:00 (次日)
			expectedOffset: 8 * 60 * 60,
		},
		{
			name:           "跨天边界 - UTC 23:59",
			inputTime:      time.Date(2024, 1, 15, 23, 59, 0, 0, time.UTC),
			expectedHour:   7, // UTC 23:59 = Shanghai 07:59 (次日)
			expectedOffset: 8 * 60 * 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 将输入时间转换为上海时区
			shanghaiTime := tt.inputTime.In(Shanghai)

			// 验证小时
			if shanghaiTime.Hour() != tt.expectedHour {
				t.Errorf("小时不匹配: expected %d, got %d", tt.expectedHour, shanghaiTime.Hour())
			}

			// 验证偏移量
			_, offset := shanghaiTime.Zone()
			if offset != tt.expectedOffset {
				t.Errorf("时区偏移不匹配: expected %d, got %d", tt.expectedOffset, offset)
			}
		})
	}
}

// 注：中国自1991年起不再使用夏令时，但时区库应正确处理其他地区的夏令时.
func TestDaylightSavingTime(t *testing.T) {
	tests := []struct {
		testTime       time.Time
		name           string
		locationName   string
		expectedOffset int
		expectDST      bool
	}{
		{
			name:           "上海时区不使用夏令时",
			locationName:   "Asia/Shanghai",
			testTime:       time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC), // 夏季
			expectDST:      false,
			expectedOffset: 8 * 60 * 60, // 始终是 +8
		},
		{
			name:           "上海时区冬季不使用夏令时",
			locationName:   "Asia/Shanghai",
			testTime:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC), // 冬季
			expectDST:      false,
			expectedOffset: 8 * 60 * 60, // 始终是 +8
		},
		{
			name:           "美国东部夏令时",
			locationName:   "America/New_York",
			testTime:       time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC), // 夏季
			expectDST:      true,
			expectedOffset: -4 * 60 * 60, // EDT = UTC-4
		},
		{
			name:           "美国东部标准时间",
			locationName:   "America/New_York",
			testTime:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC), // 冬季
			expectDST:      false,
			expectedOffset: -5 * 60 * 60, // EST = UTC-5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := time.LoadLocation(tt.locationName)
			if err != nil {
				t.Skipf("无法加载时区 %s: %v", tt.locationName, err)
				return
			}

			localTime := tt.testTime.In(loc)
			_, offset := localTime.Zone()

			if offset != tt.expectedOffset {
				t.Errorf("时区偏移不匹配: expected %d, got %d", tt.expectedOffset, offset)
			}
		})
	}
}

// TestTodayConsistency 测试 Today() 与 TodayStr() 的一致性.
func TestTodayConsistency(t *testing.T) {
	// Today() 返回的时间应该与 TodayStr() 表示的日期一致
	today := Today()
	todayStr := TodayStr()

	// TodayStr() 基于上海时区，Today() 也基于上海时区的当前日期
	// 但 Today() 返回的是 UTC 包装的时间
	// 需要比较的是日期值，而不是字符串表示
	nowShanghai := Now()
	expectedDate := time.Date(nowShanghai.Year(), nowShanghai.Month(), nowShanghai.Day(), 0, 0, 0, 0, time.UTC)

	if !today.Equal(expectedDate) {
		t.Errorf("Today() 日期不匹配: expected %v, got %v", expectedDate, today)
	}

	// TodayStr 应该表示上海时区的当前日期
	expectedStr := nowShanghai.Format("2006-01-02")
	if todayStr != expectedStr {
		t.Errorf("TodayStr() 不匹配: expected %s, got %s", expectedStr, todayStr)
	}
}

// TestMonthStrConsistency 测试 MonthStr() 与 Now() 的一致性.
func TestMonthStrConsistency(t *testing.T) {
	now := Now()
	monthStr := MonthStr()
	expectedStr := now.Format("2006-01")

	if monthStr != expectedStr {
		t.Errorf("MonthStr() 不匹配: expected %s, got %s", expectedStr, monthStr)
	}
}

// BenchmarkNow 基准测试 Now().
func BenchmarkNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Now()
	}
}

// BenchmarkToday 基准测试 Today().
func BenchmarkToday(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Today()
	}
}

// BenchmarkTodayStr 基准测试 TodayStr().
func BenchmarkTodayStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		TodayStr()
	}
}

// BenchmarkMonthStr 基准测试 MonthStr().
func BenchmarkMonthStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MonthStr()
	}
}

// TestLoadShanghaiLocation 测试 loadShanghaiLocation 函数的两种分支.
func TestLoadShanghaiLocation(t *testing.T) {
	tests := []struct {
		name           string
		tzName         string
		expectedName   string
		expectedOffset int
	}{
		{
			name:           "加载有效的 Asia/Shanghai 时区",
			tzName:         "Asia/Shanghai",
			expectedName:   "Asia/Shanghai",
			expectedOffset: 8 * 60 * 60,
		},
		{
			name:           "加载无效的时区应回退到 CST",
			tzName:         "Invalid/Timezone",
			expectedName:   "CST",
			expectedOffset: 8 * 60 * 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := loadShanghaiLocation(tt.tzName)

			if loc == nil {
				t.Fatal("loadShanghaiLocation() 返回 nil")
			}

			// 验证时区名称
			name := loc.String()
			if name != tt.expectedName {
				t.Errorf("时区名称错误: expected %s, got %s", tt.expectedName, name)
			}

			// 验证偏移量
			ref := time.Date(2024, 1, 1, 12, 0, 0, 0, loc)
			_, offset := ref.Zone()
			if offset != tt.expectedOffset {
				t.Errorf("时区偏移错误: expected %d, got %d", tt.expectedOffset, offset)
			}
		})
	}
}

// 验证手动创建的 CST 时区与从数据库加载的 Asia/Shanghai 功能等效.
func TestShanghaiTimezoneFallback(t *testing.T) {
	// 创建手动 CST 时区（模拟 init 中的回退逻辑）
	cstZone := time.FixedZone("CST", 8*60*60)
	if cstZone == nil {
		t.Error("无法创建 CST 时区")
	}

	// 验证手动创建的时区名称
	if cstZone.String() != "CST" {
		t.Errorf("CST 时区名称错误: expected CST, got %s", cstZone.String())
	}

	// 验证手动创建的时区可以正确转换时间
	utcTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	cstTime := utcTime.In(cstZone)

	// UTC 12:00 = CST 20:00
	if cstTime.Hour() != 20 {
		t.Errorf("CST 时间转换错误: expected hour 20, got %d", cstTime.Hour())
	}

	// 验证 Shanghai 时区已被正确初始化（无论是加载的还是手动创建的）
	if Shanghai == nil {
		t.Error("Shanghai 时区不应为 nil")
	}

	// 验证 Shanghai 和 CST 时区功能等效
	shanghaiTime := utcTime.In(Shanghai)
	if shanghaiTime.Hour() != 20 {
		t.Errorf("Shanghai 时间转换错误: expected hour 20, got %d", shanghaiTime.Hour())
	}
}
