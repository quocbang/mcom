package mcom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateResourceOptions(t *testing.T) {
	assert := assert.New(t)

	const (
		stockInWarehouseID       = "W1"
		stockInWarehouseLocation = "A"
	)

	// with options.
	{
		o := ParseCreateMaterialResourcesOptions([]CreateMaterialResourcesOption{
			WithStockIn(Warehouse{
				ID:       stockInWarehouseID,
				Location: stockInWarehouseLocation,
			}),
		})
		assert.Equal(CreateMaterialResourcesOptions{
			Warehouse: Warehouse{
				ID:       stockInWarehouseID,
				Location: stockInWarehouseLocation,
			},
		}, o)
	}
	// no option.
	{
		expectedOptions := newCreateMaterialResourcesOptions()
		assert.Equal(expectedOptions, ParseCreateMaterialResourcesOptions([]CreateMaterialResourcesOption{}))
		assert.Equal(expectedOptions, ParseCreateMaterialResourcesOptions(nil))
	}
}
