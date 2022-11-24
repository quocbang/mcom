package impl

import (
	"context"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func Test_ProductPlan(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	dm, db, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}
	defer dm.Close()

	cm := newClearMaster(db, &models.Recipe{}, &models.ProductionPlan{}, &models.Department{}, &models.RecipeProcessDefinition{})
	assert.NoError(cm.Clear())

	// 統一時間
	timeNow := time.Now()
	monkey.Patch(time.Now, func() time.Time {
		return timeNow
	})
	defer monkey.UnpatchAll()

	err = createMockRecipeProcessDefinitionData(db)
	if !assert.NoError(err) {
		return
	}

	{ // CreateProductPlan: insufficient request.
		err := dm.CreateProductPlan(ctx, mcom.CreateProductionPlanRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // CreateProductPlan: quantity is zero.
		err := dm.CreateProductPlan(ctx, mcom.CreateProductionPlanRequest{
			Date: timeNow,
			Product: mcom.Product{
				ID:   testProductID,
				Type: testProductType,
			},
			DepartmentOID: testDepartmentID,
			Quantity:      decimal.Zero,
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "the Quantity is <= 0",
		}, err)
	}
	{ // CreateProductPlan: departmentID not found.
		err := dm.CreateProductPlan(ctx, mcom.CreateProductionPlanRequest{
			Date: timeNow,
			Product: mcom.Product{
				ID:   testProductID,
				Type: testProductType,
			},
			DepartmentOID: testDepartmentID,
			Quantity:      decimal.NewFromInt(600),
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_DEPARTMENT_NOT_FOUND,
		}, err)
	}
	{ // CreateDepartments: good case, add new department.
		err := dm.CreateDepartments(ctx, []string{testDepartmentID})
		assert.NoError(err)
	}
	{ // ListProductPlans: insufficient request.
		_, err := dm.ListProductPlans(ctx, mcom.ListProductPlansRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // CreateProductPlan: good case.
		err := dm.CreateProductPlan(ctx, mcom.CreateProductionPlanRequest{
			Date: timeNow,
			Product: mcom.Product{
				ID:   testProductID,
				Type: testProductType,
			},
			DepartmentOID: testDepartmentID,
			Quantity:      decimal.NewFromFloat(600.123456),
		})
		assert.NoError(err)

		// ListProductPlans: without reserved work order.
		result, err := dm.ListProductPlans(ctx, mcom.ListProductPlansRequest{
			Date:          timeNow,
			DepartmentOID: testDepartmentID,
			ProductType:   testProductType,
		})
		assert.NoError(err)
		assert.Equal(mcom.ListProductPlansReply{
			ProductPlans: []mcom.ProductPlan{{
				ProductID: testProductID,
				Quantity: mcom.ProductPlanQuantity{
					Daily:    decimal.NewFromFloat(600.123456),
					Week:     decimal.NewFromFloat(600.123456),
					Stock:    decimal.Zero,
					Reserved: decimal.Zero,
				},
			}},
		}, result)

		releasedAt := types.ToTimeNano(time.Now())
		assert.NoError(dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
			Recipes: []mcom.Recipe{{
				ID:      testRecipeID,
				Product: mcom.Product{ID: "product id", Type: "product type"},
				Version: mcom.RecipeVersion{
					Major:      "1",
					Minor:      "1",
					Stage:      "1",
					ReleasedAt: releasedAt,
				},
				Processes: []*mcom.Process{{
					OID:           "f310b071-fb63-464a-9cfc-f60d599a92d4",
					OptionalFlows: []*mcom.RecipeOptionalFlow{},
				}},
				ProcessDefinitions: []mcom.ProcessDefinition{{
					OID:     "f310b071-fb63-464a-9cfc-f60d599a92d4",
					Name:    "name",
					Type:    "type",
					Configs: []*mcom.RecipeProcessConfig{},
					Output:  mcom.OutputProduct{},
				}},
			}},
		}))

		// CreateWorkOrders: good case.
		workOrderIDs, err := dm.CreateWorkOrders(ctx, mcom.CreateWorkOrdersRequest{
			WorkOrders: []mcom.CreateWorkOrder{
				{
					ProcessOID:    testProcessOID,
					ProcessName:   testProcessName,
					ProcessType:   testProcessType,
					RecipeID:      testRecipeID,
					DepartmentOID: testDepartmentID,
					Station:       testWorkOrderStation,
					Date:          timeNow,
					Unit:          "kg",
					BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
					}),
					Parent: "parentWorkOrder",
				},
			},
		})
		assert.NoError(err)

		// ListProductPlans: with reserved quantity.
		result, err = dm.ListProductPlans(ctx, mcom.ListProductPlansRequest{
			Date:          timeNow,
			DepartmentOID: testDepartmentID,
			ProductType:   testProductType,
		})
		assert.NoError(err)
		assert.Equal(mcom.ListProductPlansReply{
			ProductPlans: []mcom.ProductPlan{{
				ProductID: testProductID,
				Quantity: mcom.ProductPlanQuantity{
					Daily:    decimal.NewFromFloat(600.123456),
					Week:     decimal.NewFromFloat(600.123456),
					Stock:    decimal.Zero,
					Reserved: decimal.NewFromInt(1200),
				},
			}},
		}, result)

		// delete mock workorder.
		err = deleteMockWorkOrderData(db, workOrderIDs.IDs)
		assert.NoError(err)
	}
	assert.NoError(cm.Clear())
}
