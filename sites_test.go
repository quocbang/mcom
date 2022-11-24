package mcom

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func Test_ToBoundResource(t *testing.T) {
	assert := assert.New(t)
	actual := BindMaterialResource{
		Material: models.Material{
			ID:    "A",
			Grade: "A",
		},
		Quantity:    types.Decimal.NewFromInt32(10),
		ResourceID:  "A",
		ProductType: "A",
		ExpiryTime:  0,
		Warehouse: Warehouse{
			ID:       "w",
			Location: "01",
		},
	}.ToBoundResource()
	expected := models.BoundResource{
		Material: &models.MaterialSite{
			Material: models.Material{
				ID:    "A",
				Grade: "A",
			},
			Quantity:    types.Decimal.NewFromInt32(10),
			ResourceID:  "A",
			ProductType: "A",
			ExpiryTime:  0,
		},
	}
	assert.Equal(expected, actual)
}
