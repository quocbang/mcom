package impl

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/commons/v2/proto/golang/mes/v2/workorder"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/bindtype"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func TestDataManager_CreateBatch(t *testing.T) {
	// #region Arrange
	ctx, dm, db := initializeDB(t)
	defer dm.Close()

	assert := assert.New(t)
	// #endregion Arrange
	{ // case : ordinary case (should pass)
		// #region Act
		testRequest := mcom.CreateBatchRequest{WorkOrder: "test work order", Number: 1, Note: "test note"}
		assert.NoError(dm.CreateBatch(ctx, testRequest), "func error : Datamanager.CreateBatch")
		actual, err := dm.ListBatches(ctx, mcom.ListBatchesRequest{
			WorkOrder: testRequest.WorkOrder,
		})
		assert.NoError(err)
		// #endregion Act
		// #region Assert
		assert.Len(actual, 1)
		expected := models.Batch{
			WorkOrder: testRequest.WorkOrder,
			Number:    testRequest.Number,
			RecordsID: []string{},
			Note:      testRequest.Note,
			Status:    workorder.BatchStatus_BATCH_PREPARING,
		}
		actualBatch := convertBatchInfoToBatch(actual[0])
		assertBatch(expected, actualBatch, assert)
		// #endregion Assert
		clearBatch(db, assert)
	}
	{ // case : with status.
		// #region Act
		testRequest := mcom.CreateBatchRequest{
			WorkOrder: "test work order",
			Number:    1, Note: "test note",
			Status: workorder.BatchStatus_BATCH_CANCELLED,
		}
		assert.NoError(dm.CreateBatch(ctx, testRequest), "func error : Datamanager.CreateBatch")
		actual, err := dm.ListBatches(ctx, mcom.ListBatchesRequest{
			WorkOrder: testRequest.WorkOrder,
		})
		assert.NoError(err)
		// #endregion Act
		// #region Assert
		assert.Len(actual, 1)
		expected := models.Batch{
			WorkOrder: testRequest.WorkOrder,
			Number:    testRequest.Number,
			RecordsID: []string{},
			Note:      testRequest.Note,
			Status:    workorder.BatchStatus_BATCH_CANCELLED,
		}
		actualBatch := convertBatchInfoToBatch(actual[0])
		assertBatch(expected, actualBatch, assert)
		// #endregion Assert
		clearBatch(db, assert)
	}
	{ // case : create with empty workorder (should not pass)
		// #region Act
		testRequest := mcom.CreateBatchRequest{Number: 1, Note: "test note"}
		actual := dm.CreateBatch(ctx, testRequest)
		// #endregion Act
		// #region Assert
		expected := mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "empty work order is not allowed"}
		assert.ErrorIs(expected, actual, "error mismatch")
		// #endregion Assert
		clearBatch(db, assert)
	}
	{ // case : batch already exists (should not pass)
		// #region Act
		testRequest := mcom.CreateBatchRequest{WorkOrder: "test work order", Number: 1, Note: "test note"}
		assert.NoError(dm.CreateBatch(ctx, testRequest))
		actual := dm.CreateBatch(ctx, testRequest)
		// #endregion Act
		// #region Assert
		expected := mcomErr.Error{Code: mcomErr.Code_BATCH_ALREADY_EXISTS}
		assert.ErrorIs(expected, actual, "error mismatch")
		// #endregion Assert
		clearBatch(db, assert)
	}
	{ // case : Number should be greater than zero (should not pass)
		// #region Act
		testRequest := mcom.CreateBatchRequest{WorkOrder: "test work order", Number: -4, Note: "test note"}
		actual := dm.CreateBatch(ctx, testRequest)
		// #endregion Act
		// #region Assert
		expected := mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "number should be greater than zero"}
		assert.ErrorIs(expected, actual, "error mismatch")
		// #endregion Assert
		clearBatch(db, assert)
	}
}

func Test_GetBatch(t *testing.T) {
	ctx, dm, db := initializeDB(t)
	defer dm.Close()

	assert := assert.New(t)
	cm := newClearMaster(db, &models.Batch{})
	assert.NoError(cm.Clear())
	// #region bad cases
	{ // batch not found.
		_, err := dm.GetBatch(ctx, mcom.GetBatchRequest{
			WorkOrder: "not found",
			Number:    100,
		})
		assert.ErrorIs(mcomErr.Error{Code: mcomErr.Code_BATCH_NOT_FOUND}, err)
	}
	// #endregion bad cases
	// #region good cases
	{
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{
			WorkOrder: "01234567890123456789",
			Number:    1,
			Note:      "test get batch",
		}))

		actual, err := dm.GetBatch(ctx, mcom.GetBatchRequest{
			WorkOrder: "01234567890123456789",
			Number:    1,
		})
		assert.NoError(err)

		expected := mcom.GetBatchReply{
			Info: mcom.BatchInfo{
				WorkOrder: "01234567890123456789",
				Number:    1,
				Records:   []models.FeedRecord{},
				Note:      "test get batch",
				UpdatedAt: actual.Info.UpdatedAt,
			},
		}
		assert.Equal(expected, actual)
	}
	// #endregion good cases
	assert.NoError(cm.Clear())
}

func TestDataManager_ListBatches(t *testing.T) {
	// #region Arrange
	ctx, dm, db := initializeDB(t)
	defer dm.Close()

	assert := assert.New(t)
	// #endregion Arrange
	{ // case : empty work order.
		// #region Arrange
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo1", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo2", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo3", Number: 3}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo4", Number: 1, Note: "note_wo4"}))
		// #endregion Arrange
		// #region Act
		testRequest := mcom.ListBatchesRequest{}
		_, actual := dm.ListBatches(ctx, testRequest)
		expected := mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "empty work order is not allowed"}
		// #endregion Act
		// #region Assert
		assert.ErrorIs(expected, actual, "error mismatch")
		// #endregion Assert
		clearBatch(db, assert)
	}
	{ // case : search by work order - return single(should pass)
		// #region Arrange
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo1", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo2", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo3", Number: 3}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo4", Number: 1, Note: "note_wo4"}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "01234567890123456789", Number: 4, Note: "the target"}))
		// #endregion Arrange
		// #region Act
		testRequest := mcom.ListBatchesRequest{WorkOrder: "01234567890123456789"}

		actual, err := dm.ListBatches(ctx, testRequest)
		assert.NoError(err, "func error : DataManager.ListBatches")
		// #endregion Act
		// #region Assert
		expected := []mcom.BatchInfo{
			{
				WorkOrder: "01234567890123456789",
				Number:    4,
				Records:   []models.FeedRecord{},
				Note:      "the target",
			},
		}
		assertListBatchesReply(expected, actual, assert)
		// #endregion Assert
		clearBatch(db, assert)
	}
	{ // case : search by worker order - return multiple(should pass)
		// #region Arrange
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo1", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo2", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo3", Number: 3}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo4", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo4", Number: 3, Note: "note_wo4"}))
		// #endregion Arrange
		// #region Act
		testRequest := mcom.ListBatchesRequest{WorkOrder: "wo4"}

		actual, err := dm.ListBatches(ctx, testRequest)
		assert.NoError(err)
		// #endregion Act
		// #region Assert
		expected := []mcom.BatchInfo{
			{WorkOrder: "wo4", Number: 1, Records: []models.FeedRecord{}},
			{WorkOrder: "wo4", Number: 3, Note: "note_wo4", Records: []models.FeedRecord{}},
		}

		assertListBatchesReply(expected, actual, assert)
		// #endregion Assert
		clearBatch(db, assert)
	}
	{ // case : target batch does not exist (should pass)
		// #region Arrange
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo1", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo2", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo3", Number: 3}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo4", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo4", Number: 3, Note: "note_wo4"}))
		// #endregion Arrange
		// #region Act
		testRequest := mcom.ListBatchesRequest{WorkOrder: "not found"}
		actual, err := dm.ListBatches(ctx, testRequest)
		assert.NoError(err, "func error : DataManager.ListBatches")
		// #endregion Act
		// #region Assert
		expected := []mcom.BatchInfo{}
		assertListBatchesReply(expected, actual, assert)
		// #endregion Assert
		clearBatch(db, assert)
	}
	{ // case test order (should pass)
		// #region Arrange
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo2", Number: 5}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo2", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo2", Number: 4}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo2", Number: 2}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo2", Number: 3}))
		// #endregion Arrange
		// #region Act
		testRequest := mcom.ListBatchesRequest{WorkOrder: "wo2"}
		actual, err := dm.ListBatches(ctx, testRequest)
		assert.NoError(err, "func error : DataManager.ListBatches")
		// #endregion Act
		// #region Assert
		expected := []mcom.BatchInfo{
			{
				WorkOrder: "wo2",
				Number:    1,
				Records:   []models.FeedRecord{},
			},
			{
				WorkOrder: "wo2",
				Number:    2,
				Records:   []models.FeedRecord{},
			},
			{
				WorkOrder: "wo2",
				Number:    3,
				Records:   []models.FeedRecord{},
			},
			{
				WorkOrder: "wo2",
				Number:    4,
				Records:   []models.FeedRecord{},
			},
			{
				WorkOrder: "wo2",
				Number:    5,
				Records:   []models.FeedRecord{},
			}}
		assertListBatchesReply(expected, actual, assert)
		// #endregion Assert
		clearBatch(db, assert)
	}
	{ // case: should get correct feed records
		// this case will be tested in TestDataManager_Feed, case name: "test ListBatches feed records reply (it is easier to use existing feed records than to create new test records)"
	}
}

func TestDataManager_UpdateBatch(t *testing.T) {
	// #region Arrange
	ctx, dm, db := initializeDB(t)
	defer dm.Close()

	assert := assert.New(t)
	// #endregion
	{ // case : edit number (should pass)
		// #region Arrange
		// targetWorkOrder uses a magic string because of the 20-char limit
		targetWorkOrder := "01234567890123456789"
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: targetWorkOrder, Number: 2, Note: "the target"}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo1", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo2", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo3", Number: 3}))
		// #endregion Arrange
		// #region Act
		assert.NoError(dm.UpdateBatch(ctx, mcom.UpdateBatchRequest{WorkOrder: targetWorkOrder, Number: 2, Note: "New Note"}), "func error : DataManager.UpdateBatch")
		var actual models.Batch
		assert.NoError(db.Where(&models.Batch{WorkOrder: targetWorkOrder}).First(&actual).Error, "gorm function executing error")
		// #endregion Act
		// #region Assert
		expected := models.Batch{
			WorkOrder: targetWorkOrder,
			Number:    2,
			Note:      "New Note",
			RecordsID: []string{},
		}
		assertBatch(expected, actual, assert)
		// #endregion Assert
		clearBatch(db, assert)
	}
	{ // case : empty work order (should not pass)
		// #region Arrange
		// targetWorkOrder uses a magic string because of the 20-char limit
		targetWorkOrder := "01234567890123456789"
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: targetWorkOrder, Number: 2, Note: "the target"}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo1", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo2", Number: 1}))
		assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{WorkOrder: "wo3", Number: 3}))
		// #endregion Arrange
		// #region Act
		actual := dm.UpdateBatch(ctx, mcom.UpdateBatchRequest{WorkOrder: "", Number: 3})
		// #endregion Act
		// #region Assert
		expected := mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "empty work order or empty number"}
		assert.ErrorIs(expected, actual, "error mismatch")
		clearBatch(db, assert)
		// #endregion Assert
	}
}

func TestDataManager_Feed(t *testing.T) {
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	assert := assert.New(t)
	cm := newClearMaster(db, &models.WorkOrder{}, &models.Site{}, &models.SiteContents{}, &models.Station{}, &models.Batch{}, &models.RecipeProcessDefinition{}, &models.Recipe{}, &models.MaterialResource{}, &models.Warehouse{}, &models.WarehouseStock{})
	assert.NoError(cm.Clear())

	type resourceQuantity struct {
		Quantity decimal.Decimal
	}
	{ // invalid number (should not pass)
		_, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{
				WorkOrder: "Hello",
				Number:    1,
			},
			FeedContent: []mcom.FeedPerSite{mcom.FeedPerSiteType1{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "site1",
						Index: 0,
					},
					Station: "station",
				},
				FeedAll:  false,
				Quantity: decimal.NewFromInt(-123),
			}},
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_INVALID_NUMBER})
	}

	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            "station",
		DepartmentOID: "m2100",
		Sites: []mcom.SiteInformation{{
			Name:    "slot",
			Index:   0,
			Type:    sites.Type_SLOT,
			SubType: sites.SubType_MATERIAL,
		}, {
			Name:    "container",
			Index:   0,
			Type:    sites.Type_CONTAINER,
			SubType: sites.SubType_MATERIAL,
		}, {
			Name:    "collection",
			Index:   0,
			Type:    sites.Type_COLLECTION,
			SubType: sites.SubType_MATERIAL,
		}, {
			Name:    "colqueue",
			Index:   0,
			Type:    sites.Type_COLQUEUE,
			SubType: sites.SubType_MATERIAL,
		}, {
			Name:    "queue",
			Index:   0,
			Type:    sites.Type_QUEUE,
			SubType: sites.SubType_MATERIAL,
		}},
		State: stations.State_IDLE,
	}))

	expiryB := time.Now()
	_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{
			{
				Type:       "A",
				ID:         "A",
				Grade:      "A",
				Status:     resources.MaterialStatus_AVAILABLE,
				Quantity:   decimal.RequireFromString("100"),
				Unit:       "kg",
				ResourceID: "A",
			}, {
				Type:       "B",
				ID:         "B",
				Grade:      "B",
				Status:     resources.MaterialStatus_AVAILABLE,
				Quantity:   decimal.RequireFromString("100"),
				Unit:       "kg",
				ResourceID: "B",
				ExpiryTime: expiryB,
			}, {
				Type:       "C",
				ID:         "C",
				Grade:      "C",
				Status:     resources.MaterialStatus_AVAILABLE,
				Quantity:   decimal.RequireFromString("100"),
				Unit:       "kg",
				ResourceID: "C",
			}}}, mcom.WithStockIn(mcom.Warehouse{ID: testWarehouseID, Location: testWarehouseLocation}))
	assert.NoError(err)

	releasedAt := types.ToTimeNano(time.Now())
	assert.NoError(dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
		Recipes: []mcom.Recipe{{
			ID: "recipeID",
			Product: mcom.Product{
				ID:   "product name",
				Type: "product type",
			},
			Version: mcom.RecipeVersion{
				Major:      "1",
				Minor:      "1",
				Stage:      "1",
				ReleasedAt: releasedAt,
			},
			Processes: []*mcom.Process{
				{
					OID:           "f310b071-fb63-464a-9cfc-f60d599a92d4",
					OptionalFlows: []*mcom.RecipeOptionalFlow{},
				},
			},
			ProcessDefinitions: []mcom.ProcessDefinition{{
				OID:  "f310b071-fb63-464a-9cfc-f60d599a92d4",
				Name: "process name",
				Type: "process type",
				Output: mcom.OutputProduct{
					ID:   "product name",
					Type: "product type",
				},
			}},
		}},
	}))

	wo, err := dm.CreateWorkOrders(ctx, mcom.CreateWorkOrdersRequest{
		WorkOrders: []mcom.CreateWorkOrder{{
			ProcessOID:      "f310b071-fb63-464a-9cfc-f60d599a92d4",
			RecipeID:        "recipeID",
			ProcessName:     "process name",
			ProcessType:     "process type",
			Status:          workorder.Status_ACTIVE,
			DepartmentOID:   "m2100",
			Station:         "station",
			Sequence:        123,
			BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{decimal.NewFromInt(5)}),
			Date:            time.Now(),
			Unit:            "kg",
			Note:            "note",
		}}})
	assert.NoError(err)

	// #region create batch

	createTestBatches := func(number int) {
		for i := 1; i <= number; i++ {
			assert.NoError(dm.CreateBatch(ctx, mcom.CreateBatchRequest{
				WorkOrder: wo.IDs[0],
				Number:    int16(i),
				Note:      "test batch",
			}))
		}
	}
	createTestBatches(9)

	// #endregion create batch

	{ // internal error collection site can only feed all materials. (should not pass)
		_, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{
				WorkOrder: wo.IDs[0],
				Number:    1,
			},
			FeedContent: []mcom.FeedPerSite{mcom.FeedPerSiteType1{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "collection",
						Index: 0,
					},
					Station: "station",
				},
				FeedAll:  false,
				Quantity: decimal.NewFromInt(15),
			}},
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "a collection site can only feed all materials"})
	}
	{ // for a slot site (should pass)
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "station",
			Details: []mcom.MaterialBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "slot",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{{
						Material: models.Material{
							ID:    "A",
							Grade: "A",
						},
						Quantity:    types.Decimal.NewFromInt32(10),
						ResourceID:  "A",
						ProductType: "A",
						ExpiryTime:  2222,
						Warehouse:   mcom.Warehouse{},
					}},
					Option: models.BindOption{},
				},
			},
		}))

		feedRep, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{
				WorkOrder: wo.IDs[0],
				Number:    1,
			},
			FeedContent: []mcom.FeedPerSite{mcom.FeedPerSiteType1{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "slot",
						Index: 0,
					},
					Station: "station",
				},
				FeedAll:  false,
				Quantity: decimal.NewFromInt(5),
			}},
		})
		assert.NoError(err)
		assert.NotZero(feedRep)

		rep, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "slot",
			SiteIndex: 0,
		})
		assert.NoError(err)

		assert.Equal(rep.Content.Slot.Material, &models.MaterialSite{
			Material: models.Material{
				ID:    "A",
				Grade: "A",
			},
			Quantity:    types.Decimal.RequireFromString("5"),
			ResourceID:  "A",
			ProductType: "A",
			ExpiryTime:  2222,
		})
	}
	{ // slot feed with unspecified quantity.
		assert.NoError(dm.WarehousingStock(ctx, mcom.WarehousingStockRequest{
			Warehouse: mcom.Warehouse{
				ID:       "w",
				Location: "l",
			},
			ResourceIDs: []string{"A"},
		}))
		var stock models.WarehouseStock
		assert.NoError(dm.(*DataManager).db.Where(models.WarehouseStock{
			ID:       "w",
			Location: "l",
		}).Take(&stock).Error)
		assert.Equal(decimal.RequireFromString("90.000000"), stock.Quantity)
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "station",
			Details: []mcom.MaterialBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "slot",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{{
						Material: models.Material{
							ID:    "A",
							Grade: "A",
						},
						Quantity:    nil,
						ResourceID:  "A",
						ProductType: "A",
						ExpiryTime:  2222,
						Warehouse: mcom.Warehouse{
							ID:       "w",
							Location: "l",
						},
					}},
					Option: models.BindOption{},
				},
			},
		}))

		assert.NoError(dm.(*DataManager).db.Where(models.WarehouseStock{
			ID:       "w",
			Location: "l",
		}).Take(&stock).Error)
		assert.Equal(decimal.RequireFromString("95.000000"), stock.Quantity)

		feedRep, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{
				WorkOrder: wo.IDs[0],
				Number:    1,
			},
			FeedContent: []mcom.FeedPerSite{mcom.FeedPerSiteType1{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "slot",
						Index: 0,
					},
					Station: "station",
				},
				FeedAll:  false,
				Quantity: decimal.NewFromInt(5),
			}},
		})
		assert.NoError(err)
		assert.NotZero(feedRep)

		rep, err := dm.GetMaterialResource(ctx, mcom.GetMaterialResourceRequest{
			ResourceID: "A",
		})
		assert.NoError(err)

		assert.Equal(decimal.RequireFromString("90.000000"), rep[0].Material.Quantity)

		assert.NoError(dm.(*DataManager).db.Where(models.WarehouseStock{
			ID:       "w",
			Location: "l",
		}).Take(&stock).Error)
		assert.Equal(decimal.RequireFromString("90.000000"), stock.Quantity)
	}
	{ // for a container site (should pass)
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "station",
			Details: []mcom.MaterialBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_CONTAINER_BIND,
					Site: models.SiteID{
						Name:  "container",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{{
						Material: models.Material{
							ID:    "A",
							Grade: "A",
						},
						Quantity:    types.Decimal.NewFromInt32(10),
						ResourceID:  "A",
						ProductType: "A",
						ExpiryTime:  0,
						Warehouse:   mcom.Warehouse{},
					}},
					Option: models.BindOption{},
				},
			},
		}))

		feedRep, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{
				WorkOrder: wo.IDs[0],
				Number:    1,
			},
			FeedContent: []mcom.FeedPerSite{mcom.FeedPerSiteType1{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "container",
						Index: 0,
					},
					Station: "station",
				},
				FeedAll:  false,
				Quantity: decimal.NewFromInt(5),
			}},
		})
		assert.NoError(err)
		assert.NotZero(feedRep)

		rep, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "container",
			SiteIndex: 0,
		})
		assert.NoError(err)

		assert.Equal(rep.Content.Container, &models.Container{{
			Material: &models.MaterialSite{
				Material: models.Material{
					ID:    "A",
					Grade: "A",
				},
				Quantity:    types.Decimal.RequireFromString("5"),
				ResourceID:  "A",
				ProductType: "A",
			},
		}})
	}
	{ // for a collection site (should pass)
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "station",
			Details: []mcom.MaterialBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_COLLECTION_BIND,
					Site: models.SiteID{
						Name:  "collection",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{{
						Material: models.Material{
							ID:    "B",
							Grade: "B",
						},
						Quantity:    types.Decimal.NewFromInt32(10),
						ResourceID:  "B",
						ProductType: "B",
						ExpiryTime:  0,
						Warehouse: mcom.Warehouse{
							ID:       testWarehouseID,
							Location: testWarehouseLocation,
						},
					}},
					Option: models.BindOption{},
				},
			},
		}))

		feedRep, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{
				WorkOrder: wo.IDs[0],
				Number:    2,
			},
			FeedContent: []mcom.FeedPerSite{mcom.FeedPerSiteType1{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "collection",
						Index: 0,
					},
					Station: "station",
				},
				FeedAll: true,
			}},
		})
		assert.NoError(err)
		assert.NotZero(feedRep)

		rep, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "collection",
			SiteIndex: 0,
		})
		assert.NoError(err)

		assert.Equal(&models.Collection{}, rep.Content.Collection)
	}
	{ // for a colqueue site (should pass)
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "station",
			Details: []mcom.MaterialBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_BIND,
					Site: models.SiteID{
						Name:  "colqueue",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{{
						Material: models.Material{
							ID:    "C",
							Grade: "C",
						},
						Quantity:    types.Decimal.NewFromInt32(10),
						ResourceID:  "C",
						ProductType: "C",
						ExpiryTime:  0,
						Warehouse:   mcom.Warehouse{},
					}},
					Option: models.BindOption{
						Head: true,
					},
				},
			},
		}))

		feedRep, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{
				WorkOrder: wo.IDs[0],
				Number:    3,
			},
			FeedContent: []mcom.FeedPerSite{mcom.FeedPerSiteType1{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "colqueue",
						Index: 0,
					},
					Station: "station",
				},
				FeedAll: true,
			}},
		})
		assert.NoError(err)
		assert.NotZero(feedRep)

		rep, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "colqueue",
			SiteIndex: 0,
		})
		assert.NoError(err)

		assert.Equal(&models.Colqueue{}, rep.Content.Colqueue)
	}
	{ // for a queue site (should pass)
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "station",
			Details: []mcom.MaterialBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSH,
					Site: models.SiteID{
						Name:  "queue",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{{
						Material: models.Material{
							ID:    "C",
							Grade: "C",
						},
						Quantity:    types.Decimal.NewFromInt32(10),
						ResourceID:  "C",
						ProductType: "C",
						Status:      resources.MaterialStatus_HOLD,
						ExpiryTime:  0,
						Warehouse:   mcom.Warehouse{},
					}},
					Option: models.BindOption{
						Head: true,
					},
				},
			},
		}))

		feedRep, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{
				WorkOrder: wo.IDs[0],
				Number:    3,
			},
			FeedContent: []mcom.FeedPerSite{mcom.FeedPerSiteType1{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "queue",
						Index: 0,
					},
					Station: "station",
				},
				FeedAll: true,
			}},
		})
		assert.NoError(err)
		assert.NotZero(feedRep)

		rep, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "queue",
			SiteIndex: 0,
		})
		assert.NoError(err)

		assert.Equal(&models.Queue{}, rep.Content.Queue)
	}
	{ // for multiple sites (should pass)
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "station",
			Details: []mcom.MaterialBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "slot",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{{
						Material: models.Material{
							ID:    "A",
							Grade: "A",
						},
						Quantity:    types.Decimal.NewFromInt32(5),
						ResourceID:  "A",
						ProductType: "A",
						ExpiryTime:  2222,
						Warehouse:   mcom.Warehouse{},
					}},
					Option: models.BindOption{},
				},
			},
		}))

		feedRep, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{WorkOrder: wo.IDs[0], Number: 4},
			FeedContent: []mcom.FeedPerSite{
				mcom.FeedPerSiteType1{
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "slot",
							Index: 0,
						},
						Station: "station",
					},
					FeedAll: true,
				},
				mcom.FeedPerSiteType1{
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "container",
							Index: 0,
						},
						Station: "station",
					},
					FeedAll: true,
				},
			},
		})
		assert.NoError(err)
		assert.NotZero(feedRep)
	}
	{ // feed type 2.
		var qtyB resourceQuantity
		var qtyBStock resourceQuantity

		assert.NoError(db.Model(&models.MaterialResource{}).Where(models.MaterialResource{ID: "B"}).Take(&qtyB).Error)
		assert.Equal(decimal.RequireFromString("90.000000"), qtyB.Quantity)

		assert.NoError(db.Model(&models.WarehouseStock{}).Where(models.WarehouseStock{
			ID:        testWarehouseID,
			Location:  testWarehouseLocation,
			ProductID: "B",
		}).Take(&qtyBStock).Error)
		assert.Equal(decimal.RequireFromString("90.000000"), qtyBStock.Quantity)

		feedRep, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{
				WorkOrder: wo.IDs[0],
				Number:    5,
			},
			FeedContent: []mcom.FeedPerSite{mcom.FeedPerSiteType2{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "slot",
						Index: 0,
					},
					Station: "station",
				},
				Quantity:    decimal.NewFromInt(10),
				ResourceID:  "B",
				ProductType: "B",
			}},
		})
		assert.NoError(err)
		assert.NotZero(feedRep)

		assert.NoError(db.Model(&models.MaterialResource{}).Where(models.MaterialResource{ID: "B"}).Take(&qtyB).Error)
		assert.Equal(decimal.RequireFromString("80.000000"), qtyB.Quantity)

		assert.NoError(db.Model(&models.WarehouseStock{}).Where(models.WarehouseStock{
			ID:        testWarehouseID,
			Location:  testWarehouseLocation,
			ProductID: "B",
		}).Take(&qtyBStock).Error)
		assert.Equal(decimal.RequireFromString("80.000000"), qtyBStock.Quantity)
	}
	{ // feed type 2 with resources not found. (without work order)
		rep, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{},
			FeedContent: []mcom.FeedPerSite{mcom.FeedPerSiteType2{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "slot",
						Index: 0,
					},
					Station: "station",
				},
				Quantity:    decimal.NewFromInt(10),
				ResourceID:  "not found",
				ProductType: "not found",
			}},
		})
		assert.NoError(err)
		assert.NotZero(rep)
	}
	{ // feed type 3.
		var qtyB resourceQuantity
		assert.NoError(db.Model(&models.MaterialResource{}).Where(models.MaterialResource{ID: "B"}).Take(&qtyB).Error)
		assert.Equal(decimal.RequireFromString("80.000000"), qtyB.Quantity)

		rep, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{
				WorkOrder: wo.IDs[0],
				Number:    7,
			},
			FeedContent: []mcom.FeedPerSite{mcom.FeedPerSiteType3{
				Quantity:    decimal.NewFromInt(10),
				ResourceID:  "B",
				ProductType: "B",
			}},
		})
		assert.NoError(err)
		assert.NotZero(rep)

		assert.NoError(db.Model(&models.MaterialResource{}).Where(models.MaterialResource{ID: "B"}).Take(&qtyB).Error)
		assert.Equal(decimal.RequireFromString("70.000000"), qtyB.Quantity)
	}
	{ // feed with resources not found in database
		rep, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{
				WorkOrder: wo.IDs[0],
				Number:    8,
			},
			FeedContent: []mcom.FeedPerSite{
				mcom.FeedPerSiteType2{
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "slot",
							Index: 0,
						},
						Station: "station",
					},
					Quantity:    decimal.NewFromInt(20),
					ResourceID:  "not_found",
					ProductType: "",
				},
			},
		})
		assert.NoError(err)
		assert.NotZero(rep)
	}
	{ // empty site name
		assert.NoError(dm.UpdateStation(ctx, mcom.UpdateStationRequest{
			ID:            "station",
			DepartmentOID: "m2100",
			Sites: []mcom.UpdateStationSite{
				{
					ActionMode: sites.ActionType_ADD,
					Information: mcom.SiteInformation{
						Station:    "station",
						Name:       "",
						Index:      0,
						Type:       sites.Type_SLOT,
						SubType:    sites.SubType_MATERIAL,
						Limitation: []string{},
					},
				},
			},
			State: stations.State_IDLE,
		}))

		assert.NoError(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "",
							Index: 0,
						},
						Station: "station",
					},
					Resources: []mcom.BindMaterialResource{
						{
							Material: models.Material{
								ID:    "A",
								Grade: "A",
							},
							Quantity:    types.Decimal.NewFromInt32(10),
							ResourceID:  "A",
							ProductType: "A",
							Status:      resources.MaterialStatus_AVAILABLE,
							ExpiryTime:  0,
							Warehouse: mcom.Warehouse{
								ID:       "w",
								Location: "l",
							},
						},
					},
					Option: models.BindOption{},
				},
			},
		}))

		rep, err := dm.Feed(ctx, mcom.FeedRequest{
			Batch: mcom.BatchID{
				WorkOrder: wo.IDs[0],
				Number:    9,
			},
			FeedContent: []mcom.FeedPerSite{
				mcom.FeedPerSiteType1{
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "",
							Index: 0,
						},
						Station: "station",
					},
					FeedAll: true,
				},
			},
		})
		assert.NoError(err)
		assert.NotZero(rep)
	}
	{ // verify feed record.
		y, m, d := time.Now().Date()
		rep, err := dm.ListFeedRecords(ctx, mcom.ListRecordsRequest{
			Date:         time.Date(y, m, d, 0, 0, 0, 0, time.UTC),
			DepartmentID: "m2100",
		})
		assert.NoError(err)

		expected := mcom.ListFeedRecordReply{
			{
				WorkOrder:   wo.IDs[0],
				Batch:       1,
				RecipeID:    "recipeID",
				ProcessName: "process name",
				ProcessType: "process type",
				StationID:   "station",
				Materials: []mcom.FeedMaterial{
					{
						ID:          "A",
						Grade:       "A",
						ResourceID:  "A",
						ProductType: "A",
						Quantity:    decimal.RequireFromString("5"),
						Site: models.SiteID{
							Name:  "slot",
							Index: 0,
						},
						ExpiryTime: 2222,
					},
					{
						ID:          "A",
						Grade:       "A",
						ResourceID:  "A",
						ProductType: "A",
						Quantity:    decimal.RequireFromString("5"),
						Site: models.SiteID{
							Name:  "slot",
							Index: 0,
						},
						ExpiryTime: 2222,
					},
					{
						ID:          "A",
						Grade:       "A",
						ResourceID:  "A",
						ProductType: "A",
						Quantity:    decimal.RequireFromString("5"),
						Site: models.SiteID{
							Name:  "container",
							Index: 0,
						},
						ExpiryTime: 0,
					},
				},
			}, {
				WorkOrder:   wo.IDs[0],
				Batch:       2,
				RecipeID:    "recipeID",
				ProcessName: "process name",
				ProcessType: "process type",
				StationID:   "station",
				Materials: []mcom.FeedMaterial{{
					ID:          "B",
					Grade:       "B",
					ResourceID:  "B",
					ProductType: "B",
					Quantity:    decimal.RequireFromString("10"),
					Site: models.SiteID{
						Name:  "collection",
						Index: 0,
					},
					ExpiryTime: 0,
				}},
			}, {
				WorkOrder:   wo.IDs[0],
				Batch:       3,
				RecipeID:    "recipeID",
				ProcessName: "process name",
				ProcessType: "process type",
				StationID:   "station",
				Materials: []mcom.FeedMaterial{{
					ID:          "C",
					Grade:       "C",
					ResourceID:  "C",
					ProductType: "C",
					Quantity:    decimal.RequireFromString("10"),
					Site: models.SiteID{
						Name:  "colqueue",
						Index: 0,
					},
					ExpiryTime: 0,
				}, {
					ID:          "C",
					Grade:       "C",
					ResourceID:  "C",
					ProductType: "C",
					Quantity:    decimal.RequireFromString("10"),
					Status:      resources.MaterialStatus_HOLD,
					Site: models.SiteID{
						Name:  "queue",
						Index: 0,
					},
					ExpiryTime: 0,
				}},
			}, {
				WorkOrder:   wo.IDs[0],
				Batch:       4,
				RecipeID:    "recipeID",
				ProcessName: "process name",
				ProcessType: "process type",
				StationID:   "station",
				Materials: []mcom.FeedMaterial{{
					ID:          "A",
					Grade:       "A",
					ResourceID:  "A",
					ProductType: "A",
					Quantity:    decimal.RequireFromString("5"),
					Site: models.SiteID{
						Name:  "slot",
						Index: 0,
					},
					ExpiryTime: 2222,
				}, {
					ID:          "A",
					Grade:       "A",
					ResourceID:  "A",
					ProductType: "A",
					Quantity:    decimal.RequireFromString("5"),
					Site: models.SiteID{
						Name:  "container",
						Index: 0,
					},
					ExpiryTime: 0,
				}},
			}, {
				WorkOrder:   wo.IDs[0],
				Batch:       5,
				RecipeID:    "recipeID",
				ProcessName: "process name",
				ProcessType: "process type",
				StationID:   "station",
				Materials: []mcom.FeedMaterial{{
					ID:          "B",
					Grade:       "B",
					ResourceID:  "B",
					ProductType: "B",
					Quantity:    decimal.RequireFromString("10"),
					Site: models.SiteID{
						Name:  "slot",
						Index: 0,
					},
					ExpiryTime: types.ToTimeNano(expiryB),
					Status:     resources.MaterialStatus_AVAILABLE,
				}},
			}, {
				WorkOrder:   wo.IDs[0],
				Batch:       6,
				RecipeID:    "recipeID",
				ProcessName: "process name",
				ProcessType: "process type",
				StationID:   "station",
				// todo 用新的方法查詢
				Materials: nil,
			}, {
				WorkOrder:   wo.IDs[0],
				Batch:       7,
				RecipeID:    "recipeID",
				ProcessName: "process name",
				ProcessType: "process type",
				StationID:   "station",
				Materials: []mcom.FeedMaterial{{
					ID:          "B",
					Grade:       "B",
					ResourceID:  "B",
					ProductType: "B",
					Quantity:    decimal.RequireFromString("10"),
					ExpiryTime:  types.ToTimeNano(expiryB),
					Status:      resources.MaterialStatus_AVAILABLE,
				}},
			}, {
				WorkOrder:   wo.IDs[0],
				Batch:       8,
				RecipeID:    "recipeID",
				ProcessName: "process name",
				ProcessType: "process type",
				StationID:   "station",
				Materials: []mcom.FeedMaterial{{
					ID:          "",
					Grade:       "",
					ResourceID:  "not_found",
					ProductType: "",
					Quantity:    decimal.RequireFromString("20"),
					Site: models.SiteID{
						Name:  "slot",
						Index: 0,
					},
					ExpiryTime: 0,
				}},
			}, {
				WorkOrder:   wo.IDs[0],
				Batch:       9,
				RecipeID:    "recipeID",
				ProcessName: "process name",
				ProcessType: "process type",
				StationID:   "station",
				Materials: []mcom.FeedMaterial{{
					ID:          "A",
					Grade:       "A",
					ResourceID:  "A",
					ProductType: "A",
					Quantity:    decimal.RequireFromString("10"),
					Site: models.SiteID{
						Name:  "",
						Index: 0,
					},
					Status:     resources.MaterialStatus_AVAILABLE,
					ExpiryTime: 0,
				}},
			}}

		assert.ElementsMatch(expected, rep)
	}
	{ // test ListBatches feed records reply (it is easier to use existing feed records than to create new test records)
		rep, err := dm.ListBatches(ctx, mcom.ListBatchesRequest{
			WorkOrder: wo.IDs[0],
		})
		assert.NoError(err)
		filterBatchNumber := func(rep mcom.ListBatchesReply, number int16) mcom.BatchInfo {
			for _, b := range rep {
				if b.Number == number {
					return b
				}
			}
			return mcom.BatchInfo{}
		}

		// it is trivial to confirm all records.
		// the 3rd batch which has two feed records is a nice object to be tested.
		actual := filterBatchNumber(rep, 3)
		assert.Equal(mcom.BatchInfo{
			WorkOrder: wo.IDs[0],
			Number:    3,
			Status:    int32(workorder.Status_PENDING),
			Records: []models.FeedRecord{
				{
					ID:         actual.Records[0].ID,
					OperatorID: "",
					Materials: models.FeedDetails{
						{
							Resources: []models.FeedResource{
								{
									ResourceID:  "C",
									ProductID:   "C",
									ProductType: "C",
									Grade:       "C",
									Status:      0,
									ExpiryTime:  actual.Records[0].Materials[0].Resources[0].ExpiryTime,
									Quantity:    decimal.NewFromInt(10),
								},
							}, UniqueSite: models.UniqueSite{
								SiteID: models.SiteID{
									Name:  "colqueue",
									Index: 0,
								},
								Station: "station",
							},
						},
					},
					Time: actual.Records[0].Time,
				},
				{
					ID:         actual.Records[1].ID,
					OperatorID: "",
					Materials: models.FeedDetails{
						{
							Resources: []models.FeedResource{
								{
									ResourceID:  "C",
									ProductID:   "C",
									ProductType: "C",
									Grade:       "C",
									Status:      2,
									ExpiryTime:  actual.Records[1].Materials[0].Resources[0].ExpiryTime,
									Quantity:    decimal.NewFromInt(10),
								},
							}, UniqueSite: models.UniqueSite{
								SiteID: models.SiteID{
									Name:  "queue",
									Index: 0,
								},
								Station: "station",
							},
						},
					},
					Time: actual.Records[1].Time,
				}},
			Note:      "test batch",
			UpdatedAt: actual.UpdatedAt,
		}, actual)

	}
	assert.NoError(cm.Clear())
}

// #region Private_functions_for_batch_test

func clearBatch(db *gorm.DB, assert *assert.Assertions) {
	assert.NoError(newClearMaster(db, &models.Batch{}).Clear())
}

func assertBatch(batch1 models.Batch, batch2 models.Batch, assert *assert.Assertions) bool {
	return assert.Equal(strings.TrimRight(batch1.WorkOrder, " "), strings.TrimRight(batch2.WorkOrder, " ")) &&
		assert.Equal(batch1.Note, batch2.Note) &&
		assert.Equal(batch1.Number, batch2.Number) &&
		assert.Equal(batch1.RecordsID, batch2.RecordsID) &&
		assert.Equal(batch1.Status, batch2.Status)
}

func assertListBatchesReply(expected, actual []mcom.BatchInfo, assert *assert.Assertions) bool {
	if !assert.Equal(len(expected), len(actual)) {
		return false
	}
	for i, reply := range actual {
		switch {
		case !assert.Equal(expected[i].WorkOrder, strings.TrimRight(reply.WorkOrder, " ")):
			return false
		case !assert.Equal(expected[i].Note, reply.Note):
			return false
		case !assert.Equal(expected[i].Number, reply.Number):
			return false
		case !assert.Equal(expected[i].Records, reply.Records):
			return false
		case !assert.Equal(expected[i].Status, reply.Status):
			return false
		}
	}
	return true
}

func convertBatchInfoToBatch(rep mcom.BatchInfo) models.Batch {
	recordsID := make([]string, len(rep.Records))
	for i, rec := range rep.Records {
		recordsID[i] = rec.ID
	}
	return models.Batch{
		WorkOrder: rep.WorkOrder,
		Number:    rep.Number,
		Status:    workorder.BatchStatus(rep.Status),
		RecordsID: recordsID,
		Note:      rep.Note,
		UpdatedAt: rep.UpdatedAt,
		UpdatedBy: rep.UpdatedBy,
	}
}

// #endregion Private_functions_for_batch_test
