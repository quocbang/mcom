package impl

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	pbWorkOrder "gitlab.kenda.com.tw/kenda/commons/v2/proto/golang/mes/v2/workorder"
	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

const (
	recordTestOperatorID   = "123456"
	recordTestDepartmentID = "T1234"
	recordTestResourceID   = "test_resource"

	feedRecipeID   = "test_recipe"
	feedProcessOID = "12345678-1234-1234-1234-123456789012"
	feedStation    = "test"
	feedSiteName   = "test"
	feedSiteIndex  = 0

	resourceOID          = "1a3c907b-3d12-4d65-c888-522bbfc02f02"
	resourceID           = "A123456"
	resourceOIDNotExists = "6f1c907b-2c63-4d65-b888-522bbfc02f02"

	testRecordWorkOrderID = "a5d8e3z5a4w1q2a3s5z7"
	testRecordLotNumber   = "a1b2c3d4"
)

var (
	recordsWorkDate = time.Date(2021, 01, 01, 0, 0, 0, 0, time.Local)
)

func createWorkOrderForListFeedRecordTest(db *gorm.DB, departmentID string) (*models.WorkOrder, error) {
	workOrder := &models.WorkOrder{
		RecipeID:     feedRecipeID,
		ProcessOID:   feedProcessOID,
		DepartmentID: departmentID,
		Status:       pbWorkOrder.Status_CLOSED,
		Station:      feedStation,
		ReservedDate: recordsWorkDate,
		UpdatedBy:    recordTestOperatorID,
		CreatedBy:    recordTestOperatorID,
	}
	if err := db.Create(workOrder).Error; err != nil {
		return nil, err
	}
	return workOrder, nil
}

func createBatchForListFeedRecordTest(db *gorm.DB, workOrderID string, resourceOID string) (*models.Batch, error) {
	recID := strings.ToUpper(xid.New().String())

	recs := []models.FeedRecord{{
		ID:         recID,
		OperatorID: recordTestOperatorID,
		Materials: []models.FeedDetail{
			{
				Resources: []models.FeedResource{
					{
						ResourceID:  "test_resource",
						ProductID:   recordTestResourceID,
						ProductType: "RUBB",
						Grade:       "A",
						Quantity:    decimal.NewFromFloat(123.456),
						ExpiryTime:  types.TimeNano(12345).Time(),
					},
				},
				UniqueSite: models.UniqueSite{
					Station: feedStation,
					SiteID: models.SiteID{
						Name:  feedSiteName,
						Index: feedSiteIndex,
					},
				},
			},
		},
		Time: recordsWorkDate,
	}}

	if err := db.Model(&models.FeedRecord{}).Create(&recs).Error; err != nil {
		return nil, err
	}

	batch := &models.Batch{
		WorkOrder: workOrderID,
		Number:    1,
		Status:    pbWorkOrder.BatchStatus_BATCH_CLOSED,
		RecordsID: []string{recID},
		UpdatedBy: recordTestOperatorID,
	}
	if err := db.Create(batch).Error; err != nil {
		return nil, err
	}
	return batch, nil
}

func createResourceForListFeedRecordTest(db *gorm.DB) (*models.MaterialResource, error) {
	var e *models.MaterialResource
	monkey.PatchInstanceMethod(reflect.TypeOf(e), "BeforeCreate", func(r *models.MaterialResource, db *gorm.DB) error {
		r.OID = resourceOID
		return nil
	})
	resource := &models.MaterialResource{
		ID:          recordTestResourceID,
		ProductID:   recordTestResourceID,
		ProductType: "RUBB",
		Quantity:    decimal.Zero,
		ExpiryTime:  12345,
		Info: models.MaterialInfo{
			Grade: "A",
		},
		UpdatedBy:     recordTestOperatorID,
		FeedRecordsID: []string{},
	}
	if err := db.Create(&resource).Error; err != nil {
		return nil, err
	}
	monkey.UnpatchAll()
	return resource, nil
}

func createMockFeedRecord(db *gorm.DB) (*models.WorkOrder, *models.Batch, error) {
	// create mock department.
	dep := &models.Department{
		ID: recordTestDepartmentID,
	}
	err := db.Create(dep).Error
	if err != nil {
		return nil, nil, err
	}

	// create mock resource.
	res, err := createResourceForListFeedRecordTest(db)
	if err != nil {
		return nil, nil, err
	}

	// create mock work order.
	workOrder, err := createWorkOrderForListFeedRecordTest(db, dep.ID)
	if err != nil {
		return nil, nil, err
	}

	// create mock batch with bad resource oid.
	batch, err := createBatchForListFeedRecordTest(db, workOrder.ID, res.OID)
	if err != nil {
		return nil, nil, err
	}

	return workOrder, batch, nil
}

func deleteDataForListFeedRecordTest(db *gorm.DB) error {
	return newClearMaster(db, &models.Department{}, &models.WorkOrder{}, &models.MaterialResource{}, &models.Batch{}).Clear()
}

func Test_ListFeedRecord(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	dm, db, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}

	sDB, err := db.DB()
	if !assert.NoError(err) {
		return
	}
	defer sDB.Close()

	assert.NoError(deleteDataForListFeedRecordTest(db))

	workOrder, batch, err := createMockFeedRecord(db)
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(workOrder) {
		return
	}
	if !assert.NotNil(batch) {
		return
	}

	{ // insufficient request: Date, DepartmentID.
		_, err := dm.ListFeedRecords(ctx, mcom.ListRecordsRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // resource not found.
		mockDM := &DataManager{db: db}
		mockSession := mockDM.newSession(ctx)
		_, err := mockSession.getResourceInfo(resourceOIDNotExists)
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_RESOURCE_NOT_FOUND,
		}, err)
	}
	{ // good case.
		feedRecord, err := dm.ListFeedRecords(ctx, mcom.ListRecordsRequest{
			Date:         recordsWorkDate,
			DepartmentID: recordTestDepartmentID,
		})
		assert.NoError(err)
		assert.Equal(mcom.ListFeedRecordReply{{
			WorkOrder:   workOrder.ID,
			RecipeID:    feedRecipeID,
			Batch:       int32(batch.Number),
			ProcessName: workOrder.ProcessName,
			ProcessType: workOrder.ProcessType,
			StationID:   feedStation,
			Materials: []mcom.FeedMaterial{{
				ID:          recordTestResourceID,
				Grade:       "A",
				ResourceID:  recordTestResourceID,
				ProductType: "RUBB",
				Quantity:    decimal.NewFromFloat(123.456),
				Site: models.SiteID{
					Name:  feedSiteName,
					Index: feedSiteIndex,
				},
				ExpiryTime: 12345,
			}},
		}}, feedRecord)
	}

	assert.NoError(deleteDataForListFeedRecordTest(db))
}

func (session *session) getResourceInfo(oid string) (models.MaterialResource, error) {
	var results models.MaterialResource
	err := session.db.
		Where(` "oid" = ? `, oid).
		Take(&results).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.MaterialResource{}, mcomErr.Error{
				Code: mcomErr.Code_RESOURCE_NOT_FOUND,
			}
		}
		return models.MaterialResource{}, err
	}
	return results, nil
}

func deleteCollectRecordMockData(db *gorm.DB) error {
	return newClearMaster(db, &models.MaterialResource{}, &models.CollectRecord{}, &models.WorkOrder{}, &models.Department{}).Clear()
}

func Test_CreateCollectRecord(t *testing.T) {
	assert := assert.New(t)
	ctx := commonsCtx.WithUserID(context.Background(), testUser)
	dm, db, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}
	defer dm.Close()

	assert.NoError(deleteCollectRecordMockData(db))

	sDB, err := db.DB()
	if !assert.NoError(err) {
		return
	}
	defer sDB.Close()

	timeNow := time.Now()
	monkey.Patch(time.Now, func() time.Time {
		return timeNow
	})
	defer monkey.UnpatchAll()

	{ // insufficient request.
		err := dm.CreateCollectRecord(ctx, mcom.CreateCollectRecordRequest{})
		assert.ErrorIs(err, mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		})
	}
	{ // good case.
		r := models.MaterialResource{
			ID:          recordTestResourceID,
			ProductID:   testProductID,
			ProductType: "RUBB",
			Status:      0,
			Quantity:    decimal.NewFromInt(600),
			ExpiryTime:  types.ToTimeNano(time.Now()),
			Info: models.MaterialInfo{
				Grade:          "A",
				Unit:           "kg",
				LotNumber:      testRecordLotNumber,
				ProductionTime: time.Now(),
			},
			UpdatedBy:     testUser,
			FeedRecordsID: []string{},
		}
		err := db.Create(&r).Error
		assert.NoError(err)

		err = dm.CreateCollectRecord(ctx, mcom.CreateCollectRecordRequest{
			WorkOrder:   testRecordWorkOrderID,
			Sequence:    1,
			LotNumber:   testRecordLotNumber,
			Station:     testStationA,
			ResourceOID: r.OID,
			Quantity:    decimal.NewFromFloat(123.456),
			BatchCount:  6,
		})
		assert.NoError(err)

		// check result.
		collectRecord := models.CollectRecord{}
		err = db.Where(` work_order = ? AND sequence = ? AND lot_number = ? `,
			testRecordWorkOrderID, 1, testRecordLotNumber,
		).Find(&collectRecord).Error
		assert.NoError(err)

		resource := models.MaterialResource{}
		err = db.Where(` id = ? `, recordTestResourceID).Find(&resource).Error
		assert.NoError(err)

		expected := models.CollectRecord{
			WorkOrder:   testRecordWorkOrderID,
			Sequence:    1,
			LotNumber:   testRecordLotNumber,
			Station:     testStationA,
			ResourceOID: resource.OID,
			Detail: models.CollectRecordDetail{
				OperatorID: testUser,
				Quantity:   decimal.NewFromFloat(123.456),
				BatchCount: 6,
			},
			CreatedAt: types.ToTimeNano(timeNow),
		}
		assert.Equal(expected, collectRecord)
	}
	{ // record already existed.
		err = dm.CreateCollectRecord(ctx, mcom.CreateCollectRecordRequest{
			WorkOrder:   testRecordWorkOrderID,
			Sequence:    1,
			LotNumber:   testRecordLotNumber,
			Station:     testStationA,
			ResourceOID: resourceOID,
			Quantity:    decimal.NewFromFloat(123.456),
			BatchCount:  6,
		})
		assert.ErrorIs(err, mcomErr.Error{
			Code: mcomErr.Code_RECORD_ALREADY_EXISTS,
			Details: fmt.Sprintf("collect record already exists, work_order: %s, sequence: %d, lot_number: %s",
				testRecordWorkOrderID, 1, testRecordLotNumber,
			),
		})
	}

	assert.NoError(deleteCollectRecordMockData(db))
}

func Test_GetCollectRecord(t *testing.T) {
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

	timeNow := time.Now()
	monkey.Patch(time.Now, func() time.Time {
		return timeNow
	})
	defer monkey.UnpatchAll()
	{ // insufficient request: missing the work order id.
		_, err := dm.GetCollectRecord(ctx, mcom.GetCollectRecordRequest{
			WorkOrder: "",
			Sequence:  1,
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // insufficient request: missing the sequence.
		_, err := dm.GetCollectRecord(ctx, mcom.GetCollectRecordRequest{
			WorkOrder: testRecordWorkOrderID,
			Sequence:  0,
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // record not found.
		_, err := dm.GetCollectRecord(ctx, mcom.GetCollectRecordRequest{
			WorkOrder: testRecordWorkOrderID,
			Sequence:  1,
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_RECORD_NOT_FOUND,
		}, err)
	}
	{ // find the record.
		productID := "AAA"
		productType := "QQ"
		collectedQuantity := decimal.NewFromFloat(123.456)

		// create mock data.
		rs, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{
				{
					Type:           productType,
					ID:             productID,
					Status:         resources.MaterialStatus_AVAILABLE,
					Quantity:       decimal.NewFromFloat(789.456),
					Unit:           "KG",
					LotNumber:      testRecordLotNumber,
					ProductionTime: time.Now(),
					ExpiryTime:     time.Now().Add(time.Hour),
				},
			},
		})
		assert.NoError(err)
		assert.NoError(db.Create(&models.CollectRecord{
			WorkOrder:   testRecordWorkOrderID,
			Sequence:    1,
			LotNumber:   testRecordLotNumber,
			Station:     feedStation,
			ResourceOID: rs[0].OID,
			Detail: models.CollectRecordDetail{
				OperatorID: recordTestOperatorID,
				Quantity:   collectedQuantity,
				BatchCount: 1,
			},
		}).Error)
		monkey.PatchInstanceMethod(reflect.TypeOf(&DataManager{}), "GetWorkOrder", func(dm *DataManager, ctx context.Context, req mcom.GetWorkOrderRequest) (mcom.GetWorkOrderReply, error) {
			return mcom.GetWorkOrderReply{
				ID: req.ID,
				Product: mcom.Product{
					ID:   productID,
					Type: productType,
				},
				RecipeID: feedRecipeID,
			}, nil
		})

		record, err := dm.GetCollectRecord(ctx, mcom.GetCollectRecordRequest{
			WorkOrder: testRecordWorkOrderID,
			Sequence:  1,
		})
		assert.NoError(err)
		assert.Equal(mcom.GetCollectRecordReply{
			ResourceID: rs[0].ID,
			LotNumber:  testRecordLotNumber,
			WorkOrder:  testRecordWorkOrderID,
			RecipeID:   feedRecipeID,
			ProductID:  productID,
			BatchCount: 1,
			Quantity:   collectedQuantity,
			OperatorID: recordTestOperatorID,
		}, record)

		monkey.UnpatchAll()
	}
}

func Test_ListCollectRecord(t *testing.T) {
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

	assert.NoError(deleteCollectRecordMockData(db))

	sDB, err := db.DB()
	if !assert.NoError(err) {
		return
	}
	defer sDB.Close()

	timeNow := time.Now()
	monkey.Patch(time.Now, func() time.Time {
		return timeNow
	})
	defer monkey.UnpatchAll()
	{ // insufficient request: date.
		_, err := dm.ListCollectRecords(ctx, mcom.ListRecordsRequest{
			DepartmentID: testDepartmentID,
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // insufficient request: department.
		_, err := dm.ListCollectRecords(ctx, mcom.ListRecordsRequest{
			Date: timeNow,
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // department not found.
		_, err := dm.ListCollectRecords(ctx, mcom.ListRecordsRequest{
			Date:         timeNow,
			DepartmentID: testDepartmentID,
		})
		assert.NoError(err)
	}
	{ // resource not found.
		// create mock data.
		err := db.Create(&models.Department{
			ID: testDepartmentID,
		}).Error
		assert.NoError(err)

		err = db.Create(&models.WorkOrder{
			RecipeID:     testRecipeID,
			ProcessOID:   testProcessOID,
			DepartmentID: testDepartmentID,
			Station:      testStationA,
			ReservedDate: timeNow,
			UpdatedBy:    testUser,
			CreatedBy:    testUser,
		}).Error
		assert.NoError(err)

		workOrder := models.WorkOrder{}
		err = db.Where(`station = ?`, testStationA).Find(&workOrder).Error
		assert.NoError(err)

		err = db.Create(&models.CollectRecord{
			WorkOrder:   workOrder.ID,
			Sequence:    1,
			LotNumber:   testRecordLotNumber,
			Station:     testStationA,
			ResourceOID: resourceOID,
			Detail: models.CollectRecordDetail{
				OperatorID: testUser,
				Quantity:   decimal.NewFromInt(600),
				BatchCount: 3,
			},
		}).Error
		assert.NoError(err)

		_, err = dm.ListCollectRecords(ctx, mcom.ListRecordsRequest{
			Date:         timeNow,
			DepartmentID: testDepartmentID,
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_RESOURCE_NOT_FOUND,
		}, err)
	}
	{ // good case.
		workOrder := models.WorkOrder{}
		err = db.Where(`station = ?`, testStationA).Find(&workOrder).Error
		assert.NoError(err)

		var e *models.MaterialResource
		monkey.PatchInstanceMethod(reflect.TypeOf(e), "BeforeCreate", func(r *models.MaterialResource, db *gorm.DB) error {
			r.OID = resourceOID
			return nil
		})

		err := db.Create(&models.MaterialResource{
			ID:          resourceID,
			ProductID:   testProductID,
			ProductType: "RUBB",
			Status:      0,
			Quantity:    decimal.Zero,
			ExpiryTime:  types.ToTimeNano(time.Now()),
			Info: models.MaterialInfo{
				Grade:          "",
				Unit:           "",
				LotNumber:      testRecordLotNumber,
				ProductionTime: time.Time{},
			},
			UpdatedBy:     testUser,
			FeedRecordsID: []string{},
		}).Error
		assert.NoError(err)

		results, err := dm.ListCollectRecords(ctx, mcom.ListRecordsRequest{
			Date:         timeNow,
			DepartmentID: testDepartmentID,
		})
		assert.NoError(err)
		assert.Equal(mcom.ListCollectRecordsReply{{
			ResourceID: resourceID,
			WorkOrder:  workOrder.ID,
			RecipeID:   testRecipeID,
			ProductID:  testProductID,
			BatchCount: 3,
			Quantity:   decimal.NewFromInt(600),
		}}, results)
	}

	assert.NoError(deleteCollectRecordMockData(db))
}
