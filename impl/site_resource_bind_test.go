package impl

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/bindtype"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func TestDataManager_MaterialResourceBind(t *testing.T) {
	// #region setup
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	assert := assert.New(t)
	assertion := bindOperationAssertion{
		dm:  dm,
		db:  db,
		ctx: ctx,
		t:   t,
	}

	assert.NoError(clearMaterialResourceBindTestDB(db))

	// create test resource data.
	_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{
			{
				Type:       "A",
				ID:         "A",
				Status:     0,
				Quantity:   decimal.NewFromInt(100),
				Unit:       "A",
				LotNumber:  "A",
				ExpiryTime: types.TimeNano(0).Time(),
				ResourceID: "A",
			},
			{
				Type:       "B",
				ID:         "B",
				Status:     0,
				Quantity:   decimal.NewFromInt(150),
				Unit:       "B",
				LotNumber:  "B",
				ExpiryTime: types.TimeNano(0).Time(),
				ResourceID: "B",
			},
		},
	}, mcom.WithStockIn(mcom.Warehouse{
		ID:       "w",
		Location: "w",
	}))
	assert.NoError(err)
	// #endregion setup
	// #region bad cases
	{ // insufficient request.
		assert.ErrorIs(
			dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{}),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing station or only one shared site can be targeted",
			},
		)

		assert.NoError(dm.(*DataManager).createTestSiteAndStation(ctx, createTestSiteAndStationRequest{
			stationID: "InsufficientReq",
			siteName:  "InsufficientReq",
			siteIndex: 1,
			Type:      sites.Type_SLOT,
		}))
		assert.ErrorIs(
			dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
				Station: "InsufficientReq",
				Details: []mcom.MaterialBindRequestDetail{{}},
			}), mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing bind type",
			},
		)

		assert.ErrorIs(
			dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
				Station: "InsufficientReq",
				Details: []mcom.MaterialBindRequestDetail{{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				}},
			}), mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing SiteID",
			},
		)
	}
	{ // Station not found.
		assert.ErrorIs(
			dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
				Station: "NotFound",
				Details: []mcom.MaterialBindRequestDetail{},
			}), mcomErr.Error{
				Code:    mcomErr.Code_STATION_NOT_FOUND,
				Details: "station not found, id: NotFound",
			},
		)
	}
	{ // invalid sub type.
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "invalidSubType",
			DepartmentOID: "invalidSubType" + "OID",
			Sites: []mcom.SiteInformation{{
				Name:    "invalidSubType",
				Index:   2,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_TOOL,
			}},
			State: stations.State_IDLE,
		}))
		assert.ErrorIs(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "invalidSubType",
			Details: []mcom.MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR,
				Site: models.SiteID{
					Name:  "invalidSubType",
					Index: 2,
				},
				Resources: []mcom.BindMaterialResource{},
			}},
		}), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "invalid site sub type, station: invalidSubType name: invalidSubType index: 2",
		})
	}
	{ // BadRequest.
		// redundant bind resources for slot clear.
		assert.NoError(dm.(*DataManager).createTestSiteAndStation(ctx, createTestSiteAndStationRequest{
			stationID: "BadReqSlot",
			siteName:  "BadReqSlot",
			siteIndex: 1,
			Type:      sites.Type_SLOT,
		}))
		assert.ErrorIs(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "BadReqSlot",
			Details: []mcom.MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR,
				Site: models.SiteID{
					Name:  "BadReqSlot",
					Index: 1,
				},
				Resources: []mcom.BindMaterialResource{{}},
			}},
		}), mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "unspecified quantity only applicable to slot bind or queue bind"})

		// missing option.
		assert.NoError(dm.(*DataManager).createTestSiteAndStation(ctx, createTestSiteAndStationRequest{
			stationID: "BadReqColQ",
			siteName:  "BadReqColQ",
			siteIndex: 1,
			Type:      sites.Type_COLQUEUE,
		}))
		assert.ErrorIs(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "BadReqColQ",
			Details: []mcom.MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD,
				Site: models.SiteID{
					Name:  "BadReqColQ",
					Index: 1,
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
				}},
			}},
		}), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "invalid option",
		})

		// site type mismatch.
		assert.ErrorIs(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "BadReqColQ",
			Details: []mcom.MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_CONTAINER_BIND,
				Site: models.SiteID{
					Name:  "BadReqColQ",
					Index: 1,
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
				}},
			}},
		}), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "site type mismatch",
		})
	}
	// #endregion bad cases

	// #region ordinary cases
	{ // slot bind.
		assertion.stationID = "Station1"
		assert.NoError(dm.(*DataManager).createTestSiteAndStation(ctx, createTestSiteAndStationRequest{
			stationID: "Station1",
			siteName:  "Site",
			siteIndex: 1,
			Type:      sites.Type_SLOT,
		}))
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "Station1",
			Details: []mcom.MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{
					Name:  "Site",
					Index: 1,
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
					Warehouse: mcom.Warehouse{
						ID:       "w",
						Location: "w",
					},
					Status:     resources.MaterialStatus_AVAILABLE,
					ExpiryTime: 0,
				}},
			}},
		}))

		expectedSiteContents := models.SiteContents{
			Name:  "Site",
			Index: 1,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "A",
							Grade: "A",
						},
						Quantity:    types.Decimal.NewFromInt32(10),
						ResourceID:  "A",
						ProductType: "A",
						Status:      resources.MaterialStatus_AVAILABLE,
						ExpiryTime:  0,
					},
				},
			},
		}
		assertion.assertSiteContents(expectedSiteContents)

		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("90.000000"), decimal.RequireFromString("150.000000"))
	}
	{ // slot bind.
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "Station1",
			Details: []mcom.MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{
					Name:  "Site",
					Index: 1,
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    types.Decimal.NewFromInt32(20),
					ResourceID:  "B",
					ProductType: "B",
					Warehouse: mcom.Warehouse{
						ID:       "w",
						Location: "w",
					},
					ExpiryTime: 0,
				}},
			}},
		}))

		expectedSiteContents := models.SiteContents{
			Name:  "Site",
			Index: 1,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "B",
							Grade: "B",
						},
						Quantity:    types.Decimal.NewFromInt32(20),
						ResourceID:  "B",
						ProductType: "B",
						ExpiryTime:  0,
					},
				},
			},
		}
		assertion.assertSiteContents(expectedSiteContents)

		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("100.000000"), decimal.RequireFromString("130.000000"))
	}
	// #endregion ordinary cases

	// #region special cases
	{ // bind with incomplete resource (resource C not exists)
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "Station1",
			Details: []mcom.MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{
					Name:  "Site",
					Index: 1,
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "C",
						Grade: "C",
					},
					Quantity:    types.Decimal.NewFromInt32(30),
					ResourceID:  "C",
					ProductType: "C",
					ExpiryTime:  0,
				}},
			}},
		}))

		expectedSiteContents := models.SiteContents{
			Name:  "Site",
			Index: 1,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "C",
							Grade: "C",
						},
						Quantity:    types.Decimal.NewFromInt32(30),
						ResourceID:  "C",
						ProductType: "C",
						ExpiryTime:  0,
					},
				},
			},
		}
		assertion.assertSiteContents(expectedSiteContents)

		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("100.000000"), decimal.RequireFromString("150.000000"))
	}
	{ // bind with incomplete resource (resource C and D not exists)
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "Station1",
			Details: []mcom.MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{
					Name:  "Site",
					Index: 1,
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "D",
						Grade: "D",
					},
					Quantity:    types.Decimal.NewFromInt32(30),
					ResourceID:  "D",
					ProductType: "D",
					ExpiryTime:  0,
				}},
			}},
		}))

		expectedSiteContents := models.SiteContents{
			Name:  "Site",
			Index: 1,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "D",
							Grade: "D",
						},
						Quantity:    types.Decimal.NewFromInt32(30),
						ResourceID:  "D",
						ProductType: "D",
						ExpiryTime:  0,
					},
				},
			},
		}
		assertion.assertSiteContents(expectedSiteContents)
		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("100.000000"), decimal.RequireFromString("150.000000"))
	}
	{ // slot bind with unspecified quantity
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "Station1",
			Details: []mcom.MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{
					Name:  "Site",
					Index: 1,
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    nil,
					ResourceID:  "B",
					ProductType: "B",
					Warehouse: mcom.Warehouse{
						ID:       "w",
						Location: "w",
					},
					ExpiryTime: 0,
				}},
			}},
		}))

		expectedSiteContents := models.SiteContents{
			Name:  "Site",
			Index: 1,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "B",
							Grade: "B",
						},
						Quantity:    nil,
						ResourceID:  "B",
						ProductType: "B",
						ExpiryTime:  0,
					},
				},
			},
		}
		assertion.assertSiteContents(expectedSiteContents)

		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("100.000000"), decimal.RequireFromString("150.000000"))
	}
	{ // test sub type checking
		assertion.stationID = "testSubTypeStation"
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "testSubTypeStation",
			DepartmentOID: "departmentOID",
			Sites: []mcom.SiteInformation{
				{
					Name:       "materialSlot",
					Index:      0,
					Type:       sites.Type_SLOT,
					SubType:    sites.SubType_MATERIAL,
					Limitation: []string{},
				},
				{
					Name:       "toolSlot",
					Index:      0,
					Type:       sites.Type_SLOT,
					SubType:    sites.SubType_TOOL,
					Limitation: []string{},
				},
			},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
		}))

		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "testSubTypeStation",
			Details: []mcom.MaterialBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "materialSlot",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{
						{
							Material: models.Material{
								ID:    "B",
								Grade: "B",
							},
							Quantity:    types.Decimal.NewFromInt32(5),
							ResourceID:  "B",
							ProductType: "B",
							Warehouse: mcom.Warehouse{
								ID:       "w",
								Location: "w",
							},
							ExpiryTime: 0,
						},
					},
					Option: models.BindOption{},
				},
			},
		}))

		expectedMaterialSiteContents := models.SiteContents{
			Name:  "materialSlot",
			Index: 0,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "B",
							Grade: "B",
						},
						Quantity:    types.Decimal.NewFromInt32(5),
						ResourceID:  "B",
						ProductType: "B",
						ExpiryTime:  0,
					},
				},
			},
		}

		expectedToolSiteContents := models.SiteContents{
			Name:    "toolSlot",
			Index:   0,
			Station: "testSubTypeStation",
			Content: models.SiteContent{
				Slot: &models.Slot{},
			},
		}

		assertion.assertSiteContents(expectedMaterialSiteContents)
		assertion.assertSiteContents(expectedToolSiteContents)
		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("100.000000"), decimal.RequireFromString("145.000000"))
	}
	{ // queue bind with unspecified quantity
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "Queue",
			DepartmentOID: "Queue",
			Sites: []mcom.SiteInformation{
				{
					Name:       "queue",
					Index:      0,
					Type:       sites.Type_QUEUE,
					SubType:    sites.SubType_MATERIAL,
					Limitation: []string{},
				},
			},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
		}))
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "Queue",
			Details: []mcom.MaterialBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_QUEUE_BIND,
					Site: models.SiteID{
						Name:  "queue",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{{
						Material: models.Material{
							ID:    "B",
							Grade: "B",
						},
						Quantity:    nil,
						ResourceID:  "B",
						ProductType: "B",
						Warehouse: mcom.Warehouse{
							ID:       "w",
							Location: "w",
						},
						ExpiryTime: 0,
					},
					},
					Option: models.BindOption{
						Head: true,
					},
				},
			},
		}))
		assertion.stationID = "Queue"

		expectedMaterialSiteContents := models.SiteContents{
			Name:  "queue",
			Index: 0,
			Content: models.SiteContent{
				Queue: &models.Queue{
					{
						Material: &models.MaterialSite{
							Material: models.Material{
								ID:    "B",
								Grade: "B",
							},
							Quantity:    nil,
							ResourceID:  "B",
							ProductType: "B",
							ExpiryTime:  0,
						},
					},
				},
			},
		}

		assertion.assertSiteContents(expectedMaterialSiteContents)
		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("100.000000"), decimal.RequireFromString("145.000000"))
	}
	// #endregion special cases

	assert.NoError(clearMaterialResourceBindTestDB(db))
}

func TestDataManager_MaterialResourceBindV2(t *testing.T) {
	// #region setup
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	assert := assert.New(t)
	assertion := bindOperationAssertion{
		dm:  dm,
		db:  db,
		ctx: ctx,
		t:   t,
	}

	assert.NoError(clearMaterialResourceBindTestDB(db))

	// create test resource data
	_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{
			{
				Type:       "A",
				ID:         "A",
				Status:     0,
				Quantity:   decimal.NewFromInt(100),
				Unit:       "A",
				LotNumber:  "A",
				ExpiryTime: types.TimeNano(0).Time(),
				ResourceID: "A",
			},
			{
				Type:       "B",
				ID:         "B",
				Status:     0,
				Quantity:   decimal.NewFromInt(150),
				Unit:       "B",
				LotNumber:  "B",
				ExpiryTime: types.TimeNano(0).Time(),
				ResourceID: "B",
			},
		},
	}, mcom.WithStockIn(mcom.Warehouse{
		ID:       "w",
		Location: "w",
	}))
	assert.NoError(err)
	// #endregion setup
	// #region bad cases
	{ // invalid sub type
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "invalidSubType",
			DepartmentOID: "invalidSubType" + "OID",
			Sites: []mcom.SiteInformation{{
				Name:    "invalidSubType",
				Index:   2,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_TOOL,
			}},
			State: stations.State_IDLE,
		}))
		assert.ErrorIs(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "invalidSubType",
						Index: 2,
					},
					Station: "invalidSubType",
				},
				Resources: []mcom.BindMaterialResource{},
			}},
		}), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "invalid site sub type, station: invalidSubType name: invalidSubType index: 2",
		})
	}
	{ // BadRequest
		// redundant bind resources for slot clear
		assert.NoError(dm.(*DataManager).createTestSiteAndStation(ctx, createTestSiteAndStationRequest{
			stationID: "BadReqSlot",
			siteName:  "BadReqSlot",
			siteIndex: 1,
			Type:      sites.Type_SLOT,
		}))
		assert.ErrorIs(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "BadReqSlot",
						Index: 1,
					},
					Station: "BadReqSlot",
				},
				Resources: []mcom.BindMaterialResource{{}},
			}},
		}), mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "unspecified quantity only applicable to slot bind or queue bind"})

		// missing option
		assert.NoError(dm.(*DataManager).createTestSiteAndStation(ctx, createTestSiteAndStationRequest{
			stationID: "BadReqColQ",
			siteName:  "BadReqColQ",
			siteIndex: 1,
			Type:      sites.Type_COLQUEUE,
		}))
		assert.ErrorIs(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "BadReqColQ",
						Index: 1,
					},
					Station: "BadReqColQ",
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
				}},
			}},
		}), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "invalid option",
		})

		// site type mismatch.
		assert.ErrorIs(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_CONTAINER_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "BadReqColQ",
						Index: 1,
					},
					Station: "BadReqColQ",
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
				}},
			}},
		}), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "site type mismatch",
		})
	}
	// #endregion bad cases

	// #region ordinary cases
	{ // slot bind.
		assertion.stationID = "Station1"
		assert.NoError(dm.(*DataManager).createTestSiteAndStation(ctx, createTestSiteAndStationRequest{
			stationID: "Station1",
			siteName:  "Site",
			siteIndex: 1,
			Type:      sites.Type_SLOT,
		}))
		assert.NoError(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "Site",
						Index: 1,
					},
					Station: "Station1",
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
					Warehouse: mcom.Warehouse{
						ID:       "w",
						Location: "w",
					},
					Status:     resources.MaterialStatus_AVAILABLE,
					ExpiryTime: 0,
				}},
			}},
		}))

		expectedSiteContents := models.SiteContents{
			Name:  "Site",
			Index: 1,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "A",
							Grade: "A",
						},
						Quantity:    types.Decimal.NewFromInt32(10),
						ResourceID:  "A",
						ProductType: "A",
						Status:      resources.MaterialStatus_AVAILABLE,
						ExpiryTime:  0,
					},
				},
			},
		}
		assertion.assertSiteContents(expectedSiteContents)

		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("90.000000"), decimal.RequireFromString("150.000000"))
	}
	{ // slot bind
		assert.NoError(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "Site",
						Index: 1,
					},
					Station: "Station1",
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    types.Decimal.NewFromInt32(20),
					ResourceID:  "B",
					ProductType: "B",
					Warehouse: mcom.Warehouse{
						ID:       "w",
						Location: "w",
					},
					ExpiryTime: 0,
				}},
			}},
		}))

		expectedSiteContents := models.SiteContents{
			Name:  "Site",
			Index: 1,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "B",
							Grade: "B",
						},
						Quantity:    types.Decimal.NewFromInt32(20),
						ResourceID:  "B",
						ProductType: "B",
						ExpiryTime:  0,
					},
				},
			},
		}
		assertion.assertSiteContents(expectedSiteContents)

		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("100.000000"), decimal.RequireFromString("130.000000"))
	}
	// #endregion ordinary cases

	// #region special cases
	{ // bind with incomplete resource (resource C not exists)
		assert.NoError(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "Site",
						Index: 1,
					},
					Station: "Station1",
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "C",
						Grade: "C",
					},
					Quantity:    types.Decimal.NewFromInt32(30),
					ResourceID:  "C",
					ProductType: "C",
					ExpiryTime:  0,
				}},
			}},
		}))

		expectedSiteContents := models.SiteContents{
			Name:  "Site",
			Index: 1,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "C",
							Grade: "C",
						},
						Quantity:    types.Decimal.NewFromInt32(30),
						ResourceID:  "C",
						ProductType: "C",
						ExpiryTime:  0,
					},
				},
			},
		}
		assertion.assertSiteContents(expectedSiteContents)

		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("100.000000"), decimal.RequireFromString("150.000000"))
	}
	{ // bind with incomplete resource (resource C and D not exists)
		assert.NoError(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "Site",
						Index: 1,
					},
					Station: "Station1",
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "D",
						Grade: "D",
					},
					Quantity:    types.Decimal.NewFromInt32(30),
					ResourceID:  "D",
					ProductType: "D",
					ExpiryTime:  0,
				}},
			}},
		}))

		expectedSiteContents := models.SiteContents{
			Name:  "Site",
			Index: 1,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "D",
							Grade: "D",
						},
						Quantity:    types.Decimal.NewFromInt32(30),
						ResourceID:  "D",
						ProductType: "D",
						ExpiryTime:  0,
					},
				},
			},
		}
		assertion.assertSiteContents(expectedSiteContents)
		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("100.000000"), decimal.RequireFromString("150.000000"))
	}
	{ // slot bind with unspecified quantity
		assert.NoError(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "Site",
						Index: 1,
					},
					Station: "Station1",
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    nil,
					ResourceID:  "B",
					ProductType: "B",
					Warehouse: mcom.Warehouse{
						ID:       "w",
						Location: "w",
					},
					ExpiryTime: 0,
				}},
			}},
		}))

		expectedSiteContents := models.SiteContents{
			Name:  "Site",
			Index: 1,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "B",
							Grade: "B",
						},
						Quantity:    nil,
						ResourceID:  "B",
						ProductType: "B",
						ExpiryTime:  0,
					},
				},
			},
		}
		assertion.assertSiteContents(expectedSiteContents)

		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("100.000000"), decimal.RequireFromString("150.000000"))
	}
	{ // test sub type checking.
		assertion.stationID = "testSubTypeStation"
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "testSubTypeStation",
			DepartmentOID: "departmentOID",
			Sites: []mcom.SiteInformation{
				{
					Name:       "materialSlot",
					Index:      0,
					Type:       sites.Type_SLOT,
					SubType:    sites.SubType_MATERIAL,
					Limitation: []string{},
				},
				{
					Name:       "toolSlot",
					Index:      0,
					Type:       sites.Type_SLOT,
					SubType:    sites.SubType_TOOL,
					Limitation: []string{},
				},
			},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
		}))

		assert.NoError(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "materialSlot",
							Index: 0,
						},
						Station: "testSubTypeStation",
					},
					Resources: []mcom.BindMaterialResource{
						{
							Material: models.Material{
								ID:    "B",
								Grade: "B",
							},
							Quantity:    types.Decimal.NewFromInt32(5),
							ResourceID:  "B",
							ProductType: "B",
							Warehouse: mcom.Warehouse{
								ID:       "w",
								Location: "w",
							},
							ExpiryTime: 0,
						},
					},
					Option: models.BindOption{},
				},
			},
		}))

		expectedMaterialSiteContents := models.SiteContents{
			Name:  "materialSlot",
			Index: 0,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "B",
							Grade: "B",
						},
						Quantity:    types.Decimal.NewFromInt32(5),
						ResourceID:  "B",
						ProductType: "B",
						ExpiryTime:  0,
					},
				},
			},
		}

		expectedToolSiteContents := models.SiteContents{
			Name:    "toolSlot",
			Index:   0,
			Station: "testSubTypeStation",
			Content: models.SiteContent{
				Slot: &models.Slot{},
			},
		}

		assertion.assertSiteContents(expectedMaterialSiteContents)
		assertion.assertSiteContents(expectedToolSiteContents)
		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("100.000000"), decimal.RequireFromString("145.000000"))
	}
	{ // queue bind with unspecified quantity.
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "Queue",
			DepartmentOID: "Queue",
			Sites: []mcom.SiteInformation{
				{
					Name:       "queue",
					Index:      0,
					Type:       sites.Type_QUEUE,
					SubType:    sites.SubType_MATERIAL,
					Limitation: []string{},
				},
			},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
		}))
		assert.NoError(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_QUEUE_BIND,
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "queue",
							Index: 0,
						},
						Station: "Queue",
					},
					Resources: []mcom.BindMaterialResource{{
						Material: models.Material{
							ID:    "B",
							Grade: "B",
						},
						Quantity:    nil,
						ResourceID:  "B",
						ProductType: "B",
						Warehouse: mcom.Warehouse{
							ID:       "w",
							Location: "w",
						},
						ExpiryTime: 0,
					},
					},
					Option: models.BindOption{
						Head: true,
					},
				},
			},
		}))
		assertion.stationID = "Queue"

		expectedMaterialSiteContents := models.SiteContents{
			Name:  "queue",
			Index: 0,
			Content: models.SiteContent{
				Queue: &models.Queue{
					{
						Material: &models.MaterialSite{
							Material: models.Material{
								ID:    "B",
								Grade: "B",
							},
							Quantity:    nil,
							ResourceID:  "B",
							ProductType: "B",
							ExpiryTime:  0,
						},
					},
				},
			},
		}

		assertion.assertSiteContents(expectedMaterialSiteContents)
		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("100.000000"), decimal.RequireFromString("145.000000"))
	}
	{ // bind with empty site name
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "emptySiteName",
			DepartmentOID: "emptySiteNameOID",
			Sites: []mcom.SiteInformation{{
				Station:    "emptySiteName",
				Name:       "",
				Index:      0,
				Type:       sites.Type_SLOT,
				SubType:    sites.SubType_MATERIAL,
				Limitation: []string{},
			}},
			State: stations.State_IDLE,
		}))

		assert.NoError(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "",
						Index: 0,
					},
					Station: "emptySiteName",
				},
				Resources: []mcom.BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(15),
					ResourceID:  "A",
					ProductType: "A",
					Status:      resources.MaterialStatus_AVAILABLE,
					ExpiryTime:  0,
					Warehouse: mcom.Warehouse{
						ID:       "w",
						Location: "w",
					},
				}},
				Option: models.BindOption{},
			}},
		}))

		assertion.stationID = "emptySiteName"

		expectedMaterialSiteContents := models.SiteContents{
			Name:  "",
			Index: 0,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "A",
							Grade: "A",
						},
						Quantity:    types.Decimal.NewFromInt32(15),
						ResourceID:  "A",
						ProductType: "A",
						Status:      resources.MaterialStatus_AVAILABLE,
						ExpiryTime:  0,
					},
				},
			},
		}

		assertion.assertSiteContents(expectedMaterialSiteContents)
		assertion.assertResourcesAndWarehouseStock(decimal.RequireFromString("85.000000"), decimal.RequireFromString("145.000000"))
	}
	// #endregion special cases

	assert.NoError(clearMaterialResourceBindTestDB(db))
}

func (dm *DataManager) createTestSiteAndStation(ctx context.Context, req createTestSiteAndStationRequest) error {
	if err := dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            req.stationID,
		DepartmentOID: req.stationID + "OID",
		Sites: []mcom.SiteInformation{{
			Name:    req.siteName,
			Index:   int(req.siteIndex),
			Type:    req.Type,
			SubType: sites.SubType_MATERIAL,
		}},
		State: stations.State_IDLE,
	}); err != nil {
		return err
	}

	return nil
}

type createTestSiteAndStationRequest struct {
	stationID string
	siteName  string
	siteIndex int16
	Type      sites.Type
}

type bindOperationAssertion struct {
	dm        mcom.DataManager
	db        *gorm.DB
	ctx       context.Context
	t         *testing.T
	stationID string
}

// assertResourcesAndWarehouseStock checks the quantity of product A and B in the corresponding warehouse location.
// Notice that this method may go run while products A and B are stored in more than one location.
func (aso bindOperationAssertion) assertResourcesAndWarehouseStock(expectedWarehouseStockA, expectedWarehouseStockB decimal.Decimal) {
	assert := assert.New(aso.t)
	var latestResources []models.MaterialResource
	assert.NoError(aso.db.Find(&latestResources).Error)
	var actualResourceA, actualResourceB decimal.Decimal
	for _, resource := range latestResources {
		switch resource.ID {
		case "A":
			actualResourceA = resource.Quantity
		case "B":
			actualResourceB = resource.Quantity
		}
	}
	assert.Equal(expectedWarehouseStockA, actualResourceA)
	assert.Equal(expectedWarehouseStockB, actualResourceB)

	var latestWarehouseStocks []models.WarehouseStock
	assert.NoError(aso.db.Find(&latestWarehouseStocks).Error)
	var actualStockA, actualStockB decimal.Decimal
	for _, stock := range latestWarehouseStocks {
		switch stock.ProductID {
		case "A":
			actualStockA = stock.Quantity
		case "B":
			actualStockB = stock.Quantity
		}
	}
	assert.Equal(expectedWarehouseStockA, actualStockA)
	assert.Equal(expectedWarehouseStockB, actualStockB)
}

func (aso bindOperationAssertion) assertSiteContents(expectedSiteContents models.SiteContents) {
	assert := assert.New(aso.t)
	actual, err := aso.dm.GetSite(aso.ctx, mcom.GetSiteRequest{
		StationID: aso.stationID,
		SiteName:  expectedSiteContents.Name,
		SiteIndex: expectedSiteContents.Index,
	})
	assert.NoError(err)
	assert.Equal(expectedSiteContents.Content, actual.Content)
}

func clearMaterialResourceBindTestDB(db *gorm.DB) error {
	return newClearMaster(db,
		&models.MaterialResource{},
		&models.SiteContents{},
		&models.Site{},
		&models.Station{},
		&models.WarehouseStock{},
		&models.BindRecords{}).Clear()
}
