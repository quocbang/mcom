package impl

import (
	"context"
	"reflect"
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
)

const (
	testResourceMaterialType      = "prs"
	testResourceMaterialID        = "Bonjour"
	testResourceMaterialGrade     = ""
	testResourceMaterialStatus    = resources.MaterialStatus_INSPECTION
	testResourceMaterialQuantity  = "20"
	testResourceMaterialUnit      = "piece"
	testResourceMaterialLotNumber = "777-1101"

	testCreateResourceWarehouseID       = "A"
	testCreateResourceWarehouseLocation = "10"

	testSpecifyResourceID = "XXX123"
)

var (
	testResourceProductionTime = time.Date(2021, 8, 8, 0, 0, 0, 0, time.Local)
	testResourceExpirationTime = time.Date(2021, 8, 11, 0, 0, 0, 0, time.Local)
)

func queryMaterialResourcesMockData(db *gorm.DB, res []mcom.CreatedMaterialResource) ([]models.MaterialResource, error) {
	ids := make([]string, len(res))
	for i, v := range res {
		ids[i] = v.ID
	}
	var resources []models.MaterialResource
	if err := db.Where(` id in ? `, ids).
		Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

func queryResourcesStockInMockData(db *gorm.DB, ids []string) ([]models.WarehouseStock, error) {
	var stocks []models.WarehouseStock
	if err := db.Where(` product_id in ? `, ids).
		Find(&stocks).Error; err != nil {
		return nil, err
	}
	return stocks, nil
}

func deleteResourcesStockInMockData(db *gorm.DB) error {
	if err := db.Where(` 1=1 `).
		Delete(&models.WarehouseStock{}).Error; err != nil {
		return err
	}
	return nil
}

func createSpecifyResourceIDMockData(db *gorm.DB) error {
	return db.Create(&models.MaterialResource{
		ID:            testSpecifyResourceID,
		ProductType:   testResourceMaterialType,
		Quantity:      decimal.Zero,
		Info:          models.MaterialInfo{},
		UpdatedBy:     "x",
		FeedRecordsID: []string{},
	}).Error
}

func deleteAllMaterialResources(db *gorm.DB) error {
	return newClearMaster(db, &models.MaterialResource{}).Clear()
}

func TestDataManager_CreateMaterialResources(t *testing.T) {
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

	sDB, err := db.DB()
	if !assert.NoError(err) {
		return
	}
	defer sDB.Close()
	assert.NoError(deleteAllMaterialResources(db))
	assert.NoError(deleteAllWarehouseStocks(db))

	{ // nil request error.
		_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{Materials: nil})

		assert.ErrorIs(err, mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "there is no resource to add",
		})
	}
	{ // unique violation (create resource)
		assert.NoError(createSpecifyResourceIDMockData(db))

		req := mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{
				{
					Type:           testResourceMaterialType,
					ID:             testResourceMaterialID,
					Grade:          testResourceMaterialGrade,
					Status:         testResourceMaterialStatus,
					Quantity:       decimal.RequireFromString(testResourceMaterialQuantity),
					Unit:           testResourceMaterialUnit,
					LotNumber:      testResourceMaterialLotNumber,
					ProductionTime: testResourceProductionTime,
					ExpiryTime:     testResourceExpirationTime,
					ResourceID:     testSpecifyResourceID,
				},
			},
		}
		_, err = dm.CreateMaterialResources(ctx, req)
		assert.ErrorIs(err, mcomErr.Error{
			Code:    mcomErr.Code_RESOURCE_EXISTED,
			Details: "resources to create: \nResource ID: XXX123, Product Type: prs\n",
		})
		assert.NoError(deleteAllMaterialResources(db))
	}
	{ // good case; create material resource without stock in.
		req := mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{{
				Type:           testResourceMaterialType,
				ID:             testResourceMaterialID,
				Grade:          testResourceMaterialGrade,
				Status:         testResourceMaterialStatus,
				Quantity:       decimal.RequireFromString(testResourceMaterialQuantity),
				Unit:           testResourceMaterialUnit,
				LotNumber:      testResourceMaterialLotNumber,
				ProductionTime: testResourceProductionTime,
				ExpiryTime:     testResourceExpirationTime,
				ResourceID:     testSpecifyResourceID,
			}},
		}
		result, err := dm.CreateMaterialResources(ctx, req)
		assert.NoError(err)
		assert.Len(result, 1)

		// make sure we have created resources before delete it.
		resources, err := queryMaterialResourcesMockData(db, result)
		assert.NoError(err)
		assert.Len(resources, 1)
		assert.Equal(result[0].ID, resources[0].ID)
		assert.Equal(result[0].OID, resources[0].OID)

		assert.NoError(deleteAllMaterialResources(db))
	}
	{ // good case; create material resource with stock in.
		monkey.PatchInstanceMethod(reflect.TypeOf(&models.MaterialResource{}), "BeforeCreate", func(r *models.MaterialResource, db *gorm.DB) error {
			r.OID = resourceOID
			return nil
		})

		req := mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{{
				Type:           testResourceMaterialType,
				ID:             testResourceMaterialID,
				Grade:          testResourceMaterialGrade,
				Status:         testResourceMaterialStatus,
				Quantity:       decimal.RequireFromString(testResourceMaterialQuantity),
				Unit:           testResourceMaterialUnit,
				LotNumber:      testResourceMaterialLotNumber,
				ProductionTime: testResourceProductionTime,
				ExpiryTime:     testResourceExpirationTime,
				ResourceID:     testSpecifyResourceID,
			}},
		}
		result, err := dm.CreateMaterialResources(ctx, req,
			mcom.WithStockIn(mcom.Warehouse{
				ID:       testCreateResourceWarehouseID,
				Location: testCreateResourceWarehouseLocation,
			}))
		assert.NoError(err)
		assert.Len(result, 1)
		monkey.UnpatchAll()

		// make sure we have created resources stock in data before delete it.
		var stock models.WarehouseStock
		assert.NoError(db.Where(&models.WarehouseStock{
			ID:        testCreateResourceWarehouseID,
			Location:  testCreateResourceWarehouseLocation,
			ProductID: testResourceMaterialID,
		}).Take(&stock).Error)

		// for decimal numeric(16, 6) comparison.
		assert.Equal(decimal.RequireFromString(testResourceMaterialQuantity+".000000"), stock.Quantity)

		assert.NoError(deleteResourcesStockInMockData(db))

		// make sure we have created resources before delete it.
		resources, err := queryMaterialResourcesMockData(db, result)
		assert.NoError(err)
		assert.Len(resources, 1)
		assert.Equal(result[0].ID, resources[0].ID)
		assert.Equal(result[0].OID, resources[0].OID)

		assert.NoError(deleteAllMaterialResources(db))
	}
	{ // good case: with feed records
		t := time.Date(1990, 9, 9, 9, 9, 9, 9, time.Local)
		_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{
				{
					Type:           testResourceMaterialType,
					ID:             testResourceMaterialID,
					Grade:          testResourceMaterialGrade,
					Status:         testResourceMaterialStatus,
					Quantity:       decimal.NewFromInt(10),
					Station:        "station",
					Unit:           testResourceMaterialUnit,
					LotNumber:      testResourceMaterialLotNumber,
					ProductionTime: testResourceProductionTime,
					ExpiryTime:     t,
					ResourceID:     testSpecifyResourceID,
					MinDosage:      decimal.NewFromInt(10),
					FeedRecordID:   "rec1",
				},
			},
		})
		assert.NoError(err)

		rep, err := dm.GetMaterialResource(ctx, mcom.GetMaterialResourceRequest{
			ResourceID: testSpecifyResourceID,
		})
		assert.NoError(err)

		assert.Equal(mcom.GetMaterialResourceReply{
			{
				Material: mcom.Material{
					Type:            testResourceMaterialType,
					ID:              testResourceMaterialID,
					Grade:           testResourceMaterialGrade,
					Status:          testResourceMaterialStatus,
					Quantity:        decimal.New(10000000, -6),
					PlannedQuantity: decimal.New(0, 0),
					Station:         "station",
					Unit:            testResourceMaterialUnit,
					LotNumber:       testResourceMaterialLotNumber,
					ProductionTime:  testResourceProductionTime,
					ExpiryTime:      t,
					ResourceID:      testSpecifyResourceID,
					MinDosage:       decimal.NewFromInt(10),
					UpdatedAt:       rep[0].Material.UpdatedAt,
					CreatedAt:       rep[0].Material.CreatedAt,
					FeedRecordsID:   []string{"rec1"},
				},
				Warehouse: mcom.Warehouse{},
			},
		}, rep)

		// test append
		_, err = dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{
				{
					Type:           testResourceMaterialType,
					ID:             testResourceMaterialID,
					Grade:          testResourceMaterialGrade,
					Status:         testResourceMaterialStatus,
					Quantity:       decimal.NewFromInt(10),
					Station:        "station",
					Unit:           testResourceMaterialUnit,
					LotNumber:      testResourceMaterialLotNumber,
					ProductionTime: testResourceProductionTime,
					ExpiryTime:     t,
					ResourceID:     testSpecifyResourceID,
					MinDosage:      decimal.NewFromInt(10),
					FeedRecordID:   "rec2",
				},
			},
		})
		assert.NoError(err)

		rep, err = dm.GetMaterialResource(ctx, mcom.GetMaterialResourceRequest{
			ResourceID: testSpecifyResourceID,
		})
		assert.NoError(err)

		assert.Equal(mcom.GetMaterialResourceReply{
			{
				Material: mcom.Material{
					Type:            testResourceMaterialType,
					ID:              testResourceMaterialID,
					Grade:           testResourceMaterialGrade,
					Status:          testResourceMaterialStatus,
					Quantity:        decimal.New(20000000, -6),
					PlannedQuantity: decimal.New(0, 0),
					Station:         "station",
					Unit:            testResourceMaterialUnit,
					LotNumber:       testResourceMaterialLotNumber,
					ProductionTime:  testResourceProductionTime,
					ExpiryTime:      t,
					ResourceID:      testSpecifyResourceID,
					MinDosage:       decimal.NewFromInt(10),
					UpdatedAt:       rep[0].Material.UpdatedAt,
					CreatedAt:       rep[0].Material.CreatedAt,
					FeedRecordsID:   []string{"rec1", "rec2"},
				},
				Warehouse: mcom.Warehouse{},
			},
		}, rep)
		assert.NoError(deleteAllMaterialResources(db))
	}
	{ // good case: 0 actual quantity case (should pass)
		result, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{{
				Type:            testResourceMaterialType,
				ID:              testResourceMaterialID,
				Grade:           testResourceMaterialGrade,
				Status:          testResourceMaterialStatus,
				Quantity:        decimal.Zero,
				PlannedQuantity: decimal.NewFromInt(10),
				Unit:            testResourceMaterialUnit,
				LotNumber:       testResourceMaterialLotNumber,
				ProductionTime:  testResourceProductionTime,
				ExpiryTime:      testResourceExpirationTime,
				ResourceID:      testSpecifyResourceID,
			}},
		}, mcom.WithStockIn(mcom.Warehouse{
			ID:       testWarehouseID,
			Location: testWarehouseLocation,
		}))
		assert.NoError(err)
		assert.Len(result, 1)

		actual, err := dm.GetMaterialResource(ctx, mcom.GetMaterialResourceRequest{
			ResourceID: result[0].ID,
		})
		assert.NoError(err)
		assert.Equal(decimal.RequireFromString("0.000000"), actual[0].Material.Quantity)
		assert.Equal(decimal.NewFromInt(10), actual[0].Material.PlannedQuantity)
	}
	{ // bad case: negative quantity
		_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{{
				Type:       testResourceMaterialType,
				ID:         testResourceMaterialID,
				Grade:      testResourceMaterialGrade,
				Status:     testResourceMaterialStatus,
				Quantity:   decimal.NewFromInt(-2),
				Unit:       testResourceMaterialUnit,
				LotNumber:  testResourceMaterialLotNumber,
				ResourceID: testSpecifyResourceID,
			}},
		}, mcom.WithStockIn(mcom.Warehouse{
			ID:       testWarehouseID,
			Location: testWarehouseLocation,
		}))
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_INVALID_NUMBER})
	}
	{ // bad case: both the actual quantity and the planned quantity equal to 0
		_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{{
				Type:            testResourceMaterialType,
				ID:              testResourceMaterialID,
				Grade:           testResourceMaterialGrade,
				Status:          testResourceMaterialStatus,
				Quantity:        decimal.Zero,
				PlannedQuantity: decimal.Zero,
				Unit:            testResourceMaterialUnit,
				LotNumber:       testResourceMaterialLotNumber,
				ResourceID:      testSpecifyResourceID,
			}},
		}, mcom.WithStockIn(mcom.Warehouse{
			ID:       testWarehouseID,
			Location: testWarehouseLocation,
		}))
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_INVALID_NUMBER})
	}
}
