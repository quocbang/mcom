package models

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/callbacks"
)

func TestWorkOrderInformation_implementation(t *testing.T) {
	assert := assert.New(t)

	// good.
	{
		expected := WorkOrderInformation{
			BatchQuantityDetails: BatchQuantityDetails{
				QuantityForBatches: []decimal.Decimal{
					decimal.NewFromFloat(123.7),
					decimal.NewFromFloat(123.7),
					decimal.NewFromFloat(57.69),
				},
				FixedQuantity: &FixedQuantity{PlanQuantity: decimal.New(0, 0)},
				PlanQuantity:  &PlanQuantity{PlanQuantity: decimal.New(0, 0)},
			},
			Unit: "unit",
		}
		v, err := expected.Value()
		assert.NoError(err)

		var s WorkOrderInformation
		assert.NoError(s.Scan(v))
		assert.Equal(expected, s)
	}
}

func TestWorkOrder_BeforeCreate(t *testing.T) {
	assert := assert.New(t)

	assert.Implements(new(callbacks.BeforeCreateInterface), new(WorkOrder))

	var wo WorkOrder
	assert.NoError(wo.BeforeCreate(nil))
	assert.NotEmpty(wo.ID)
}

func TestCollectRecordDetail_implementation(t *testing.T) {
	assert := assert.New(t)

	// good.
	{
		expected := CollectRecordDetail{
			OperatorID: "O1",
			Quantity:   decimal.NewFromFloat(203.54),
			BatchCount: 5,
		}
		v, err := expected.Value()
		assert.NoError(err)

		var s CollectRecordDetail
		assert.NoError(s.Scan(v))
		// ensure reflect.DeepEqual to be properly.
		assert.Equal(expected, s)
	}
}
