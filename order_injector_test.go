package mcom

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

func TestOrderableFieldsManager_register(t *testing.T) {
	assert := assert.New(t)
	ofm := newOrderableFieldsManager()
	{ // case register with a not existed field name.
		assert.EqualError(ofm.register(models.Account{}, OrderableField{
			FieldName:  "NotFound",
			ColumnName: "not_found",
		}), "field name not found, table: account, field: NotFound")
	}
	{ // good case.
		assert.NoError(ofm.register(models.PackRecord{}, OrderableField{
			FieldName:  "SerialNumber",
			ColumnName: "serial_number",
		}, OrderableField{
			FieldName:  "Quantity",
			ColumnName: "quantity",
		}))

		actual := ofm.get(models.PackRecord{})
		expected := map[string]struct{}(map[string]struct{}{"quantity": {}, "serial_number": {}})
		assert.Equal(expected, actual)
	}
}
