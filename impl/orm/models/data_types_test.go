package models

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestEncryptedData_GormValue(t *testing.T) {
	assert := assert.New(t)

	assert.Implements(new(gorm.Valuer), EncryptedData{})

	const v = "test data"
	var data = EncryptedData(v)
	expr := data.GormValue(context.TODO(), nil)
	assert.Equal("?", expr.SQL)
	assert.Len(expr.Vars, 1)
	assert.NotEqual([]byte(v), expr.Vars[0])
}
