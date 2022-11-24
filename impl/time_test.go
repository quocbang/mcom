package impl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_getWeekFirstAndLastDate(t *testing.T) {
	assert := assert.New(t)

	var (
		sunday    = time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC)
		monday    = time.Date(2021, time.January, 11, 0, 0, 0, 0, time.UTC)
		tuesday   = time.Date(2021, time.January, 12, 0, 0, 0, 0, time.UTC)
		wednesday = time.Date(2021, time.January, 13, 0, 0, 0, 0, time.UTC)
		thursday  = time.Date(2021, time.January, 14, 0, 0, 0, 0, time.UTC)
		friday    = time.Date(2021, time.January, 15, 0, 0, 0, 0, time.UTC)
		saturday  = time.Date(2021, time.January, 16, 0, 0, 0, 0, time.UTC)
	)

	for _, d := range []time.Time{
		sunday,
		monday,
		tuesday,
		wednesday,
		thursday,
		friday,
		saturday,
	} {
		first, last := getWeekFirstAndLastDate(d)
		assert.Equalf(sunday, first, "date: %s", d.Format(time.RFC3339))
		assert.Equalf(saturday, last, "date: %s", d.Format(time.RFC3339))
	}
}
