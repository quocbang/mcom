package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanJSON(t *testing.T) {
	assert := assert.New(t)

	type testStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	{
		const (
			id   = 88
			name = "（ ゜- ゜）つロ 乾杯~"
		)
		data := fmt.Sprintf(`{"id":%d,"name":"%s"}`, id, name)
		expected := testStruct{
			ID:   id,
			Name: name,
		}
		// string type.
		{
			var s testStruct
			assert.NoError(ScanJSON(data, &s))
			assert.Equal(expected, s)
		}
		// []byte type
		{
			var s testStruct
			assert.NoError(ScanJSON([]byte(data), &s))
			assert.Equal(expected, s)
		}
	}
	// bad data type.
	{
		var s testStruct
		err := ScanJSON(132, &s)
		assert.EqualError(err, "bad src type [int] for struct [*models.testStruct]")
		assert.Empty(s)
	}
}
