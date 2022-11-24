package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimitaryHour_TableName(t *testing.T) {
	assert := assert.New(t)

	expected := "limitary_hour"

	var s LimitaryHour
	assert.Equal(expected, s.TableName())
}
