package models

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func TestRecipeProcessConfigs_implementation(t *testing.T) {
	assert := assert.New(t)

	// good.
	{
		batchSize := decimal.NewFromInt(30)
		expected := RecipeProcessConfigs{
			{
				BatchSize: &batchSize,
				Stations:  []string{"S1", "S2"},
				Tools: []RecipeTool{
					{
						Type:     "3",
						ID:       "abc",
						Required: false,
					}, {
						Type:     "4",
						ID:       "xyz",
						Required: true,
					},
				},
				Steps: []RecipeProcessStep{
					{
						Materials: []RecipeMaterial{
							{
								ID:    "mID_1",
								Grade: "B",
								Value: RecipeMaterialParameter{
									High: types.Decimal.NewFromFloat64(7.5),
									Mid:  types.Decimal.NewFromInt16(5),
									Low:  types.Decimal.NewFromFloat64(2.5),
									Unit: "cm",
								},
								Site: "11",
							}, {
								ID:    "mID_2",
								Grade: "",
								Value: RecipeMaterialParameter{
									High: types.Decimal.NewFromFloat64(335.21),
									Mid:  types.Decimal.NewFromFloat64(329.47),
									Low:  types.Decimal.NewFromFloat64(325.56),
									Unit: "g",
								},
							},
						},
						Controls: []RecipeProperty{
							{
								Name: "ABC",
								Param: RecipePropertyParameter{
									High: nil,
									Mid:  nil,
									Low:  types.Decimal.NewFromFloat64(15.79),
									Unit: "kg",
								},
							}, {
								Name: "FFF",
								Param: RecipePropertyParameter{
									High: nil,
									Mid:  nil,
									Low:  nil,
									Unit: "",
								},
							},
						},
						Measurements: nil,
					}, {
						Materials: nil,
						Controls: []RecipeProperty{
							{
								Name: "ABC",
								Param: RecipePropertyParameter{
									High: nil,
									Mid:  nil,
									Low:  types.Decimal.NewFromFloat64(21.5),
									Unit: "kg",
								},
							}, {
								Name: "FFF",
								Param: RecipePropertyParameter{
									High: nil,
									Mid:  nil,
									Low:  nil,
									Unit: "",
								},
							},
						},
						Measurements: nil,
					},
				},
			},
		}
		v, err := expected.Value()
		assert.NoError(err)

		var s RecipeProcessConfigs
		assert.NoError(s.Scan(v))
		assert.Equal(expected, s)
	}
}

func TestRecipeProcesses_implementation(t *testing.T) {
	assert := assert.New(t)

	// good case.
	{
		expected := RecipeProcesses{
			{
				ReferenceOID: "abc",
				OptionalFlows: []RecipeOptionalFlow{
					{
						Name:           "retest",
						Processes:      []string{"a", "b", "c"},
						MaxRepetitions: 3,
					},
				},
			},
		}
		v, err := expected.Value()
		assert.NoError(err)

		var s RecipeProcesses
		assert.NoError(s.Scan(v))
		assert.Equal(expected, s)
	}
}
