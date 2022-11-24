package impl

import (
	"context"
	"fmt"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

const (
	testUpdatedBy = "tester"

	testResourceID        = "R0123456789"
	testResourceType      = "RUBB"
	testWarehouseID       = "A"
	testWarehouseLocation = "00"

	testWarehouseIDNew       = "X"
	testWarehouseLocationNew = "11"

	testResourceNotFound = "YOW"
)

// use to create resources with the quantity <= 0.
func createMaterialResources(db *gorm.DB, resource mcom.Material, stock mcom.Warehouse) error {
	recordsID := []string{}
	if resource.FeedRecordsID != nil {
		recordsID = resource.FeedRecordsID
	}
	materialResource := models.MaterialResource{
		ID:          resource.ResourceID,
		ProductID:   resource.ID,
		ProductType: resource.Type,
		Quantity:    resource.Quantity,
		Status:      resource.Status,
		ExpiryTime:  types.ToTimeNano(resource.ExpiryTime),
		Info: models.MaterialInfo{
			Grade:          resource.Grade,
			Unit:           resource.Unit,
			LotNumber:      resource.LotNumber,
			ProductionTime: resource.ProductionTime,
		},
		UpdatedBy:         "Rockefeller",
		WarehouseID:       stock.ID,
		WarehouseLocation: stock.Location,
		FeedRecordsID:     recordsID,
	}

	return db.Create(&materialResource).Error
}

func createMockWarehouseStockData(dm mcom.DataManager, resourceID string) (*models.WarehouseStock, error) {
	_, err := dm.CreateMaterialResources(context.Background(), mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{{
			Type:       testResourceType,
			ID:         testResourceMaterialID,
			Quantity:   decimal.RequireFromString(testResourceMaterialQuantity),
			ResourceID: resourceID,
		}},
	}, mcom.WithStockIn(mcom.Warehouse{
		ID:       testWarehouseID,
		Location: testWarehouseLocation,
	}))
	if err != nil {
		return nil, err
	}

	return &models.WarehouseStock{
		ID:        testWarehouseID,
		Location:  testWarehouseLocation,
		ProductID: testResourceMaterialID,
		Quantity:  decimal.RequireFromString(testResourceMaterialQuantity),
	}, nil
}

func deleteAllWarehouseStocks(db *gorm.DB) error {
	return newClearMaster(db, &models.WarehouseStock{}).Clear()
}

func TestDataManager_GetResourceWarehouse(t *testing.T) {
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

	assert.NoError(deleteAllWarehouseStocks(db))
	assert.NoError(deleteAllMaterialResources(db))

	sDB, err := db.DB()
	if !assert.NoError(err) {
		return
	}
	defer sDB.Close()

	// 統一時間
	timeNow := time.Now().Local()
	assert.NoError(err)
	monkey.Patch(time.Now, func() time.Time {
		return timeNow
	})
	defer monkey.UnpatchAll()

	{ // empty resource id.
		_, err := dm.GetMaterialResource(ctx, mcom.GetMaterialResourceRequest{})

		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_RESOURCE_NOT_FOUND,
		}, err)
	}
	{ // material not found.
		_, err := dm.GetMaterialResource(ctx, mcom.GetMaterialResourceRequest{ResourceID: testResourceID})

		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_RESOURCE_NOT_FOUND,
		}, err)
	}
	{ // good case.
		err = db.Create(&models.MaterialResource{
			ID:          testResourceID,
			ProductID:   testMaterialName,
			ProductType: testResourceType,
			Status:      0,
			Quantity:    decimal.NewFromFloat(123.456),
			ExpiryTime:  types.ToTimeNano(time.Now()),
			Info: models.MaterialInfo{
				Grade:          testMaterialGrade,
				Unit:           "kg",
				LotNumber:      "A123456",
				ProductionTime: time.Now(),
				MinDosage:      decimal.NewFromInt(10),
			},
			UpdatedBy:     testUpdatedBy,
			FeedRecordsID: []string{},
		}).Error
		assert.NoError(err)

		resp, err := dm.GetMaterialResource(ctx, mcom.GetMaterialResourceRequest{ResourceID: testResourceID})
		if assert.NoError(err) {
			assert.Equal(mcom.GetMaterialResourceReply{
				{
					Material: mcom.Material{
						Type:            testResourceType,
						ID:              testMaterialName,
						Grade:           testMaterialGrade,
						Status:          1,
						Quantity:        decimal.RequireFromString("123.456000"),
						PlannedQuantity: decimal.NewFromInt(0),
						Unit:            "kg",
						LotNumber:       "A123456",
						MinDosage:       decimal.NewFromInt(10),
						ProductionTime:  time.Now(),
						ExpiryTime:      time.Now(),
						ResourceID:      testResourceID,
						CreatedAt:       resp[0].Material.CreatedAt,
						UpdatedAt:       resp[0].Material.UpdatedAt,
						UpdatedBy:       "tester",
						FeedRecordsID:   []string{},
					},
					Warehouse: mcom.Warehouse{
						ID:       "",
						Location: "",
					},
				},
			}, resp)
		}
	}

	assert.NoError(deleteAllWarehouseStocks(db))
	assert.NoError(deleteAllMaterialResources(db))
}

func TestDataManager_GetMaterialResourceIdentity(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	assert.NoError(deleteAllWarehouseStocks(db))
	assert.NoError(deleteAllMaterialResources(db))

	{ // normal case (should pass)
		_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{{
				Type:       "Type1",
				ID:         "ID1",
				Grade:      "A",
				Status:     resources.MaterialStatus_AVAILABLE,
				Quantity:   decimal.NewFromInt(2),
				Unit:       "g",
				LotNumber:  "10",
				ResourceID: "RID1",
			}, {
				Type:       "Type2",
				ID:         "ID2",
				Grade:      "A",
				Status:     resources.MaterialStatus_AVAILABLE,
				Quantity:   decimal.NewFromInt(5),
				MinDosage:  decimal.NewFromInt(2),
				Station:    "station",
				Unit:       "g",
				LotNumber:  "10",
				ResourceID: "RID2",
			}},
		}, mcom.WithStockIn(mcom.Warehouse{
			ID:       testWarehouseID,
			Location: testWarehouseLocation,
		}))
		assert.NoError(err)
		actual, err := dm.GetMaterialResourceIdentity(ctx, mcom.GetMaterialResourceIdentityRequest{
			ResourceID:  "RID2",
			ProductType: "Type2",
		})
		assert.NoError(err)
		assert.Equal(mcom.GetMaterialResourceIdentityReply{
			Material: mcom.Material{
				Type:            "Type2",
				ID:              "ID2",
				Grade:           "A",
				Status:          resources.MaterialStatus_AVAILABLE,
				Quantity:        decimal.RequireFromString("5.000000"),
				PlannedQuantity: decimal.NewFromInt(0),
				ProductionTime:  actual.Material.ProductionTime,
				ExpiryTime:      actual.Material.ExpiryTime,
				ResourceID:      "RID2",
				Unit:            "g",
				LotNumber:       "10",
				MinDosage:       decimal.NewFromInt(2),
				Station:         "station",
				UpdatedAt:       actual.Material.UpdatedAt,
				CreatedAt:       actual.Material.CreatedAt,
				FeedRecordsID:   []string{},
			},
			Warehouse: mcom.Warehouse{
				ID:       "A",
				Location: "00",
			},
		}, actual)
	}
	{ // type not found (should not pass)
		_, err := dm.GetMaterialResourceIdentity(ctx, mcom.GetMaterialResourceIdentityRequest{
			ResourceID:  "RID1",
			ProductType: "Type6",
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_RESOURCE_NOT_FOUND})
	}

	assert.NoError(deleteAllWarehouseStocks(db))
	assert.NoError(deleteAllMaterialResources(db))
}

func TestDataManager_ListMaterialResourceIdentities(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	assert.NoError(deleteAllWarehouseStocks(db))
	assert.NoError(deleteAllMaterialResources(db))
	{ // not found (should pass)
		rep, err := dm.ListMaterialResourceIdentities(ctx, mcom.ListMaterialResourceIdentitiesRequest{Details: []mcom.GetMaterialResourceIdentityRequest{{
			ResourceID:  "not found",
			ProductType: "not found",
		}}})
		assert.NoError(err)
		assert.Equal(rep, mcom.ListMaterialResourceIdentitiesReply{Replies: []*mcom.MaterialReply{nil}})
	}
	{ // normal case (should pass)
		_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{{
				Type:       "Type1",
				ID:         "ID1",
				Grade:      "A",
				Status:     resources.MaterialStatus_AVAILABLE,
				Quantity:   decimal.NewFromInt(2),
				Unit:       "g",
				LotNumber:  "10",
				ResourceID: "RID1",
			}, {
				Type:       "Type2",
				ID:         "ID2",
				Grade:      "A",
				Status:     resources.MaterialStatus_AVAILABLE,
				Quantity:   decimal.NewFromInt(5),
				Unit:       "g",
				LotNumber:  "10",
				ResourceID: "RID2",
			}},
		}, mcom.WithStockIn(mcom.Warehouse{
			ID:       testWarehouseID,
			Location: testWarehouseLocation,
		}))
		assert.NoError(err)

		actual, err := dm.ListMaterialResourceIdentities(ctx, mcom.ListMaterialResourceIdentitiesRequest{
			Details: []mcom.GetMaterialResourceIdentityRequest{
				{
					ResourceID:  "RID1",
					ProductType: "Type1",
				},
				{
					ResourceID:  "RID2",
					ProductType: "Type2",
				},
				{
					ResourceID:  "not found",
					ProductType: "not found",
				},
			}})
		assert.NoError(err)

		expected := mcom.ListMaterialResourceIdentitiesReply{
			Replies: []*mcom.MaterialReply{{
				Material: mcom.Material{
					Type:            "Type1",
					ID:              "ID1",
					Grade:           "A",
					Status:          resources.MaterialStatus_AVAILABLE,
					Quantity:        decimal.RequireFromString("2.000000"),
					PlannedQuantity: decimal.NewFromInt(0),
					Unit:            "g",
					LotNumber:       "10",
					MinDosage:       decimal.NewFromInt(0),
					ProductionTime:  actual.Replies[0].Material.ProductionTime,
					ExpiryTime:      actual.Replies[0].Material.ExpiryTime,
					ResourceID:      "RID1",
					CreatedAt:       actual.Replies[0].Material.CreatedAt,
					UpdatedAt:       actual.Replies[0].Material.UpdatedAt,
					FeedRecordsID:   []string{},
				},
				Warehouse: mcom.Warehouse{
					ID:       testWarehouseID,
					Location: testWarehouseLocation,
				},
			},
				{
					Material: mcom.Material{
						Type:            "Type2",
						ID:              "ID2",
						Grade:           "A",
						Status:          resources.MaterialStatus_AVAILABLE,
						Quantity:        decimal.RequireFromString("5.000000"),
						PlannedQuantity: decimal.NewFromInt(0),
						Unit:            "g",
						LotNumber:       "10",
						MinDosage:       decimal.NewFromInt(0),
						ProductionTime:  actual.Replies[1].Material.ProductionTime,
						ExpiryTime:      actual.Replies[1].Material.ExpiryTime,
						ResourceID:      "RID2",
						CreatedAt:       actual.Replies[1].Material.CreatedAt,
						UpdatedAt:       actual.Replies[1].Material.UpdatedAt,
						FeedRecordsID:   []string{},
					},
					Warehouse: mcom.Warehouse{
						ID:       testWarehouseID,
						Location: testWarehouseLocation,
					},
				}, nil}}
		assert.Equal(expected, actual)
	}
	{ // empty resource id and product type.
		rep, err := dm.ListMaterialResourceIdentities(ctx, mcom.ListMaterialResourceIdentitiesRequest{
			Details: []mcom.GetMaterialResourceIdentityRequest{
				{
					ResourceID:  "",
					ProductType: "",
				},
			},
		})
		assert.NoError(err)
		assert.Equal(mcom.ListMaterialResourceIdentitiesReply{Replies: []*mcom.MaterialReply{nil}}, rep)
	}
	assert.NoError(deleteAllWarehouseStocks(db))
	assert.NoError(deleteAllMaterialResources(db))
}

func Test_ListMaterialResources(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	assert.NoError(deleteAllWarehouseStocks(db))
	assert.NoError(deleteAllMaterialResources(db))
	patchDate1_0 := time.Date(2022, 2, 22, 2, 22, 22, 22, time.Local)
	patchDate1_1 := patchDate1_0.Add(time.Second)
	patchDate1_2 := patchDate1_1.Add(time.Second)
	patchDate1_3 := patchDate1_2.Add(time.Second)

	patchDate2_0 := time.Date(2033, 3, 33, 3, 33, 33, 33, time.Local)
	patchDate2_1 := patchDate2_0.Add(time.Second)
	patchDate2_2 := patchDate2_1.Add(time.Second)
	patchDate2_3 := patchDate2_2.Add(time.Second)

	patch := monkey.Patch(time.Now, func() time.Time { return patchDate1_0 })
	afterPatchTime := types.ToTimeNano(patchDate1_3) + 1

	generateCreateResourcesRequest := func(resourceID, productID, productType string, status resources.MaterialStatus) mcom.CreateMaterialResourcesRequestDetail {
		return mcom.CreateMaterialResourcesRequestDetail{
			Type:            productType,
			ID:              productID,
			Grade:           "A",
			Status:          status,
			Quantity:        decimal.RequireFromString("10.000000"),
			PlannedQuantity: decimal.NewFromInt(0),
			Unit:            "kg",
			LotNumber:       "1",
			MinDosage:       decimal.NewFromInt(0),
			ProductionTime:  time.Date(2022, 3, 7, 0, 0, 0, 0, time.Local),
			ExpiryTime:      time.Date(2022, 3, 7, 0, 0, 0, 0, time.Local),
			ResourceID:      resourceID,
		}
	}

	generateMaterialReply := func(material mcom.CreateMaterialResourcesRequestDetail, createdAt types.TimeNano) mcom.MaterialReply {
		// for gitlab ci.
		material.ExpiryTime = material.ExpiryTime.Local()
		material.ProductionTime = material.ProductionTime.Local()

		return mcom.MaterialReply{Material: mcom.Material{
			Type:            material.Type,
			ID:              material.ID,
			Grade:           material.Grade,
			Status:          material.Status,
			Quantity:        material.Quantity,
			PlannedQuantity: material.PlannedQuantity,
			Unit:            material.Unit,
			LotNumber:       material.LotNumber,
			ProductionTime:  material.ProductionTime,
			ExpiryTime:      material.ExpiryTime,
			ResourceID:      material.ResourceID,
			MinDosage:       material.MinDosage,
			Inspections:     material.Inspections,
			Remark:          material.Remark,
			CarrierID:       material.CarrierID,
			UpdatedAt:       createdAt,
			CreatedAt:       createdAt,
			FeedRecordsID:   []string{},
		}, Warehouse: mcom.Warehouse{}}
	}

	setLocalTimeToReply := func(reps mcom.ListMaterialResourcesReply) mcom.ListMaterialResourcesReply {
		for i := range reps.Resources {
			reps.Resources[i].Material.ExpiryTime = reps.Resources[i].Material.ExpiryTime.Local()
			reps.Resources[i].Material.ProductionTime = reps.Resources[i].Material.ProductionTime.Local()
		}
		return reps
	}

	req1 := generateCreateResourcesRequest("RID0", "PID0", "PType0", resources.MaterialStatus_AVAILABLE)
	req2 := generateCreateResourcesRequest("RID0", "PID0", "PType1", resources.MaterialStatus_AVAILABLE)
	req3 := generateCreateResourcesRequest("RID1", "PID1", "PType0", resources.MaterialStatus_AVAILABLE)
	req4 := generateCreateResourcesRequest("RID2", "PID0", "PType0", resources.MaterialStatus_HOLD)

	rep1 := generateMaterialReply(req1, types.ToTimeNano(patchDate1_0))
	_ = generateMaterialReply(req2, types.ToTimeNano(patchDate1_1))
	rep3 := generateMaterialReply(req3, types.ToTimeNano(patchDate1_2))
	rep4 := generateMaterialReply(req4, types.ToTimeNano(patchDate1_3))

	_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{req1},
	})
	assert.NoError(err)

	patch.Unpatch()
	patch = monkey.Patch(time.Now, func() time.Time { return patchDate1_1 })

	_, err = dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{req2},
	})
	assert.NoError(err)

	patch.Unpatch()
	patch = monkey.Patch(time.Now, func() time.Time { return patchDate1_2 })

	_, err = dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{req3},
	})
	assert.NoError(err)

	patch.Unpatch()
	patch = monkey.Patch(time.Now, func() time.Time { return patchDate1_3 })

	_, err = dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{req4},
	})
	assert.NoError(err)

	patch.Unpatch()
	patch = monkey.Patch(time.Now, func() time.Time { return patchDate2_0 })

	req5 := generateCreateResourcesRequest("RID3", "PID0", "PType0", resources.MaterialStatus_AVAILABLE)
	req6 := generateCreateResourcesRequest("RID3", "PID0", "PType1", resources.MaterialStatus_AVAILABLE)
	req7 := generateCreateResourcesRequest("RID4", "PID1", "PType0", resources.MaterialStatus_AVAILABLE)
	req8 := generateCreateResourcesRequest("RID5", "PID0", "PType0", resources.MaterialStatus_HOLD)

	rep5 := generateMaterialReply(req5, types.ToTimeNano(patchDate2_0))
	_ = generateMaterialReply(req6, types.ToTimeNano(patchDate2_1))
	rep7 := generateMaterialReply(req7, types.ToTimeNano(patchDate2_2))
	rep8 := generateMaterialReply(req8, types.ToTimeNano(patchDate2_3))

	_, err = dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{req5},
	})
	assert.NoError(err)

	patch.Unpatch()
	patch = monkey.Patch(time.Now, func() time.Time { return patchDate2_1 })

	_, err = dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{req6},
	})
	assert.NoError(err)

	patch.Unpatch()
	patch = monkey.Patch(time.Now, func() time.Time { return patchDate2_2 })

	_, err = dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{req7},
	})
	assert.NoError(err)

	patch.Unpatch()
	patch = monkey.Patch(time.Now, func() time.Time { return patchDate2_3 })

	_, err = dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{req8},
	})
	assert.NoError(err)

	patch.Unpatch()

	{ // empty request.
		rep, err := dm.ListMaterialResources(ctx, mcom.ListMaterialResourcesRequest{})
		assert.NoError(err)
		assert.Equal(mcom.ListMaterialResourcesReply{Resources: []mcom.MaterialReply{}}, rep)
	}
	{ // search by product type.
		actual, err := dm.ListMaterialResources(ctx, mcom.ListMaterialResourcesRequest{ProductType: "PType0"}.WithOrder(mcom.Order{Name: "created_at"}))
		assert.NoError(err)
		assert.Equal(mcom.ListMaterialResourcesReply{
			Resources: []mcom.MaterialReply{
				rep1,
				rep3,
				rep4,
				rep5,
				rep7,
				rep8,
			},
		}, setLocalTimeToReply(actual))
	}
	{ // search by product type and status.
		actual, err := dm.ListMaterialResources(ctx, mcom.ListMaterialResourcesRequest{ProductType: "PType0", Status: resources.MaterialStatus_HOLD}.WithOrder(mcom.Order{Name: "created_at"}))
		assert.NoError(err)
		assert.Equal(mcom.ListMaterialResourcesReply{
			Resources: []mcom.MaterialReply{
				rep4, rep8,
			}}, setLocalTimeToReply(actual))
	}
	{ // search by product type and id.
		actual, err := dm.ListMaterialResources(ctx, mcom.ListMaterialResourcesRequest{ProductType: "PType0", ProductID: "PID0"}.WithOrder(mcom.Order{Name: "created_at"}))
		assert.NoError(err)
		assert.Equal(mcom.ListMaterialResourcesReply{
			Resources: []mcom.MaterialReply{
				rep1, rep4, rep5, rep8,
			}}, setLocalTimeToReply(actual))
	}
	{ // search by product type and created at.
		actual, err := dm.ListMaterialResources(ctx, mcom.ListMaterialResourcesRequest{ProductType: "PType0", CreatedAt: afterPatchTime}.WithOrder(mcom.Order{Name: "created_at"}))
		assert.NoError(err)
		assert.Equal(mcom.ListMaterialResourcesReply{
			Resources: []mcom.MaterialReply{
				rep5, rep7, rep8,
			}}, setLocalTimeToReply(actual))
	}
	{ // search by product type with pagination.
		actual, err := dm.ListMaterialResources(ctx, mcom.ListMaterialResourcesRequest{ProductType: "PType0"}.
			WithPagination(mcom.PaginationRequest{
				PageCount:      2,
				ObjectsPerPage: 2,
			}).WithOrder(mcom.Order{Name: "created_at"}))
		assert.NoError(err)
		assert.Equal(mcom.ListMaterialResourcesReply{
			Resources: []mcom.MaterialReply{
				rep4,
				rep5,
			},
			PaginationReply: mcom.PaginationReply{
				AmountOfData: 6,
			},
		}, setLocalTimeToReply(actual))
	}
	{ // search by product type with pagination out of range.
		actual, err := dm.ListMaterialResources(ctx, mcom.ListMaterialResourcesRequest{ProductType: "PType0"}.
			WithPagination(mcom.PaginationRequest{
				PageCount:      2,
				ObjectsPerPage: 20,
			}).WithOrder(mcom.Order{Name: "created_at"}))
		assert.NoError(err)
		assert.Equal(mcom.ListMaterialResourcesReply{
			Resources: []mcom.MaterialReply{},
			PaginationReply: mcom.PaginationReply{
				AmountOfData: 6,
			},
		}, setLocalTimeToReply(actual))
	}
	{ // search by product type.
		actual, err := dm.ListMaterialResources(ctx, mcom.ListMaterialResourcesRequest{ProductType: "PType0"}.WithOrder(mcom.Order{
			Name:       "created_at",
			Descending: true,
		}))
		assert.NoError(err)
		assert.Equal(mcom.ListMaterialResourcesReply{
			Resources: []mcom.MaterialReply{
				rep8,
				rep7,
				rep5,
				rep4,
				rep3,
				rep1,
			},
		}, setLocalTimeToReply(actual))
	}
	assert.NoError(deleteAllWarehouseStocks(db))
	assert.NoError(deleteAllMaterialResources(db))
}

func Test_GetResourceWarehouse(t *testing.T) {
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

	assert.NoError(deleteAllWarehouseStocks(db))

	sDB, err := db.DB()
	if !assert.NoError(err) {
		return
	}
	defer sDB.Close()
	{ // insufficient request.
		_, err = dm.GetResourceWarehouse(ctx, mcom.GetResourceWarehouseRequest{ResourceID: ""})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "empty resourceID",
		}, err)
	}
	{ // warehouse resource not found.
		_, err = dm.GetResourceWarehouse(ctx, mcom.GetResourceWarehouseRequest{ResourceID: resourceID})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_RESOURCE_NOT_FOUND,
			Details: fmt.Sprintf("warehouse resource not found: %v", resourceID),
		}, err)
	}
	{ // good case.
		ws, err := createMockWarehouseStockData(dm, resourceID)
		assert.NoError(err)

		result, err := dm.GetResourceWarehouse(ctx, mcom.GetResourceWarehouseRequest{ResourceID: resourceID})
		assert.NoError(err)
		assert.Equal(mcom.GetResourceWarehouseReply{
			ID:       ws.ID,
			Location: ws.Location,
		}, result)
	}

	assert.NoError(deleteAllWarehouseStocks(db))
}

func createMockWarehouseData(db *gorm.DB) error {
	return db.Create(&models.Warehouse{
		ID:           testWarehouseIDNew,
		DepartmentID: "A0000",
	}).Error
}

func deleteAllWarehouseData(db *gorm.DB) error {
	return newClearMaster(db, &models.Warehouse{}).Clear()
}

func queryMockCreatedTransportRecord(db *gorm.DB) (models.ResourceTransportRecord, error) {
	var record models.ResourceTransportRecord
	err := db.Where(` resource_id = ? `, testResourceID).Find(&record).Error
	return record, err
}

func deleteAllCreatedTransportRecord(db *gorm.DB) error {
	return newClearMaster(db, &models.ResourceTransportRecord{}).Clear()
}

func queryMockCreatedWarehouseStockData(db *gorm.DB) (models.WarehouseStock, error) {
	var resource models.MaterialResource
	err := db.Where(`id = ? `, testResourceID).Find(&resource).Error
	return models.WarehouseStock{
		ID:        resource.WarehouseID,
		Location:  resource.WarehouseLocation,
		ProductID: resource.ProductID,
		Quantity:  resource.Quantity,
	}, err
}

func TestDataManager_WarehousingStock(t *testing.T) {
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

	assert.NoError(deleteAllWarehouseData(db))
	assert.NoError(deleteAllMaterialResources(db))
	assert.NoError(deleteAllWarehouseStocks(db))
	assert.NoError(deleteAllCreatedTransportRecord(db))

	sDB, err := db.DB()
	if !assert.NoError(err) {
		return
	}
	defer sDB.Close()

	assert.NoError(createMockWarehouseData(db))

	{ // insufficient request.
		assert.ErrorIs(dm.WarehousingStock(ctx, mcom.WarehousingStockRequest{}),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing required request",
			})
	}
	{ // insufficient request.

		assert.ErrorIs(dm.WarehousingStock(ctx, mcom.WarehousingStockRequest{
			Warehouse: mcom.Warehouse{
				ID:       testWarehouseID,
				Location: testWarehouseLocation,
			},
			ResourceIDs: []string{testResourceID, ""},
		}), mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "some resourceID is empty",
		})
	}
	{ // not found resource data.
		err := dm.WarehousingStock(ctx, mcom.WarehousingStockRequest{
			Warehouse: mcom.Warehouse{
				ID:       testWarehouseIDNew,
				Location: testWarehouseLocation,
			},
			ResourceIDs: []string{testResourceNotFound},
		})

		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_RESOURCE_NOT_FOUND,
		}, err)

	}
	{ // some resource not found.
		resourceIDs := []string{testResourceNotFound, testResourceID}
		err := dm.WarehousingStock(ctx, mcom.WarehousingStockRequest{
			Warehouse: mcom.Warehouse{
				ID:       testWarehouseIDNew,
				Location: testWarehouseLocation,
			},
			ResourceIDs: resourceIDs,
		})

		assert.ErrorIs(err, mcomErr.Error{
			Code: mcomErr.Code_RESOURCE_NOT_FOUND,
		})
	}
	{ // contains different product types of material (good case)
		_, err := createMockWarehouseStockData(dm, testResourceID)
		assert.NoError(err)
		assert.NoError(createMaterialResources(db, mcom.Material{
			Type:       "T1",
			ID:         testResourceID,
			Quantity:   decimal.Zero,
			ResourceID: testResourceID,
		}, mcom.Warehouse{
			ID:       testWarehouseID,
			Location: testWarehouseLocation,
		}))
		assert.NoError(createMaterialResources(db, mcom.Material{
			Type:       "T2",
			ID:         testResourceID,
			Quantity:   decimal.Zero,
			ResourceID: testResourceID,
		}, mcom.Warehouse{
			ID:       testWarehouseID,
			Location: testWarehouseLocation,
		}))

		assert.NoError(dm.WarehousingStock(ctx, mcom.WarehousingStockRequest{
			Warehouse: mcom.Warehouse{
				ID:       testWarehouseIDNew,
				Location: testWarehouseLocation,
			},
			ResourceIDs: []string{testResourceID},
		}))

		assert.NoError(deleteAllWarehouseStocks(db))
		assert.NoError(deleteAllMaterialResources(db))
	}
	{ // warehouse stock resource not found (should pass)
		_, err = dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{{
				Type:       testResourceType,
				ID:         testResourceID,
				Quantity:   decimal.NewFromInt(10),
				ResourceID: testResourceID,
			}},
		})
		assert.NoError(err)

		rep, err := dm.GetMaterialResource(ctx, mcom.GetMaterialResourceRequest{
			ResourceID: testResourceID,
		})
		assert.NoError(err)
		assert.Len(rep, 1)
		assert.Equal(rep[0].Warehouse.ID, "")
		assert.Equal(rep[0].Warehouse.Location, "")

		resourceIDs := []string{testResourceID}
		assert.NoError(dm.WarehousingStock(ctx, mcom.WarehousingStockRequest{
			Warehouse: mcom.Warehouse{
				ID:       testWarehouseIDNew,
				Location: testWarehouseLocationNew,
			},
			ResourceIDs: resourceIDs,
		}))

		// make sure we have created the stock data before delete it.
		stock, err := queryMockCreatedWarehouseStockData(db)
		assert.NoError(err)
		assert.Equal(models.WarehouseStock{
			ID:        testWarehouseIDNew,
			ProductID: testResourceID,
			Location:  testWarehouseLocationNew,
			Quantity:  decimal.RequireFromString("10.000000"),
		}, stock)

		rep, err = dm.GetMaterialResource(ctx, mcom.GetMaterialResourceRequest{
			ResourceID: testResourceID,
		})
		assert.NoError(err)
		assert.Len(rep, 1)
		assert.Equal(rep[0].Warehouse.ID, testWarehouseIDNew)
		assert.Equal(rep[0].Warehouse.Location, testWarehouseLocationNew)

		// make sure we have created the record before delete it.
		record, err := queryMockCreatedTransportRecord(db)
		assert.NoError(err)
		assert.Equal(testResourceID, record.ResourceID)

		assert.NoError(deleteAllWarehouseStocks(db))
		assert.NoError(deleteAllMaterialResources(db))
	}
	{ // good case; update.
		_, err := createMockWarehouseStockData(dm, testResourceID)
		assert.NoError(err)

		resourceIDs := []string{testResourceID}
		err = dm.WarehousingStock(ctx, mcom.WarehousingStockRequest{
			Warehouse: mcom.Warehouse{
				ID:       testWarehouseIDNew,
				Location: testWarehouseLocationNew,
			},
			ResourceIDs: resourceIDs,
		})
		assert.NoError(err)

		// make sure we have created the stock data before delete it.
		stock, err := queryMockCreatedWarehouseStockData(db)
		assert.NoError(err)
		assert.Equal(models.WarehouseStock{
			ID:        testWarehouseIDNew,
			ProductID: "Bonjour",
			Location:  testWarehouseLocationNew,
			Quantity:  decimal.RequireFromString("20.000000"),
		}, stock)

		// make sure we have created the record before delete it.
		record, err := queryMockCreatedTransportRecord(db)
		assert.NoError(err)
		assert.Equal(testResourceID, record.ResourceID)
		assert.NoError(deleteAllMaterialResources(db))
		assert.NoError(deleteAllWarehouseStocks(db))
	}
	{ // complex case (should pass)
		// create the same product id in A-00.
		// create at the same request (should pass)
		_, err := dm.CreateMaterialResources(ctx,
			mcom.CreateMaterialResourcesRequest{
				Materials: []mcom.CreateMaterialResourcesRequestDetail{{
					Type:       "T1",
					ID:         "Product1",
					Grade:      "A",
					Status:     resources.MaterialStatus_AVAILABLE,
					Quantity:   decimal.NewFromInt(50),
					Unit:       "kg",
					ResourceID: "Resource1",
				}, {
					Type:       "T1",
					ID:         "Product1",
					Grade:      "A",
					Status:     resources.MaterialStatus_AVAILABLE,
					Quantity:   decimal.NewFromInt(50),
					Unit:       "kg",
					ResourceID: "Resource2",
				}},
			},
			mcom.WithStockIn(mcom.Warehouse{
				ID:       "A",
				Location: "00",
			}))
		assert.NoError(err)

		// check quantity.
		stocks, err := queryResourcesStockInMockData(db, []string{"Product1"})
		assert.NoError(err)
		assert.Equal(decimal.RequireFromString("100.000000"), stocks[0].Quantity)

		// create 0-quantity material test
		_, err = dm.CreateMaterialResources(ctx,
			mcom.CreateMaterialResourcesRequest{
				Materials: []mcom.CreateMaterialResourcesRequestDetail{{
					Type:       "T1",
					ID:         "Product1",
					Grade:      "A",
					Status:     resources.MaterialStatus_AVAILABLE,
					Quantity:   decimal.NewFromInt(20),
					Unit:       "kg",
					ResourceID: "Resource3",
				},
				}},
			mcom.WithStockIn(mcom.Warehouse{
				ID:       "A",
				Location: "01",
			}))
		assert.NoError(err)

		assert.NoError(createMaterialResources(db, mcom.Material{
			Type:       "T2",
			ID:         "Product2",
			Grade:      "A",
			Status:     resources.MaterialStatus_AVAILABLE,
			Quantity:   decimal.Zero,
			Unit:       "kg",
			ResourceID: "Resource1",
		}, mcom.Warehouse{
			ID:       "A",
			Location: "01",
		}))

		// 0-quantity material could be found in material resource table.
		rep, err := dm.GetMaterialResource(ctx, mcom.GetMaterialResourceRequest{
			ResourceID: "Resource1",
		})
		assert.NoError(err)
		assert.True(func() bool {
			for _, material := range rep {
				if material.Material.ID == "Product2" {
					return true
				}
			}
			return false
		}())

		// 0-quantity material could NOT be found in warehouse stock table.
		stocks, err = queryResourcesStockInMockData(db, []string{"Product2"})
		assert.NoError(err)
		assert.Len(stocks, 0)

		// create negative-quantity material
		assert.NoError(createMaterialResources(db, mcom.Material{
			Type:       "T3",
			ID:         "Product3",
			Grade:      "A",
			Status:     resources.MaterialStatus_AVAILABLE,
			Quantity:   decimal.NewFromInt(-50),
			Unit:       "kg",
			ResourceID: "Resource1",
		}, mcom.Warehouse{
			ID:       "A",
			Location: "02",
		}))

		// do warehousing stocks.
		db.Create(&models.Warehouse{ID: "B", DepartmentID: "M2100"})
		assert.NoError(dm.WarehousingStock(ctx, mcom.WarehousingStockRequest{
			Warehouse: mcom.Warehouse{
				ID:       "B",
				Location: "00",
			},
			ResourceIDs: []string{"Resource1"},
		}))

		// check stocks.
		assert.NoError(db.Where(&models.WarehouseStock{
			ID:       "B",
			Location: "00",
		}).Find(&stocks).Error)

		assert.Len(stocks, 0) // 50+0+(-50)

		// check transport record.
		var actual []comparableTransportRecord
		assert.NoError(db.Model(&models.ResourceTransportRecord{}).
			Where(&models.ResourceTransportRecord{ResourceID: "Resource1"}).Find(&actual).Error)
		expected := []comparableTransportRecord{{
			OldWarehouseID: "A",
			OldLocation:    "00",
			NewWarehouseID: "B",
			NewLocation:    "00",
			ResourceID:     "Resource1",
		}, {
			OldWarehouseID: "A",
			OldLocation:    "01",
			NewWarehouseID: "B",
			NewLocation:    "00",
			ResourceID:     "Resource1",
		}, {
			OldWarehouseID: "A",
			OldLocation:    "02",
			NewWarehouseID: "B",
			NewLocation:    "00",
			ResourceID:     "Resource1",
		}}
		assert.ElementsMatch(expected, actual)
	}

	assert.NoError(deleteAllWarehouseData(db))
	assert.NoError(deleteAllMaterialResources(db))
	assert.NoError(deleteAllWarehouseStocks(db))
	assert.NoError(deleteAllCreatedTransportRecord(db))
}

func TestDataManager_ListMaterialResourceStatus(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, _ := initializeDB(t)
	defer dm.Close()
	actual, err := dm.ListMaterialResourceStatus(ctx)
	assert.NoError(err)
	assert.Equal(mcom.ListMaterialResourceStatusReply{"AVAILABLE", "HOLD", "INSPECTION", "MOUNTED", "UNAVAILABLE"}, actual)
}

func TestDataManager_SplitMaterialResource(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)

	defer dm.Close()
	cm := newClearMaster(db, &models.MaterialResource{})
	assert.NoError(cm.Clear())
	// region bad cases.
	{ // insufficient request.
		_, err := dm.SplitMaterialResource(ctx, mcom.SplitMaterialResourceRequest{
			ProductType:   "ProductType",
			Quantity:      decimal.NewFromInt(5),
			InspectionIDs: []int{1, 2, 3},
			Remark:        "Remark",
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST})

		_, err = dm.SplitMaterialResource(ctx, mcom.SplitMaterialResourceRequest{
			ResourceID:    "SourceResourceID",
			Quantity:      decimal.NewFromInt(5),
			InspectionIDs: []int{1, 2, 3},
			Remark:        "Remark",
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST})
	}
	{ // resource not found.
		_, err := dm.SplitMaterialResource(ctx, mcom.SplitMaterialResourceRequest{
			ResourceID:    "notFound",
			ProductType:   "notFound",
			Quantity:      decimal.NewFromInt(5),
			InspectionIDs: []int{1, 2, 3},
			Remark:        "Remark",
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_RESOURCE_NOT_FOUND})
	}
	// endregion bad cases.
	// region setup.
	_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{{
			Type:       "ProductType",
			ID:         "ProductID",
			Grade:      "A",
			Status:     resources.MaterialStatus_AVAILABLE,
			Quantity:   decimal.NewFromInt(50),
			Unit:       "kg",
			LotNumber:  "",
			ResourceID: "ResourceID",
			MinDosage:  decimal.NewFromInt(2),
			Inspections: []models.Inspection{
				{
					ID:     0,
					Remark: "0",
				},
				{
					ID:     1,
					Remark: "1",
				},
				{
					ID:     2,
					Remark: "2",
				},
				{
					ID:     3,
					Remark: "3",
				},
			},
			Remark:         "source remark",
			ExpiryTime:     time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
			ProductionTime: time.Date(2022, 2, 22, 22, 22, 22, 11, time.UTC),
		}},
	}, mcom.WithStockIn(mcom.Warehouse{
		ID:       "W",
		Location: "1",
	}))
	assert.NoError(err)
	// endregion setup.
	// region good cases.
	{
		rep, err := dm.SplitMaterialResource(ctx, mcom.SplitMaterialResourceRequest{
			ResourceID:    "ResourceID",
			ProductType:   "ProductType",
			Quantity:      decimal.NewFromInt(20),
			InspectionIDs: []int{2},
			Remark:        "destination remark",
		})
		assert.NoError(err)

		var actual models.MaterialResource
		assert.NoError(db.Where(&models.MaterialResource{ID: rep.NewResourceID}).Take(&actual).Error)
		expected := models.MaterialResource{
			OID:         actual.OID,
			ID:          rep.NewResourceID,
			ProductID:   "ProductID",
			ProductType: "ProductType",
			Quantity:    decimal.RequireFromString("20.000000"),
			Status:      resources.MaterialStatus_AVAILABLE,
			ExpiryTime:  types.ToTimeNano(time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC)),
			Info: models.MaterialInfo{
				Grade:          "A",
				Unit:           "kg",
				LotNumber:      "",
				ProductionTime: time.Date(2022, 2, 22, 22, 22, 22, 11, time.UTC),
				MinDosage:      decimal.NewFromInt(0),
				Inspections: []models.Inspection{
					{
						ID:     2,
						Remark: "2",
					},
				},
				PlannedQuantity: decimal.NewFromInt(0),
				Remark:          "destination remark",
			},
			WarehouseID:       "W",
			WarehouseLocation: "1",
			UpdatedAt:         actual.UpdatedAt,
			CreatedAt:         actual.CreatedAt,
			FeedRecordsID:     []string{},
		}

		assert.Equal(expected, actual)

		actual = models.MaterialResource{}
		assert.NoError(db.Where(&models.MaterialResource{ID: "ResourceID", ProductType: "ProductType"}).Take(&actual).Error)
		expected = models.MaterialResource{
			OID:         actual.OID,
			ID:          "ResourceID",
			ProductID:   "ProductID",
			ProductType: "ProductType",
			Quantity:    decimal.RequireFromString("30.000000"),
			Status:      resources.MaterialStatus_AVAILABLE,
			ExpiryTime:  types.ToTimeNano(time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC)),
			Info: models.MaterialInfo{
				Grade:          "A",
				Unit:           "kg",
				LotNumber:      "",
				ProductionTime: time.Date(2022, 2, 22, 22, 22, 22, 11, time.UTC),
				MinDosage:      decimal.NewFromInt(2),
				Inspections: []models.Inspection{
					{
						ID:     0,
						Remark: "0",
					},
					{
						ID:     1,
						Remark: "1",
					},
					{
						ID:     3,
						Remark: "3",
					},
				},
				PlannedQuantity: decimal.NewFromInt(0),
				Remark:          "source remark",
			},
			WarehouseID:       "W",
			WarehouseLocation: "1",
			UpdatedAt:         actual.UpdatedAt,
			CreatedAt:         actual.CreatedAt,
			FeedRecordsID:     []string{},
		}

		assert.Equal(expected, actual)
		assert.Equal(expected.Quantity, actual.Quantity)
	}
	// endregion good cases.
	assert.NoError(cm.Clear())
}

type comparableTransportRecord struct {
	OldWarehouseID string
	OldLocation    string
	NewWarehouseID string
	NewLocation    string
	ResourceID     string
}

func TestDataManager_ListMaterialResourcesById(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &models.MaterialResource{})
	assert.NoError(cm.Clear())
	// #region setup
	now := time.Now()
	monkey.Patch(time.Now, func() time.Time {
		return now
	})
	defer monkey.UnpatchAll()
	_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{{
			Type:           "A",
			ID:             "A",
			Grade:          "A",
			Status:         resources.MaterialStatus_AVAILABLE,
			Quantity:       decimal.NewFromInt(50),
			Unit:           "kg",
			LotNumber:      "1",
			ProductionTime: now,
			ExpiryTime:     now,
			ResourceID:     "A",
			MinDosage:      decimal.NewFromInt(5),
			Inspections:    models.Inspections{},
			Remark:         "A",
		},
			{
				Type:           "B",
				ID:             "B",
				Grade:          "B",
				Status:         resources.MaterialStatus_AVAILABLE,
				Quantity:       decimal.NewFromInt(50),
				Unit:           "kg",
				LotNumber:      "1",
				ProductionTime: now,
				ExpiryTime:     now,
				ResourceID:     "B",
				MinDosage:      decimal.NewFromInt(5),
				Inspections:    models.Inspections{},
				Remark:         "B",
			},
			{
				Type:           "B2",
				ID:             "B",
				Grade:          "B",
				Status:         resources.MaterialStatus_AVAILABLE,
				Quantity:       decimal.NewFromInt(50),
				Unit:           "kg",
				LotNumber:      "1",
				ProductionTime: now,
				ExpiryTime:     now,
				ResourceID:     "B",
				MinDosage:      decimal.NewFromInt(5),
				Inspections:    models.Inspections{},
				Remark:         "B",
			},
			{
				Type:           "C",
				ID:             "C",
				Grade:          "C",
				Status:         resources.MaterialStatus_AVAILABLE,
				Quantity:       decimal.NewFromInt(50),
				Unit:           "kg",
				LotNumber:      "1",
				ProductionTime: now,
				ExpiryTime:     now,
				ResourceID:     "C",
				MinDosage:      decimal.NewFromInt(5),
				Inspections:    models.Inspections{},
				Remark:         "C",
			}},
	})
	assert.NoError(err)
	// #endregion setup
	{ // good case.
		actual, err := dm.ListMaterialResourcesById(ctx, mcom.ListMaterialResourcesByIdRequest{
			ResourcesID: []string{"A", "C", "D", "B"},
		})
		assert.NoError(err)

		expectedMap1 := make(map[string]mcom.MaterialReply)
		expectedMap2 := make(map[string]mcom.MaterialReply)
		expectedMap3 := make(map[string]mcom.MaterialReply)
		expectedMap4 := make(map[string]mcom.MaterialReply)
		expectedMap1["A"] = mcom.MaterialReply{
			Material: mcom.Material{
				Type:            "A",
				ID:              "A",
				Grade:           "A",
				Status:          resources.MaterialStatus_AVAILABLE,
				Quantity:        decimal.RequireFromString("50.000000"),
				PlannedQuantity: decimal.NewFromInt(0),
				Unit:            "kg",
				LotNumber:       "1",
				ProductionTime:  now.Local(),
				ExpiryTime:      now.Local(),
				ResourceID:      "A",
				MinDosage:       decimal.NewFromInt(5),
				Inspections:     models.Inspections{},
				Remark:          "A",
				UpdatedAt:       types.ToTimeNano(now),
				CreatedAt:       types.ToTimeNano(now),
				FeedRecordsID:   []string{},
			},
		}
		expectedMap2["C"] = mcom.MaterialReply{
			Material: mcom.Material{
				Type:            "C",
				ID:              "C",
				Grade:           "C",
				Status:          resources.MaterialStatus_AVAILABLE,
				Quantity:        decimal.RequireFromString("50.000000"),
				PlannedQuantity: decimal.NewFromInt(0),
				Unit:            "kg",
				LotNumber:       "1",
				ProductionTime:  now.Local(),
				ExpiryTime:      now.Local(),
				ResourceID:      "C",
				MinDosage:       decimal.NewFromInt(5),
				Inspections:     models.Inspections{},
				Remark:          "C",
				UpdatedAt:       types.ToTimeNano(now),
				CreatedAt:       types.ToTimeNano(now),
				FeedRecordsID:   []string{},
			},
		}
		expectedMap4["B"] = mcom.MaterialReply{
			Material: mcom.Material{
				Type:            "B",
				ID:              "B",
				Grade:           "B",
				Status:          resources.MaterialStatus_AVAILABLE,
				Quantity:        decimal.RequireFromString("50.000000"),
				PlannedQuantity: decimal.NewFromInt(0),
				Unit:            "kg",
				LotNumber:       "1",
				ProductionTime:  now.Local(),
				ExpiryTime:      now.Local(),
				ResourceID:      "B",
				MinDosage:       decimal.NewFromInt(5),
				Inspections:     models.Inspections{},
				Remark:          "B",
				UpdatedAt:       types.ToTimeNano(now),
				CreatedAt:       types.ToTimeNano(now),
				FeedRecordsID:   []string{},
			},
		}
		expectedMap4["B2"] = mcom.MaterialReply{
			Material: mcom.Material{
				Type:            "B2",
				ID:              "B",
				Grade:           "B",
				Status:          resources.MaterialStatus_AVAILABLE,
				Quantity:        decimal.RequireFromString("50.000000"),
				PlannedQuantity: decimal.NewFromInt(0),
				Unit:            "kg",
				LotNumber:       "1",
				ProductionTime:  now.Local(),
				ExpiryTime:      now.Local(),
				ResourceID:      "B",
				MinDosage:       decimal.NewFromInt(5),
				Inspections:     models.Inspections{},
				Remark:          "B",
				UpdatedAt:       types.ToTimeNano(now),
				CreatedAt:       types.ToTimeNano(now),
				FeedRecordsID:   []string{},
			},
		}
		expected := mcom.ListMaterialResourcesByIdReply{
			TypeResourcePairs: []map[string]mcom.MaterialReply{
				expectedMap1,
				expectedMap2,
				expectedMap3,
				expectedMap4,
			},
		}
		assert.Equal(expected, actual)
	}
	assert.NoError(cm.Clear())
}
