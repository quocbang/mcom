package impl

import (
	"context"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	pbWorkorder "gitlab.kenda.com.tw/kenda/commons/v2/proto/golang/mes/v2/workorder"
	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
	"gitlab.kenda.com.tw/kenda/mcom/utils/workorder"
)

const (
	testWorkOrderUser = "123456"

	testWorkOrderDepartmentOID = "T1234"
	testWorkOrderStation       = "gundam"
	testWorkOrderStationB      = "evangelion"
	testWorkOrderID            = "ASDFGHJKL123456789"

	testProcessOIDNotExist = "5ab44172-8bc9-46c9-b4d6-1806dc97115c"

	testFeedResourceID     = "123456"
	testCollectResourceOID = "32144172-8bc9-46c9-b4d6-1806dc97115c"
)

func deleteMockWorkOrderData(db *gorm.DB, IDs []string) error {
	if err := db.Where(` id IN ? `, IDs).Delete(&models.WorkOrder{}).Error; err != nil {
		return err
	}
	return nil
}

func Test_WorkOrder(t *testing.T) {
	assert := assert.New(t)
	ctx := commonsCtx.WithUserID(context.Background(), testWorkOrderUser)

	// 統一時間
	timeNow := time.Now().Local()
	monkey.Patch(time.Now, func() time.Time {
		return timeNow
	})
	defer monkey.UnpatchAll()

	dm, db, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}
	defer dm.Close()

	cm := newClearMaster(db, models.Recipe{}, models.RecipeProcessDefinition{}, models.WorkOrder{})
	assert.NoError(cm.Clear())

	// todo: 改ListWorkOrder的時候順便改掉 createMockRecipeProcessDefinitionData.
	err = createMockRecipeProcessDefinitionData(db)
	if !assert.NoError(err) {
		return
	}

	{ // GetWorkOrder: insufficient request.
		_, err := dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{ID: ""})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // GetWorkOrder: Code_WORKORDER_NOT_FOUND.
		_, err := dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{ID: testWorkOrderID})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_WORKORDER_NOT_FOUND,
		}, err)
	}
	{ // ListWorkOrders: insufficient request.
		// nolint:staticcheck // SA1019 for test coverage.
		_, err := dm.ListWorkOrders(ctx, mcom.ListWorkOrdersRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // CreateWorkOrders: nil request.
		_, err := dm.CreateWorkOrders(ctx, mcom.CreateWorkOrdersRequest{WorkOrders: nil})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST})
	}
	{ // CreateWorkOrders: insufficient request.
		_, err := dm.CreateWorkOrders(ctx, mcom.CreateWorkOrdersRequest{
			WorkOrders: []mcom.CreateWorkOrder{
				{
					ProcessOID:      "",
					RecipeID:        "",
					DepartmentOID:   "",
					Date:            time.Time{},
					BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{}),
				},
			}})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // CreateWorkOrders: bad case, process type did not match.
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

		ids, err := dm.CreateWorkOrders(ctx, mcom.CreateWorkOrdersRequest{
			WorkOrders: []mcom.CreateWorkOrder{
				{
					ProcessOID:      testProcessOID,
					RecipeID:        testRecipeID,
					ProcessName:     testProcessName,
					ProcessType:     "not_exist",
					Status:          pbWorkorder.Status_PENDING,
					DepartmentOID:   testWorkOrderDepartmentOID,
					Station:         testWorkOrderStation,
					Sequence:        0,
					BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200)}),
					Parent:          testWorkOrderDepartmentOID,
					Date:            timeNow,
					Unit:            "kg",
					Note:            testProcessOIDNotExist,
				},
			}})
		assert.ErrorIs(err, mcomErr.Error{
			Code:    mcomErr.Code_PROCESS_NOT_FOUND,
			Details: "recipe id: recipe_id, process name: process_name, process type: not_exist",
		})

		// ListWorkOrder error.
		// nolint:staticcheck // SA1019 for test coverage.
		_, err = dm.ListWorkOrders(ctx, mcom.ListWorkOrdersRequest{
			Date:    timeNow,
			Station: testWorkOrderStation,
		})
		assert.NoError(err)

		// delete mock work order.
		err = deleteMockWorkOrderData(db, ids.IDs)
		assert.NoError(err)
	}
	{ // GetWorkOrder: good case.
		ids, err := dm.CreateWorkOrders(ctx, mcom.CreateWorkOrdersRequest{
			WorkOrders: []mcom.CreateWorkOrder{
				{
					ProcessOID:      testProcessOID,
					RecipeID:        testRecipeID,
					ProcessName:     testProcessName,
					ProcessType:     testProcessType,
					DepartmentOID:   testWorkOrderDepartmentOID,
					Station:         testWorkOrderStation,
					BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200)}),
					Parent:          "parentWorkOrder",
					Date:            time.Now(),
					Unit:            "kg",
				},
			}})
		assert.NoError(err)
		assert.Len(ids.IDs, 1)
		result, err := dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{ID: ids.IDs[0]})
		assert.NoError(err)

		resvDate, err := parseDate(timeNow)
		assert.NoError(err)

		expected := mcom.GetWorkOrderReply{
			ID: ids.IDs[0],
			Product: mcom.Product{
				ID:   "product_id",
				Type: "RUBBER",
			},
			Process: mcom.WorkOrderProcess{
				OID:  testProcessOID,
				Name: testProcessName,
				Type: testProcessType,
			},
			RecipeID:      testRecipeID,
			DepartmentOID: testWorkOrderDepartmentOID,
			Station:       testWorkOrderStation,
			Sequence:      1,
			Date:          resvDate,
			Unit:          "kg",
			BatchQuantityDetails: models.BatchQuantityDetails{
				QuantityForBatches: []decimal.Decimal{
					decimal.NewFromInt(200),
					decimal.NewFromInt(200),
					decimal.NewFromInt(200),
					decimal.NewFromInt(200),
					decimal.NewFromInt(200),
					decimal.NewFromInt(200),
				},
				BatchQuantityType: workorder.BatchSize_PER_BATCH_QUANTITIES,
			},

			CurrentBatch:      0,
			CollectedSequence: 0,
			UpdatedBy:         testWorkOrderUser,
			UpdatedAt:         time.Unix(0, timeNow.UnixNano()).Local(),
			InsertedBy:        testWorkOrderUser,
			InsertedAt:        time.Unix(0, timeNow.UnixNano()).Local(),
			Parent:            "parentWorkOrder",
			CollectedQuantity: decimal.Zero,
		}

		assert.Equal(expected, result)

		// delete mock work order.
		err = deleteMockWorkOrderData(db, ids.IDs)
		assert.NoError(err)
	}
	{ // GetWorkOrder: good case with fixed quantity.
		ids, err := dm.CreateWorkOrders(ctx, mcom.CreateWorkOrdersRequest{
			WorkOrders: []mcom.CreateWorkOrder{
				{
					ProcessOID:      testProcessOID,
					RecipeID:        testRecipeID,
					ProcessName:     testProcessName,
					ProcessType:     testProcessType,
					DepartmentOID:   testWorkOrderDepartmentOID,
					Station:         testWorkOrderStation,
					BatchesQuantity: mcom.NewFixedQuantity(5, decimal.NewFromInt32(10)),
					Parent:          "parentWorkOrder",
					Date:            time.Now(),
					Unit:            "kg",
				},
			}})
		assert.NoError(err)
		assert.Len(ids.IDs, 1)
		result, err := dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{ID: ids.IDs[0]})
		assert.NoError(err)

		resvDate, err := parseDate(timeNow)
		assert.NoError(err)

		expected := mcom.GetWorkOrderReply{
			ID: ids.IDs[0],
			Product: mcom.Product{
				ID:   "product_id",
				Type: "RUBBER",
			},
			Process: mcom.WorkOrderProcess{
				OID:  testProcessOID,
				Name: testProcessName,
				Type: testProcessType,
			},
			RecipeID:      testRecipeID,
			DepartmentOID: testWorkOrderDepartmentOID,
			Station:       testWorkOrderStation,
			Sequence:      1,
			Date:          resvDate,
			Unit:          "kg",
			BatchQuantityDetails: models.BatchQuantityDetails{
				BatchQuantityType: workorder.BatchSize_FIXED_QUANTITY,
				FixedQuantity: &models.FixedQuantity{
					BatchCount:   5,
					PlanQuantity: decimal.NewFromInt32(10),
				},
			},
			CurrentBatch:      0,
			CollectedSequence: 0,
			UpdatedBy:         testWorkOrderUser,
			UpdatedAt:         time.Unix(0, timeNow.UnixNano()).Local(),
			InsertedBy:        testWorkOrderUser,
			InsertedAt:        time.Unix(0, timeNow.UnixNano()).Local(),
			Parent:            "parentWorkOrder",
			CollectedQuantity: decimal.Zero,
		}

		assert.Equal(expected, result)

		// delete mock work order.
		err = deleteMockWorkOrderData(db, ids.IDs)
		assert.NoError(err)
	}
	{ // CreateWorkOrders: good case.
		ids, err := dm.CreateWorkOrders(ctx, mcom.CreateWorkOrdersRequest{
			WorkOrders: []mcom.CreateWorkOrder{
				{
					ProcessOID:    testProcessOID,
					ProcessName:   testProcessName,
					ProcessType:   testProcessType,
					RecipeID:      testRecipeID,
					DepartmentOID: testWorkOrderDepartmentOID,
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
				},
			}})
		assert.NoError(err)

		for _, id := range ids.IDs {
			// nolint:staticcheck // SA1019 for test coverage.
			result, err := dm.ListWorkOrders(ctx, mcom.ListWorkOrdersRequest{
				ID:      id,
				Date:    timeNow,
				Station: testWorkOrderStation,
			})
			assert.NoError(err)

			resvDate, err := parseDate(timeNow)
			assert.NoError(err)

			expected := []mcom.GetWorkOrderReply{{
				ID: id,
				Product: mcom.Product{
					ID:   "product_id",
					Type: "RUBBER",
				},
				Process: mcom.WorkOrderProcess{
					OID:  testProcessOID,
					Name: testProcessName,
					Type: testProcessType,
				},
				RecipeID:      testRecipeID,
				DepartmentOID: testWorkOrderDepartmentOID,
				Station:       testWorkOrderStation,
				Sequence:      1,
				Date:          resvDate,
				Unit:          "kg",
				BatchQuantityDetails: models.BatchQuantityDetails{
					BatchQuantityType: workorder.BatchSize_PER_BATCH_QUANTITIES,
					QuantityForBatches: []decimal.Decimal{
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
					},
				},

				CurrentBatch:      0,
				CollectedSequence: 0,
				UpdatedBy:         testWorkOrderUser,
				UpdatedAt:         time.Unix(0, timeNow.UnixNano()).Local(),
				InsertedBy:        testWorkOrderUser,
				InsertedAt:        time.Unix(0, timeNow.UnixNano()).Local(),
				CollectedQuantity: decimal.Zero,
			}}

			assert.Equal(len(expected), len(result.WorkOrders))
			for i, r := range expected {
				assert.Containsf(result.WorkOrders, r, "listWorkOrder test case failed, index: %d", i)
			}
		}

		// delete mock work order.
		err = deleteMockWorkOrderData(db, ids.IDs)
		assert.NoError(err)
	}
	{ // CreateWorkOrders: good case. create multiple work orders
		ids, err := dm.CreateWorkOrders(ctx, mcom.CreateWorkOrdersRequest{
			WorkOrders: []mcom.CreateWorkOrder{
				{
					ProcessOID:    testProcessOID,
					ProcessName:   testProcessName,
					ProcessType:   testProcessType,
					RecipeID:      testRecipeID,
					DepartmentOID: testWorkOrderDepartmentOID,
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
				},
				{
					ProcessOID:    testProcessOID,
					ProcessName:   testProcessFlowName,
					ProcessType:   testProcessType,
					RecipeID:      testRecipeID,
					DepartmentOID: testWorkOrderDepartmentOID,
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
				},
			}})
		assert.NoError(err)

		rep, err := dm.ListWorkOrdersByDuration(ctx, mcom.ListWorkOrdersByDurationRequest{
			Since:   time.Date(1911, 10, 10, 0, 0, 0, 0, time.UTC),
			Station: testWorkOrderStation,
		})
		assert.NoError(err)
		resvDate, err := parseDate(timeNow)
		assert.NoError(err)
		expected := mcom.ListWorkOrdersByDurationReply{
			Contents: []mcom.GetWorkOrderReply{
				{
					ID: ids.IDs[0],
					Product: mcom.Product{
						ID:   "product_id",
						Type: "RUBBER",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: testProcessName,
						Type: testProcessType,
					},
					RecipeID:      testRecipeID,
					DepartmentOID: testWorkOrderDepartmentOID,
					Station:       testWorkOrderStation,
					Sequence:      1,
					Date:          resvDate,
					Unit:          "kg",
					BatchQuantityDetails: models.BatchQuantityDetails{
						BatchQuantityType: workorder.BatchSize_PER_BATCH_QUANTITIES,
						QuantityForBatches: []decimal.Decimal{
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
						},
					},

					CurrentBatch:      0,
					CollectedSequence: 0,
					UpdatedBy:         testWorkOrderUser,
					UpdatedAt:         time.Unix(0, timeNow.UnixNano()).Local(),
					InsertedBy:        testWorkOrderUser,
					InsertedAt:        time.Unix(0, timeNow.UnixNano()).Local(),
					CollectedQuantity: decimal.Zero,
				},
				{
					ID: ids.IDs[1],
					Product: mcom.Product{
						ID:   "product_id",
						Type: "RUBBER",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: testProcessFlowName,
						Type: testProcessType,
					},
					RecipeID:      testRecipeID,
					DepartmentOID: testWorkOrderDepartmentOID,
					Station:       testWorkOrderStation,
					Sequence:      2,
					Date:          resvDate,
					Unit:          "kg",
					BatchQuantityDetails: models.BatchQuantityDetails{
						BatchQuantityType: workorder.BatchSize_PER_BATCH_QUANTITIES,
						QuantityForBatches: []decimal.Decimal{
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
						},
					},

					CurrentBatch:      0,
					CollectedSequence: 0,
					UpdatedBy:         testWorkOrderUser,
					UpdatedAt:         time.Unix(0, timeNow.UnixNano()).Local(),
					InsertedBy:        testWorkOrderUser,
					InsertedAt:        time.Unix(0, timeNow.UnixNano()).Local(),
					CollectedQuantity: decimal.Zero,
				},
			},
		}

		assert.Equal(expected, rep)

		// delete mock work order.
		err = deleteMockWorkOrderData(db, ids.IDs)
		assert.NoError(err)
	}
	{ // UpdateWorkOrders: insufficient request.
		err := dm.UpdateWorkOrders(ctx, mcom.UpdateWorkOrdersRequest{Orders: []mcom.UpdateWorkOrder{{ID: ""}}})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "Key: 'UpdateWorkOrdersRequest.Orders[0].ID' Error:Field validation for 'ID' failed on the 'required' tag",
		}, err)
	}
	{ // UpdateWorkOrders: good case.
		newResvDate := timeNow.Add(24 * time.Hour)
		// create workorder.
		ids, err := dm.CreateWorkOrders(ctx, mcom.CreateWorkOrdersRequest{
			WorkOrders: []mcom.CreateWorkOrder{
				{
					ProcessOID:    testProcessOID,
					ProcessName:   testProcessName,
					ProcessType:   testProcessType,
					RecipeID:      testRecipeID,
					DepartmentOID: testWorkOrderDepartmentOID,
					Station:       testWorkOrderStationB,
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
				},
				{
					ProcessOID:    testProcessOID,
					ProcessName:   testProcessName,
					ProcessType:   testProcessType,
					RecipeID:      testRecipeID,
					DepartmentOID: testWorkOrderDepartmentOID,
					Station:       testWorkOrderStation,
					Date:          newResvDate,
					Unit:          "kg",
					BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
					}),
				},
			}})
		assert.NoError(err)
		{ // UpdateWorkOrders: change reserve station.
			err = dm.UpdateWorkOrders(ctx, mcom.UpdateWorkOrdersRequest{
				Orders: []mcom.UpdateWorkOrder{{
					ID:      ids.IDs[0],
					Station: testWorkOrderStation,
					Date:    newResvDate,
				}}})
			assert.NoError(err)

			result, err := dm.ListWorkOrders(ctx, mcom.ListWorkOrdersRequest{
				ID:      ids.IDs[0],
				Date:    newResvDate,
				Station: testWorkOrderStation,
			})

			resvDate, _ := parseDate(newResvDate)
			expected := mcom.ListWorkOrdersReply{
				WorkOrders: []mcom.GetWorkOrderReply{{
					ID: ids.IDs[0],
					Product: mcom.Product{
						ID:   "product_id",
						Type: "RUBBER",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: testProcessName,
						Type: testProcessType,
					},
					RecipeID:      testRecipeID,
					DepartmentOID: testWorkOrderDepartmentOID,
					Station:       testWorkOrderStation,
					Sequence:      1,
					Date:          resvDate,
					Unit:          "kg",
					BatchQuantityDetails: models.BatchQuantityDetails{
						BatchQuantityType: workorder.BatchSize_PER_BATCH_QUANTITIES,
						QuantityForBatches: []decimal.Decimal{
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
							decimal.NewFromInt(200),
						},
					},
					CurrentBatch:      0,
					CollectedSequence: 0,
					UpdatedBy:         testWorkOrderUser,
					UpdatedAt:         time.Unix(0, timeNow.UnixNano()),
					InsertedBy:        testWorkOrderUser,
					InsertedAt:        time.Unix(0, timeNow.UnixNano()),
					CollectedQuantity: decimal.Zero,
				}}}
			assert.NoError(err)
			assert.Equal(expected, result)
		}
		{ // UpdateWorkOrders: status.
			err = dm.UpdateWorkOrders(ctx, mcom.UpdateWorkOrdersRequest{
				Orders: []mcom.UpdateWorkOrder{{
					ID:     ids.IDs[0],
					Status: pbWorkorder.Status_ACTIVE,
				}}})
			assert.NoError(err)

			result, err := dm.ListWorkOrders(ctx, mcom.ListWorkOrdersRequest{
				ID:      ids.IDs[0],
				Date:    newResvDate,
				Station: testWorkOrderStation,
			})

			resvDate, _ := parseDate(newResvDate)
			expected := mcom.ListWorkOrdersReply{WorkOrders: []mcom.GetWorkOrderReply{{
				ID: ids.IDs[0],
				Product: mcom.Product{
					ID:   "product_id",
					Type: "RUBBER",
				},
				Process: mcom.WorkOrderProcess{
					OID:  testProcessOID,
					Name: testProcessName,
					Type: testProcessType,
				},
				RecipeID:      testRecipeID,
				DepartmentOID: testWorkOrderDepartmentOID,
				Station:       testWorkOrderStation,
				Sequence:      1,
				Status:        pbWorkorder.Status_ACTIVE,
				Date:          resvDate,
				Unit:          "kg",
				BatchQuantityDetails: models.BatchQuantityDetails{
					BatchQuantityType: workorder.BatchSize_PER_BATCH_QUANTITIES,
					QuantityForBatches: []decimal.Decimal{
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
					},
				},
				CurrentBatch:      0,
				CollectedSequence: 0,
				UpdatedBy:         testWorkOrderUser,
				UpdatedAt:         time.Unix(0, timeNow.UnixNano()),
				InsertedBy:        testWorkOrderUser,
				InsertedAt:        time.Unix(0, timeNow.UnixNano()),
				CollectedQuantity: decimal.Zero,
			}}}
			assert.NoError(err)
			assert.Equal(expected, result)
		}
		{ // UpdateWorkOrders: batches quantity, abnormality and process info.
			const (
				newProcessName = "newProcessName"
				newProcessType = "newProcessType"
			)
			err = dm.UpdateWorkOrders(ctx, mcom.UpdateWorkOrdersRequest{
				Orders: []mcom.UpdateWorkOrder{{
					ID: ids.IDs[0],
					BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{
						decimal.NewFromInt(198),
						decimal.NewFromInt(199),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(201),
						decimal.NewFromInt(202),
					}),
					Abnormality: workorder.Abnormality_PLAN_MUTATION,
					ProcessName: newProcessName,
					ProcessType: newProcessType,
				}}})
			assert.NoError(err)

			result, err := dm.ListWorkOrders(ctx, mcom.ListWorkOrdersRequest{
				ID:      ids.IDs[0],
				Date:    newResvDate,
				Station: testWorkOrderStation,
			})

			resvDate, _ := parseDate(newResvDate)
			expected := mcom.ListWorkOrdersReply{WorkOrders: []mcom.GetWorkOrderReply{{
				ID: ids.IDs[0],
				Product: mcom.Product{
					ID:   "product_id",
					Type: "RUBBER",
				},
				Process: mcom.WorkOrderProcess{
					OID:  testProcessOID,
					Name: newProcessName,
					Type: newProcessType,
				},
				RecipeID:      testRecipeID,
				DepartmentOID: testWorkOrderDepartmentOID,
				Station:       testWorkOrderStation,
				Sequence:      1,
				Status:        pbWorkorder.Status_ACTIVE,
				Date:          resvDate,
				Unit:          "kg",
				BatchQuantityDetails: models.BatchQuantityDetails{
					BatchQuantityType: workorder.BatchSize_PER_BATCH_QUANTITIES,
					QuantityForBatches: []decimal.Decimal{
						decimal.NewFromInt(198),
						decimal.NewFromInt(199),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(201),
						decimal.NewFromInt(202),
					},
				},
				CurrentBatch:      0,
				CollectedSequence: 0,
				UpdatedBy:         testWorkOrderUser,
				UpdatedAt:         time.Unix(0, timeNow.UnixNano()),
				InsertedBy:        testWorkOrderUser,
				InsertedAt:        time.Unix(0, timeNow.UnixNano()),
				CollectedQuantity: decimal.Zero,
				Abnormality:       workorder.Abnormality_PLAN_MUTATION,
			}}}
			assert.NoError(err)
			assert.Equal(expected, result)

			// delete mock record data.
			err = db.Where(`work_order = ?`, ids.IDs[0]).Delete(&models.Batch{}).Error
			assert.NoError(err)
			err = db.Where(`work_order = ?`, ids.IDs[0]).Delete(&models.CollectRecord{}).Error
			assert.NoError(err)
		}
		{ // ListWorkOrders: with in/out data.
			// add mock recode.
			recID := strings.ToUpper(xid.New().String())
			tx := dm.(*DataManager).beginTx(ctx)
			defer tx.Rollback() // nolint: errcheck
			assert.NoError(tx.appendFeedRecords(appendFeedRecordRequest{
				WorkOrder: ids.IDs[0],
				Number:    1,
				Records: []models.FeedRecord{
					{
						ID:         recID,
						OperatorID: testUser,
						Materials: []models.FeedDetail{
							{
								Resources: []models.FeedResource{
									{
										ResourceID: testFeedResourceID,
										Quantity:   decimal.NewFromInt(200),
									},
								},
								UniqueSite: models.UniqueSite{Station: testWorkOrderStation},
							},
						},
						Time: timeNow,
					},
				},
			}))
			assert.NoError(tx.Commit())
			err := db.Create(&models.Batch{
				WorkOrder: ids.IDs[0],
				Number:    1,
				Status:    pbWorkorder.BatchStatus_BATCH_STARTED,
				RecordsID: []string{recID},
				UpdatedBy: testUser,
			}).Error
			assert.NoError(err)

			err = db.Create(&models.CollectRecord{
				WorkOrder:   ids.IDs[0],
				Sequence:    1,
				LotNumber:   "001",
				Station:     testWorkOrderStation,
				ResourceOID: testCollectResourceOID,
				Detail: models.CollectRecordDetail{
					OperatorID: testUser,
					Quantity:   decimal.NewFromInt(200),
					BatchCount: 1,
				},
			}).Error
			assert.NoError(err)

			result, err := dm.ListWorkOrders(ctx, mcom.ListWorkOrdersRequest{
				ID:      ids.IDs[0],
				Date:    newResvDate,
				Station: testWorkOrderStation,
			})

			resvDate, _ := parseDate(newResvDate)
			expected := mcom.ListWorkOrdersReply{WorkOrders: []mcom.GetWorkOrderReply{{
				ID: ids.IDs[0],
				Product: mcom.Product{
					ID:   "product_id",
					Type: "RUBBER",
				},
				Process: mcom.WorkOrderProcess{
					OID:  testProcessOID,
					Name: "newProcessName",
					Type: "newProcessType",
				},
				RecipeID:      testRecipeID,
				DepartmentOID: testWorkOrderDepartmentOID,
				Station:       testWorkOrderStation,
				Sequence:      1,
				Status:        pbWorkorder.Status_ACTIVE,
				Date:          resvDate,
				Unit:          "kg",
				BatchQuantityDetails: models.BatchQuantityDetails{
					BatchQuantityType: workorder.BatchSize_PER_BATCH_QUANTITIES,
					QuantityForBatches: []decimal.Decimal{
						decimal.NewFromInt(198),
						decimal.NewFromInt(199),
						decimal.NewFromInt(200),
						decimal.NewFromInt(200),
						decimal.NewFromInt(201),
						decimal.NewFromInt(202),
					},
				},
				CurrentBatch:      1,
				CollectedSequence: 1,
				UpdatedBy:         testWorkOrderUser,
				UpdatedAt:         time.Unix(0, timeNow.UnixNano()),
				InsertedBy:        testWorkOrderUser,
				InsertedAt:        time.Unix(0, timeNow.UnixNano()),
				CollectedQuantity: decimal.NewFromInt(200),
				Abnormality:       workorder.Abnormality_PLAN_MUTATION,
			}}}
			assert.NoError(err)
			assert.Equal(expected, result)

			// delete mock record data.
			err = db.Where(`work_order = ?`, ids.IDs[0]).Delete(&models.Batch{}).Error
			assert.NoError(err)
			err = db.Where(`work_order = ?`, ids.IDs[0]).Delete(&models.CollectRecord{}).Error
			assert.NoError(err)
		}

		// delete mock work order.
		err = deleteMockWorkOrderData(db, ids.IDs)
		assert.NoError(err)
	}
	assert.NoError(cm.Clear())
}

func TestDataManager_ListWorkOrdersByDuration(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &models.WorkOrder{}, &models.RecipeProcessDefinition{}, &models.Recipe{})
	assert.NoError(cm.Clear())

	// #region create test data
	releasedAt := types.ToTimeNano(time.Now())

	assert.NoError(dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
		Recipes: []mcom.Recipe{{
			ID: testRecipeID,
			Product: mcom.Product{
				ID:   "product",
				Type: "type",
			},
			Version: mcom.RecipeVersion{
				Major:      "1",
				Minor:      "0",
				Stage:      "0",
				ReleasedAt: releasedAt,
			},
			Processes: []*mcom.Process{{
				OID:           testProcessOID,
				OptionalFlows: []*mcom.RecipeOptionalFlow{},
			}},
			ProcessDefinitions: []mcom.ProcessDefinition{{
				OID:     testProcessOID,
				Name:    "process_name",
				Type:    "test",
				Configs: []*mcom.RecipeProcessConfig{},
				Output: mcom.OutputProduct{
					ID:   "product",
					Type: "type",
				},
			}},
		}},
	}))

	_, err := dm.CreateWorkOrders(ctx, mcom.CreateWorkOrdersRequest{
		WorkOrders: []mcom.CreateWorkOrder{
			{
				ProcessOID:      testProcessOID,
				RecipeID:        testRecipeID,
				ProcessName:     testProcessName,
				ProcessType:     testProcessType,
				DepartmentOID:   testWorkOrderDepartmentOID,
				Station:         testWorkOrderStation,
				BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200)}),
				Parent:          "parentWorkOrder",
				Date:            time.Date(1999, 9, 9, 9, 9, 9, 9, time.UTC),
				Unit:            "kg",
			},
			{
				ProcessOID:      testProcessOID,
				RecipeID:        testRecipeID,
				ProcessName:     testProcessName,
				ProcessType:     testProcessType,
				DepartmentOID:   testWorkOrderDepartmentOID,
				Station:         testWorkOrderStation,
				BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200)}),
				Parent:          "parentWorkOrder",
				Date:            time.Date(2000, 9, 9, 9, 9, 9, 9, time.UTC),
				Unit:            "kg",
			},
			{
				ProcessOID:      testProcessOID,
				RecipeID:        testRecipeID,
				ProcessName:     testProcessName,
				ProcessType:     testProcessType,
				DepartmentOID:   testWorkOrderDepartmentOID,
				Station:         testWorkOrderStation,
				BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200)}),
				Parent:          "parentWorkOrder",
				Date:            time.Date(2000, 8, 8, 8, 8, 8, 8, time.UTC),
				Unit:            "kg",
			},
		}})
	assert.NoError(err)
	// #endregion create test data

	{ // good case.
		actual, err := dm.ListWorkOrdersByDuration(ctx, mcom.ListWorkOrdersByDurationRequest{
			DepartmentID: testWorkOrderDepartmentOID,
			Station:      testWorkOrderStation,
			Since:        time.Date(2000, 7, 7, 7, 7, 7, 7, time.UTC),
		})
		assert.NoError(err)

		expected := mcom.ListWorkOrdersByDurationReply{
			Contents: []mcom.GetWorkOrderReply{
				{
					ID: actual.Contents[0].ID,
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					Station:       testWorkOrderStation,
					Sequence:      3,
					Date:          time.Date(2000, 8, 8, 0, 0, 0, 0, time.UTC),
					Unit:          "kg",
					Parent:        "parentWorkOrder",
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					UpdatedAt:         actual.Contents[0].InsertedAt,
					InsertedAt:        actual.Contents[0].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
				{
					ID: actual.Contents[1].ID,
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					Station:       testWorkOrderStation,
					Sequence:      2,
					Unit:          "kg",
					Date:          time.Date(2000, 9, 9, 0, 0, 0, 0, time.UTC),
					Parent:        "parentWorkOrder",
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					UpdatedAt:         actual.Contents[1].InsertedAt,
					InsertedAt:        actual.Contents[1].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
			},
		}
		assert.Equal(expected, actual)
	}
	{ // good case with Until.
		actual, err := dm.ListWorkOrdersByDuration(ctx, mcom.ListWorkOrdersByDurationRequest{
			DepartmentID: testWorkOrderDepartmentOID,
			Station:      testWorkOrderStation,
			Since:        time.Date(2000, 7, 7, 7, 7, 7, 7, time.UTC),
			Until:        time.Date(2000, 8, 30, 0, 0, 0, 0, time.UTC),
		})
		assert.NoError(err)

		expected := mcom.ListWorkOrdersByDurationReply{
			Contents: []mcom.GetWorkOrderReply{
				{
					ID: actual.Contents[0].ID,
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					Station:       testWorkOrderStation,
					Sequence:      3,
					Date:          time.Date(2000, 8, 8, 0, 0, 0, 0, time.UTC),
					Unit:          "kg",
					Parent:        "parentWorkOrder",
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					UpdatedAt:         actual.Contents[0].InsertedAt,
					InsertedAt:        actual.Contents[0].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
			},
		}
		assert.Equal(expected, actual)
	}
	{ // good case with Department.
		actual, err := dm.ListWorkOrdersByDuration(ctx, mcom.ListWorkOrdersByDurationRequest{
			DepartmentID: "T1234",
			Since:        time.Date(1999, 9, 9, 0, 0, 0, 0, time.UTC),
			Until:        time.Date(2000, 9, 9, 0, 0, 0, 0, time.UTC),
		})
		assert.NoError(err)

		expected := mcom.ListWorkOrdersByDurationReply{
			Contents: []mcom.GetWorkOrderReply{
				{
					ID: actual.Contents[0].ID,
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					Station:       testWorkOrderStation,
					Sequence:      1,
					Date:          time.Date(1999, 9, 9, 0, 0, 0, 0, time.UTC),
					Unit:          "kg",
					Parent:        "parentWorkOrder",
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					UpdatedAt:         actual.Contents[0].InsertedAt,
					InsertedAt:        actual.Contents[0].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
				{
					ID: actual.Contents[1].ID,
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					Station:       testWorkOrderStation,
					Sequence:      3,
					Date:          time.Date(2000, 8, 8, 0, 0, 0, 0, time.UTC),
					Unit:          "kg",
					Parent:        "parentWorkOrder",
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					UpdatedAt:         actual.Contents[1].InsertedAt,
					InsertedAt:        actual.Contents[1].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
				{
					ID: actual.Contents[2].ID,
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					Station:       testWorkOrderStation,
					Sequence:      2,
					Date:          time.Date(2000, 9, 9, 0, 0, 0, 0, time.UTC),
					Unit:          "kg",
					Parent:        "parentWorkOrder",
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200), decimal.NewFromInt(200)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					UpdatedAt:         actual.Contents[2].InsertedAt,
					InsertedAt:        actual.Contents[2].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
			},
		}
		assert.Equal(expected, actual)
	}
	assert.NoError(cm.Clear())
}

func TestDataManager_ListWorkOrdersByIDs(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &models.WorkOrder{}, &models.RecipeProcessDefinition{}, &models.Recipe{})
	assert.NoError(cm.Clear())

	releasedAt := types.ToTimeNano(time.Now())
	// #region create test data
	assert.NoError(dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
		Recipes: []mcom.Recipe{{
			ID: testRecipeID,
			Product: mcom.Product{
				ID:   "product",
				Type: "type",
			},
			Version: mcom.RecipeVersion{
				Major:      "1",
				Minor:      "0",
				Stage:      "0",
				ReleasedAt: releasedAt,
			},
			Processes: []*mcom.Process{{
				OID:           testProcessOID,
				OptionalFlows: []*mcom.RecipeOptionalFlow{},
			}},
			ProcessDefinitions: []mcom.ProcessDefinition{{
				OID:     testProcessOID,
				Name:    "process_name",
				Type:    "test",
				Configs: []*mcom.RecipeProcessConfig{},
				Output: mcom.OutputProduct{
					ID:   "product",
					Type: "type",
				},
			}},
		}},
	}))

	woIDs, err := dm.CreateWorkOrders(ctx, mcom.CreateWorkOrdersRequest{
		WorkOrders: []mcom.CreateWorkOrder{{
			ProcessOID:      testProcessOID,
			RecipeID:        testRecipeID,
			ProcessName:     "process_name",
			ProcessType:     "test",
			Status:          pbWorkorder.Status_PENDING,
			BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{decimal.NewFromInt(10)}),
			DepartmentOID:   testWorkOrderDepartmentOID,
			Station:         testWorkOrderStation,
			Sequence:        1,
			Date:            time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
		},
			{
				ProcessOID:      testProcessOID,
				RecipeID:        testRecipeID,
				ProcessName:     "process_name",
				ProcessType:     "test",
				Status:          pbWorkorder.Status_PENDING,
				BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{decimal.NewFromInt(10)}),
				DepartmentOID:   testWorkOrderDepartmentOID,
				Station:         testWorkOrderStation,
				Sequence:        2,
				Date:            time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
			},
			{
				ProcessOID:      testProcessOID,
				RecipeID:        testRecipeID,
				ProcessName:     "process_name",
				ProcessType:     "test",
				Status:          pbWorkorder.Status_PENDING,
				BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{decimal.NewFromInt(10)}),
				DepartmentOID:   testWorkOrderDepartmentOID,
				Station:         testWorkOrderStation,
				Sequence:        3,
				Date:            time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
			},
			{
				ProcessOID:      testProcessOID,
				RecipeID:        testRecipeID,
				ProcessName:     "process_name",
				ProcessType:     "test",
				Status:          pbWorkorder.Status_PENDING,
				BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{decimal.NewFromInt(10)}),
				DepartmentOID:   testWorkOrderDepartmentOID,
				Station:         testWorkOrderStation,
				Sequence:        4,
				Date:            time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
			},
			{
				ProcessOID:      testProcessOID,
				RecipeID:        testRecipeID,
				ProcessName:     "process_name",
				ProcessType:     "test",
				Status:          pbWorkorder.Status_PENDING,
				BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{decimal.NewFromInt(10)}),
				DepartmentOID:   testWorkOrderDepartmentOID,
				Station:         "MA",
				Sequence:        5,
				Date:            time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
			}},
	})
	assert.NoError(err)
	// #endregion create test data

	{ // good case : list all.
		actual, err := dm.ListWorkOrdersByIDs(ctx, mcom.ListWorkOrdersByIDsRequest{
			IDs: woIDs.IDs,
		})
		assert.NoError(err)
		expected := mcom.ListWorkOrdersByIDsReply{
			Contents: map[string]mcom.GetWorkOrderReply{
				woIDs.IDs[0]: {
					ID: woIDs.IDs[0],
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(10)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					Station:           testWorkOrderStation,
					Sequence:          1,
					Date:              time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
					UpdatedAt:         actual.Contents[woIDs.IDs[0]].UpdatedAt,
					InsertedAt:        actual.Contents[woIDs.IDs[0]].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
				woIDs.IDs[1]: {
					ID: woIDs.IDs[1],
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(10)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					Station:           testWorkOrderStation,
					Sequence:          2,
					Date:              time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
					UpdatedAt:         actual.Contents[woIDs.IDs[1]].UpdatedAt,
					InsertedAt:        actual.Contents[woIDs.IDs[1]].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
				woIDs.IDs[2]: {
					ID: woIDs.IDs[2],
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					BatchQuantityDetails: models.BatchQuantityDetails{
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(10)},
					},
					Station:           testWorkOrderStation,
					Sequence:          3,
					Date:              time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
					UpdatedAt:         actual.Contents[woIDs.IDs[2]].UpdatedAt,
					InsertedAt:        actual.Contents[woIDs.IDs[2]].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
				woIDs.IDs[3]: {
					ID: woIDs.IDs[3],
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					BatchQuantityDetails: models.BatchQuantityDetails{
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(10)},
					},
					Station:           testWorkOrderStation,
					Sequence:          4,
					Date:              time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
					UpdatedAt:         actual.Contents[woIDs.IDs[3]].UpdatedAt,
					InsertedAt:        actual.Contents[woIDs.IDs[3]].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
				woIDs.IDs[4]: {
					ID: woIDs.IDs[4],
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(10)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					Station:           "MA",
					Sequence:          5,
					Date:              time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
					UpdatedAt:         actual.Contents[woIDs.IDs[4]].UpdatedAt,
					InsertedAt:        actual.Contents[woIDs.IDs[4]].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
			},
		}
		assert.Equal(expected, actual)
	}
	{ // good case : list all with station.
		actual, err := dm.ListWorkOrdersByIDs(ctx, mcom.ListWorkOrdersByIDsRequest{
			IDs:     woIDs.IDs,
			Station: testWorkOrderStation,
		})
		assert.NoError(err)
		expected := mcom.ListWorkOrdersByIDsReply{
			Contents: map[string]mcom.GetWorkOrderReply{
				woIDs.IDs[0]: {
					ID: woIDs.IDs[0],
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(10)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					Station:           testWorkOrderStation,
					Sequence:          1,
					Date:              time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
					UpdatedAt:         actual.Contents[woIDs.IDs[0]].UpdatedAt,
					InsertedAt:        actual.Contents[woIDs.IDs[0]].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
				woIDs.IDs[1]: {
					ID: woIDs.IDs[1],
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(10)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					Station:           testWorkOrderStation,
					Sequence:          2,
					Date:              time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
					UpdatedAt:         actual.Contents[woIDs.IDs[1]].UpdatedAt,
					InsertedAt:        actual.Contents[woIDs.IDs[1]].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
				woIDs.IDs[2]: {
					ID: woIDs.IDs[2],
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(10)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					Station:           testWorkOrderStation,
					Sequence:          3,
					Date:              time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
					UpdatedAt:         actual.Contents[woIDs.IDs[2]].UpdatedAt,
					InsertedAt:        actual.Contents[woIDs.IDs[2]].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
				woIDs.IDs[3]: {
					ID: woIDs.IDs[3],
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(10)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					Station:           testWorkOrderStation,
					Sequence:          4,
					Date:              time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
					UpdatedAt:         actual.Contents[woIDs.IDs[3]].UpdatedAt,
					InsertedAt:        actual.Contents[woIDs.IDs[3]].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
			},
		}
		assert.Equal(expected, actual)
	}
	{ // good case : list part of ids.
		actual, err := dm.ListWorkOrdersByIDs(ctx, mcom.ListWorkOrdersByIDsRequest{
			IDs:     []string{woIDs.IDs[0], woIDs.IDs[2]},
			Station: testWorkOrderStation,
		})
		assert.NoError(err)
		expected := mcom.ListWorkOrdersByIDsReply{
			Contents: map[string]mcom.GetWorkOrderReply{
				woIDs.IDs[0]: {
					ID: woIDs.IDs[0],
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(10)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					Station:           testWorkOrderStation,
					Sequence:          1,
					Date:              time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
					UpdatedAt:         actual.Contents[woIDs.IDs[0]].UpdatedAt,
					InsertedAt:        actual.Contents[woIDs.IDs[0]].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
				woIDs.IDs[2]: {
					ID: woIDs.IDs[2],
					Product: mcom.Product{
						ID:   "product",
						Type: "type",
					},
					Process: mcom.WorkOrderProcess{
						OID:  testProcessOID,
						Name: "process_name",
						Type: "test",
					},
					RecipeID:      testRecipeID,
					Status:        pbWorkorder.Status_PENDING,
					DepartmentOID: testWorkOrderDepartmentOID,
					BatchQuantityDetails: models.BatchQuantityDetails{
						QuantityForBatches: []decimal.Decimal{decimal.NewFromInt(10)},
						BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
					},
					Station:           testWorkOrderStation,
					Sequence:          3,
					Date:              time.Date(2022, 5, 11, 0, 0, 0, 0, time.UTC),
					UpdatedAt:         actual.Contents[woIDs.IDs[2]].UpdatedAt,
					InsertedAt:        actual.Contents[woIDs.IDs[2]].InsertedAt,
					CollectedQuantity: decimal.Zero,
				},
			},
		}
		assert.Equal(expected, actual)
	}
	assert.NoError(cm.Clear())
}
