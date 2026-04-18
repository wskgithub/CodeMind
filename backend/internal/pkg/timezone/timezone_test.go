package timezone

import (
	"strings"
	"testing"
	"time"
)

// TestShanghaiInitialization tests Shanghai timezone initialization.
func TestShanghaiInitialization(t *testing.T) {
	tests := []struct {
		validate func(*time.Location) bool
		name     string
	}{
		{
			name: "Shanghai timezone should not be nil",
			validate: func(loc *time.Location) bool {
				return loc != nil
			},
		},
		{
			name: "Shanghai timezone name should be Asia/Shanghai or CST",
			validate: func(loc *time.Location) bool {
				// May be "Asia/Shanghai" or "CST" (when manually created)
				name := loc.String()
				return name == "Asia/Shanghai" || name == "CST"
			},
		},
		{
			name: "Shanghai timezone offset should be +0800",
			validate: func(loc *time.Location) bool {
				// Use reference time to check offset
				ref := time.Date(2024, 1, 1, 12, 0, 0, 0, loc)
				_, offset := ref.Zone()
				return offset == 8*60*60 // +8 hours = 28800 seconds
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.validate(Shanghai) {
				t.Errorf("Shanghai timezone validation failed: %s", tt.name)
			}
		})
	}
}

// TestNow tests Now() returns Shanghai timezone time.
func TestNow(t *testing.T) {
	tests := []struct {
		validate func(time.Time) bool
		name     string
	}{
		{
			name: "returned time timezone should be Shanghai",
			validate: func(tm time.Time) bool {
				return tm.Location().String() == Shanghai.String()
			},
		},
		{
			name: "returned time should not be zero value",
			validate: func(tm time.Time) bool {
				return !tm.IsZero()
			},
		},
		{
			name: "returned time should be within reasonable range (±1 minute)",
			validate: func(tm time.Time) bool {
				// Get current UTC time and convert to Shanghai timezone
				nowUTC := time.Now().UTC()
				diff := tm.Sub(nowUTC)
				// Time difference should be between -1 and +1 minute
				return diff > -time.Minute && diff < time.Minute
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := Now()
			if !tt.validate(now) {
				t.Errorf("Now() validation failed: %s, got: %v", tt.name, now)
			}
		})
	}
}

// TestToday tests Today() returns midnight today in UTC.
func TestToday(t *testing.T) {
	tests := []struct {
		validate func(time.Time) bool
		name     string
	}{
		{
			name: "returned time timezone should be UTC",
			validate: func(tm time.Time) bool {
				return tm.Location().String() == "UTC"
			},
		},
		{
			name: "time should be midnight (00:00:00)",
			validate: func(tm time.Time) bool {
				return tm.Hour() == 0 && tm.Minute() == 0 && tm.Second() == 0 && tm.Nanosecond() == 0
			},
		},
		{
			name: "date should match current Shanghai timezone date",
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
				t.Errorf("Today() validation failed: %s, got: %v", tt.name, today)
			}
		})
	}
}

// TestTodayStr tests TodayStr() format correctness.
func TestTodayStr(t *testing.T) {
	tests := []struct {
		validate func(string) bool
		name     string
	}{
		{
			name: "format should be YYYY-MM-DD",
			validate: func(s string) bool {
				// Check format: 4 digits - 2 digits - 2 digits
				parts := strings.Split(s, "-")
				if len(parts) != 3 {
					return false
				}
				return len(parts[0]) == 4 && len(parts[1]) == 2 && len(parts[2]) == 2
			},
		},
		{
			name: "date should match current Shanghai timezone date",
			validate: func(s string) bool {
				expected := Now().Format("2006-01-02")
				return s == expected
			},
		},
		{
			name: "should be parseable as time",
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
				t.Errorf("TodayStr() validation failed: %s, got: %s", tt.name, todayStr)
			}
		})
	}
}

// TestMonthStr tests MonthStr() format correctness.
func TestMonthStr(t *testing.T) {
	tests := []struct {
		validate func(string) bool
		name     string
	}{
		{
			name: "format should be YYYY-MM",
			validate: func(s string) bool {
				// Check format: 4 digits - 2 digits
				parts := strings.Split(s, "-")
				if len(parts) != 2 {
					return false
				}
				return len(parts[0]) == 4 && len(parts[1]) == 2
			},
		},
		{
			name: "month should match current Shanghai timezone month",
			validate: func(s string) bool {
				expected := Now().Format("2006-01")
				return s == expected
			},
		},
		{
			name: "should be parseable as time",
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
				t.Errorf("MonthStr() validation failed: %s, got: %s", tt.name, monthStr)
			}
		})
	}
}

// TestCrossDayTimezoneHandling tests cross-day timezone handling.
func TestCrossDayTimezoneHandling(t *testing.T) {
	tests := []struct {
		inputTime      time.Time
		name           string
		expectedHour   int
		expectedOffset int
	}{
		{
			name:           "UTC midnight to Shanghai timezone",
			inputTime:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedHour:   8, // UTC 00:00 = Shanghai 08:00
			expectedOffset: 8 * 60 * 60,
		},
		{
			name:           "UTC 16:00 to Shanghai timezone",
			inputTime:      time.Date(2024, 1, 15, 16, 0, 0, 0, time.UTC),
			expectedHour:   0, // UTC 16:00 = Shanghai 00:00 (next day)
			expectedOffset: 8 * 60 * 60,
		},
		{
			name:           "New York time to Shanghai timezone",
			inputTime:      time.Date(2024, 1, 15, 12, 0, 0, 0, time.FixedZone("EST", -5*60*60)),
			expectedHour:   1, // EST 12:00 = Shanghai 01:00 (next day)
			expectedOffset: 8 * 60 * 60,
		},
		{
			name:           "cross-day boundary - UTC 23:59",
			inputTime:      time.Date(2024, 1, 15, 23, 59, 0, 0, time.UTC),
			expectedHour:   7, // UTC 23:59 = Shanghai 07:59 (next day)
			expectedOffset: 8 * 60 * 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert input time to Shanghai timezone
			shanghaiTime := tt.inputTime.In(Shanghai)

			// Verify hours
			if shanghaiTime.Hour() != tt.expectedHour {
				t.Errorf("hour mismatch: expected %d, got %d", tt.expectedHour, shanghaiTime.Hour())
			}

			// Verify offset
			_, offset := shanghaiTime.Zone()
			if offset != tt.expectedOffset {
				t.Errorf("timezone offset mismatch: expected %d, got %d", tt.expectedOffset, offset)
			}
		})
	}
}

// Note: China hasn't used DST since 1991, but the timezone library should handle DST correctly for other regions.
func TestDaylightSavingTime(t *testing.T) {
	tests := []struct {
		testTime       time.Time
		name           string
		locationName   string
		expectedOffset int
		expectDST      bool
	}{
		{
			name:           "Shanghai timezone does not use DST",
			locationName:   "Asia/Shanghai",
			testTime:       time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC), // summer
			expectDST:      false,
			expectedOffset: 8 * 60 * 60, // always +8
		},
		{
			name:           "Shanghai timezone winter does not use DST",
			locationName:   "Asia/Shanghai",
			testTime:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC), // winter
			expectDST:      false,
			expectedOffset: 8 * 60 * 60, // always +8
		},
		{
			name:           "US Eastern daylight saving time",
			locationName:   "America/New_York",
			testTime:       time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC), // summer
			expectDST:      true,
			expectedOffset: -4 * 60 * 60, // EDT = UTC-4
		},
		{
			name:           "US Eastern standard time",
			locationName:   "America/New_York",
			testTime:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC), // winter
			expectDST:      false,
			expectedOffset: -5 * 60 * 60, // EST = UTC-5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := time.LoadLocation(tt.locationName)
			if err != nil {
				t.Skipf("unable to load timezone %s: %v", tt.locationName, err)
				return
			}

			localTime := tt.testTime.In(loc)
			_, offset := localTime.Zone()

			if offset != tt.expectedOffset {
				t.Errorf("timezone offset mismatch: expected %d, got %d", tt.expectedOffset, offset)
			}
		})
	}
}

// TestTodayConsistency tests consistency between Today() and TodayStr().
func TestTodayConsistency(t *testing.T) {
	// Time from Today() should match the date represented by TodayStr()
	today := Today()
	todayStr := TodayStr()

	// TodayStr() is based on Shanghai timezone, Today() is also based on Shanghai timezone current date
	// But Today() returns UTC-wrapped time
	// Compare date values, not string representations
	nowShanghai := Now()
	expectedDate := time.Date(nowShanghai.Year(), nowShanghai.Month(), nowShanghai.Day(), 0, 0, 0, 0, time.UTC)

	if !today.Equal(expectedDate) {
		t.Errorf("Today() date mismatch: expected %v, got %v", expectedDate, today)
	}

	// TodayStr should represent the current date in Shanghai timezone
	expectedStr := nowShanghai.Format("2006-01-02")
	if todayStr != expectedStr {
		t.Errorf("TodayStr() mismatch: expected %s, got %s", expectedStr, todayStr)
	}
}

// TestMonthStrConsistency tests consistency between MonthStr() and Now().
func TestMonthStrConsistency(t *testing.T) {
	now := Now()
	monthStr := MonthStr()
	expectedStr := now.Format("2006-01")

	if monthStr != expectedStr {
		t.Errorf("MonthStr() mismatch: expected %s, got %s", expectedStr, monthStr)
	}
}

// BenchmarkNow benchmarks Now().
func BenchmarkNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Now()
	}
}

// BenchmarkToday benchmarks Today().
func BenchmarkToday(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Today()
	}
}

// BenchmarkTodayStr benchmarks TodayStr().
func BenchmarkTodayStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		TodayStr()
	}
}

// BenchmarkMonthStr benchmarks MonthStr().
func BenchmarkMonthStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MonthStr()
	}
}

// TestLoadShanghaiLocation tests both branches of loadShanghaiLocation.
func TestLoadShanghaiLocation(t *testing.T) {
	tests := []struct {
		name           string
		tzName         string
		expectedName   string
		expectedOffset int
	}{
		{
			name:           "load valid Asia/Shanghai timezone",
			tzName:         "Asia/Shanghai",
			expectedName:   "Asia/Shanghai",
			expectedOffset: 8 * 60 * 60,
		},
		{
			name:           "loading invalid timezone should fall back to CST",
			tzName:         "Invalid/Timezone",
			expectedName:   "CST",
			expectedOffset: 8 * 60 * 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := loadShanghaiLocation(tt.tzName)

			if loc == nil {
				t.Fatal("loadShanghaiLocation() returned nil")
			}

			// Verify timezone name
			name := loc.String()
			if name != tt.expectedName {
				t.Errorf("timezone name error: expected %s, got %s", tt.expectedName, name)
			}

			// Verify offset
			ref := time.Date(2024, 1, 1, 12, 0, 0, 0, loc)
			_, offset := ref.Zone()
			if offset != tt.expectedOffset {
				t.Errorf("timezone offset error: expected %d, got %d", tt.expectedOffset, offset)
			}
		})
	}
}

// Verify manually created CST timezone is functionally equivalent to loaded Asia/Shanghai.
func TestShanghaiTimezoneFallback(t *testing.T) {
	// Create manual CST timezone (simulate init fallback logic)
	cstZone := time.FixedZone("CST", 8*60*60)
	if cstZone == nil {
		t.Error("unable to create CST timezone")
	}

	// Verify manually created timezone name
	if cstZone.String() != "CST" {
		t.Errorf("CST timezone name error: expected CST, got %s", cstZone.String())
	}

	// Verify manually created timezone converts time correctly
	utcTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	cstTime := utcTime.In(cstZone)

	// UTC 12:00 = CST 20:00
	if cstTime.Hour() != 20 {
		t.Errorf("CST time conversion error: expected hour 20, got %d", cstTime.Hour())
	}

	// Verify Shanghai timezone is correctly initialized (whether loaded or manually created)
	if Shanghai == nil {
		t.Error("Shanghai timezone should not be nil")
	}

	// Verify Shanghai and CST timezones are functionally equivalent
	shanghaiTime := utcTime.In(Shanghai)
	if shanghaiTime.Hour() != 20 {
		t.Errorf("Shanghai time conversion error: expected hour 20, got %d", shanghaiTime.Hour())
	}
}
