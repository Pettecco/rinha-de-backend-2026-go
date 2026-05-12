package vec

import (
	"math"
	"time"
)

func parseISO8601Z(s string) (year, month, day, hour, min, sec int, ok bool) {
	if len(s) != 20 || s[4] != '-' || s[7] != '-' || s[10] != 'T' || s[13] != ':' || s[16] != ':' || s[19] != 'Z' {
		return 0, 0, 0, 0, 0, 0, false
	}
	year = parseInt4(s[0:4])
	month = parseInt2(s[5:7])
	day = parseInt2(s[8:10])
	hour = parseInt2(s[11:13])
	min = parseInt2(s[14:16])
	sec = parseInt2(s[17:19])
	if month < 1 || month > 12 || day < 1 || day > 31 || hour > 23 || min > 59 || sec > 59 {
		return 0, 0, 0, 0, 0, 0, false
	}
	return year, month, day, hour, min, sec, true
}

func parseInt2(s string) int {
	if len(s) != 2 {
		return 0
	}
	return int(s[0]-'0')*10 + int(s[1]-'0')
}

func parseInt4(s string) int {
	if len(s) != 4 {
		return 0
	}
	return int(s[0]-'0')*1000 + int(s[1]-'0')*100 + int(s[2]-'0')*10 + int(s[3]-'0')
}

func ExtractHour(ts string) int {
	if len(ts) < 13 {
		return 0
	}
	h := parseInt2(ts[11:13])
	if h < 0 || h > 23 {
		return 0
	}
	return h
}

func ExtractDayOfWeek(ts string) int {
	y, m, d, _, _, _, ok := parseISO8601Z(ts)
	if !ok {
		return 0
	}
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	w := int(t.Weekday())
	if w == 0 {
		return 6
	}
	return w - 1
}

func MinutesBetween(ts1, ts2 string) float64 {
	y1, m1, d1, h1, min1, s1, ok1 := parseISO8601Z(ts1)
	y2, m2, d2, h2, min2, s2, ok2 := parseISO8601Z(ts2)
	if !ok1 || !ok2 {
		return math.MaxFloat64
	}
	t1 := time.Date(y1, time.Month(m1), d1, h1, min1, s1, 0, time.UTC)
	t2 := time.Date(y2, time.Month(m2), d2, h2, min2, s2, 0, time.UTC)
	diff := t2.Sub(t1)
	if diff < 0 {
		diff = -diff
	}
	return diff.Minutes()
}
