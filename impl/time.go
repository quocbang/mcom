package impl

import (
	"time"
)

func getWeekFirstAndLastDate(tm time.Time) (first, last time.Time) {
	wd := tm.Weekday()
	if interval := time.Sunday - wd; interval == 0 {
		first = tm
	} else {
		first = tm.AddDate(0, 0, int(interval))
	}
	if interval := time.Saturday - wd; interval == 0 {
		last = tm
	} else {
		last = tm.AddDate(0, 0, int(interval))
	}
	return
}
