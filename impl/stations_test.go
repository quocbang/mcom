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

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/bindtype"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

const (
	testStationGroupID = "T1234"
	testStationA       = "S01"
	testStationB       = "S02"
	testStationC       = "S03"

	testSiteA = "site_A"
	testSiteB = "site_B"
	testSiteC = "site_C"
	testSiteD = "site_D"
	testSiteE = "site_E"
	testSiteF = "site_F"
	testSiteG = "site_G"
)

func TestDataManager_StationGroup(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	dm, _, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}
	defer dm.Close()

	{ // CreateStationGroup: insufficient request: ID, Stations.
		err := dm.CreateStationGroup(ctx, mcom.StationGroupRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // UpdateStationGroup: insufficient request: ID.
		err := dm.UpdateStationGroup(ctx, mcom.StationGroupRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // DeleteStationGroup: insufficient request: ID.
		err := dm.DeleteStationGroup(ctx, mcom.DeleteStationGroupRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // UpdateStationGroup: station group not found.
		err := dm.UpdateStationGroup(ctx, mcom.StationGroupRequest{
			ID:       testStationGroupID,
			Stations: []string{testStationA, testStationB},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_STATION_GROUP_ID_NOT_FOUND,
		}, err)
	}
	{ // CreateStationGroup: good case.
		err := dm.CreateStationGroup(ctx, mcom.StationGroupRequest{
			ID:       testStationGroupID,
			Stations: []string{testStationA, testStationB},
		})
		assert.NoError(err)
	}
	{ // CreateStationGroup: station group already exists.
		err := dm.CreateStationGroup(ctx, mcom.StationGroupRequest{
			ID:       testStationGroupID,
			Stations: []string{testStationA, testStationB},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_STATION_GROUP_ALREADY_EXISTS,
		}, err)
	}
	{ // UpdateStationGroup: good case.
		err := dm.UpdateStationGroup(ctx, mcom.StationGroupRequest{
			ID:       testStationGroupID,
			Stations: []string{testStationA, testStationB, testStationC},
		})
		assert.NoError(err)
	}
	{ // DeleteStationGroup: good case.
		err := dm.DeleteStationGroup(ctx, mcom.DeleteStationGroupRequest{GroupID: testStationGroupID})
		assert.NoError(err)
	}
}

func clearStationsData(db *gorm.DB) error {
	return newClearMaster(db, &models.Station{}, &models.Site{}, &models.SiteContents{}).Clear()
}

// test sites order in GetStation / ListStations.
// test Default station status (shutdown)
func TestDataManager_StationsOrderAndDefaultStatus(t *testing.T) {
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
	assert.NoError(clearStationsData(db))

	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            "AAA",
		DepartmentOID: "AAA",
		Sites: []mcom.SiteInformation{{
			Name:    "AAA",
			Index:   5,
			Type:    sites.Type_COLLECTION,
			SubType: sites.SubType_MATERIAL,
		}, {
			Name:    "AAA",
			Index:   3,
			Type:    sites.Type_COLLECTION,
			SubType: sites.SubType_MATERIAL,
		}, {
			Name:    "AAA",
			Index:   1,
			Type:    sites.Type_COLLECTION,
			SubType: sites.SubType_MATERIAL,
		}, {
			Name:    "BBB",
			Index:   5,
			Type:    sites.Type_COLLECTION,
			SubType: sites.SubType_MATERIAL,
		}, {
			Name:    "BBB",
			Index:   1,
			Type:    sites.Type_COLLECTION,
			SubType: sites.SubType_MATERIAL,
		}},
		State:       0,
		Information: mcom.StationInformation{},
	}))

	rep, err := dm.GetStation(ctx, mcom.GetStationRequest{
		ID: "AAA",
	})
	assert.NoError(err)
	assert.Equal(mcom.GetStationReply{
		ID:                 "AAA",
		AdminDepartmentOID: "AAA",
		Sites: []mcom.ListStationSite{{
			Information: mcom.ListStationSitesInformation{
				UniqueSite: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "AAA",
						Index: 1,
					},
					Station: "AAA",
				},
				Type:    sites.Type_COLLECTION,
				SubType: sites.SubType_MATERIAL,
			},
			Content: models.SiteContent{
				Collection: &models.Collection{},
			},
		}, {
			Information: mcom.ListStationSitesInformation{
				UniqueSite: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "AAA",
						Index: 3,
					},
					Station: "AAA",
				},
				Type:    sites.Type_COLLECTION,
				SubType: sites.SubType_MATERIAL,
			},
			Content: models.SiteContent{
				Collection: &models.Collection{},
			},
		}, {
			Information: mcom.ListStationSitesInformation{
				UniqueSite: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "AAA",
						Index: 5,
					},
					Station: "AAA",
				},
				Type:    sites.Type_COLLECTION,
				SubType: sites.SubType_MATERIAL,
			},
			Content: models.SiteContent{
				Collection: &models.Collection{},
			},
		}, {
			Information: mcom.ListStationSitesInformation{
				UniqueSite: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "BBB",
						Index: 1,
					},
					Station: "AAA",
				},
				Type:    sites.Type_COLLECTION,
				SubType: sites.SubType_MATERIAL,
			},
			Content: models.SiteContent{
				Collection: &models.Collection{},
			},
		}, {
			Information: mcom.ListStationSitesInformation{
				UniqueSite: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "BBB",
						Index: 5,
					},
					Station: "AAA",
				},
				Type:    sites.Type_COLLECTION,
				SubType: sites.SubType_MATERIAL,
			},
			Content: models.SiteContent{
				Collection: &models.Collection{},
			},
		}},
		State:       stations.State_SHUTDOWN,
		Information: mcom.StationInformation{},
		UpdatedBy:   rep.UpdatedBy,
		UpdatedAt:   rep.UpdatedAt,
		InsertedBy:  rep.InsertedBy,
		InsertedAt:  rep.InsertedAt,
	}, rep)

	assert.NoError(clearStationsData(db))
}

func TestDataManager_ListStations(t *testing.T) {
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
	assert.NoError(clearStationsData(db))

	// 統一時間
	timeNow := time.Now().Local()
	monkey.Patch(time.Now, func() time.Time {
		return timeNow
	})
	defer monkey.UnpatchAll()

	{ // insufficient request.
		_, err := dm.ListStations(ctx, mcom.ListStationsRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // good case: empty content.
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            testStationB,
			DepartmentOID: testDepartmentB,
			Sites: []mcom.SiteInformation{{
				Name:    testSiteA,
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_OPERATOR,
			}, {
				Name:    testSiteB,
				Index:   0,
				Type:    sites.Type_CONTAINER,
				SubType: sites.SubType_MATERIAL,
			}, {
				Name:    testSiteC,
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_TOOL,
			}, {
				Name:    testSiteD,
				Index:   0,
				Type:    sites.Type_COLLECTION,
				SubType: sites.SubType_MATERIAL,
			}, {
				Name:    testSiteE,
				Index:   0,
				Type:    sites.Type_QUEUE,
				SubType: sites.SubType_MATERIAL,
			}, {
				Name:    testSiteF,
				Index:   0,
				Type:    sites.Type_COLQUEUE,
				SubType: sites.SubType_MATERIAL,
			}, {
				Name:    testSiteG,
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_MATERIAL,
			}},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
		}))

		results, err := dm.ListStations(ctx, mcom.ListStationsRequest{DepartmentOID: testDepartmentB})
		assert.NoError(err)
		assert.Equal(mcom.ListStationsReply{
			Stations: []mcom.Station{{
				ID:                 testStationB,
				AdminDepartmentOID: testDepartmentB,
				State:              stations.State_IDLE,
				Sites: []mcom.ListStationSite{{
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID: models.SiteID{
								Name:  testSiteA,
								Index: 0,
							},
							Station: testStationB,
						},
						Type:    sites.Type_SLOT,
						SubType: sites.SubType_OPERATOR,
					},
					Content: models.SiteContent{
						Slot: &models.Slot{},
					},
				}, {
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID: models.SiteID{
								Name:  testSiteB,
								Index: 0,
							},
							Station: testStationB,
						},
						Type:    sites.Type_CONTAINER,
						SubType: sites.SubType_MATERIAL,
					},
					Content: models.SiteContent{
						Container: &models.Container{},
					},
				}, {
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID: models.SiteID{
								Name:  testSiteC,
								Index: 0,
							},
							Station: testStationB,
						},
						Type:    sites.Type_SLOT,
						SubType: sites.SubType_TOOL,
					},
					Content: models.SiteContent{
						Slot: &models.Slot{},
					},
				}, {
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID: models.SiteID{
								Name:  testSiteD,
								Index: 0,
							},
							Station: testStationB,
						},
						Type:    sites.Type_COLLECTION,
						SubType: sites.SubType_MATERIAL,
					},
					Content: models.SiteContent{
						Collection: &models.Collection{},
					},
				}, {
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID: models.SiteID{
								Name:  testSiteE,
								Index: 0,
							},
							Station: testStationB,
						},
						Type:    sites.Type_QUEUE,
						SubType: sites.SubType_MATERIAL,
					},
					Content: models.SiteContent{
						Queue: &models.Queue{},
					},
				}, {
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID: models.SiteID{
								Name:  testSiteF,
								Index: 0,
							},
							Station: testStationB,
						},
						Type:    sites.Type_COLQUEUE,
						SubType: sites.SubType_MATERIAL,
					},
					Content: models.SiteContent{
						Colqueue: &models.Colqueue{},
					},
				}, {
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID: models.SiteID{
								Name:  testSiteG,
								Index: 0,
							},
							Station: testStationB,
						},
						Type:    sites.Type_SLOT,
						SubType: sites.SubType_MATERIAL,
					},
					Content: models.SiteContent{
						Slot: &models.Slot{},
					},
				}},
				UpdatedAt:  timeNow,
				InsertedAt: timeNow,
			}},
		}, results)
	}
	{ // good case.
		assert.NoError(db.Updates(&models.SiteContents{
			Name:  testSiteA,
			Index: 0,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Operator: &models.OperatorSite{
						EmployeeID: testUser,
						Group:      0,
						WorkDate:   timeNow,
					},
				},
			},
			UpdatedBy: testUser,
		}).Error)

		assert.NoError(db.Updates(&models.SiteContents{
			Name:  testSiteB,
			Index: 0,
			Content: models.SiteContent{
				Container: &models.Container{
					{
						Material: &models.MaterialSite{
							Material: models.Material{
								ID:    testResourceMaterialID,
								Grade: testResourceMaterialGrade,
							},
							Quantity:   types.Decimal.NewFromFloat32(123.456),
							ResourceID: testResourceID,
						},
					},
				},
			},
			UpdatedBy: testUser,
		}).Error)

		results, err := dm.ListStations(ctx, mcom.ListStationsRequest{DepartmentOID: testDepartmentB})
		assert.NoError(err)

		// We don't care time zone so we attempt to make them in the same time zone.
		// Also, we know the 0-index is our target, we only change the 0-index time zone.
		results.Stations[0].Sites[0].Content.Slot.Operator.WorkDate = results.Stations[0].Sites[0].Content.Slot.Operator.WorkDate.Local()

		assert.Equal(mcom.ListStationsReply{
			Stations: []mcom.Station{
				{
					ID:                 testStationB,
					AdminDepartmentOID: testDepartmentB,
					State:              stations.State_IDLE,
					Sites: []mcom.ListStationSite{{
						Information: mcom.ListStationSitesInformation{
							UniqueSite: models.UniqueSite{
								SiteID: models.SiteID{
									Name:  testSiteA,
									Index: 0,
								},
								Station: testStationB,
							},
							Type:    sites.Type_SLOT,
							SubType: sites.SubType_OPERATOR,
						},
						Content: models.SiteContent{
							Slot: &models.Slot{
								Operator: &models.OperatorSite{
									EmployeeID: testUser,
									Group:      0,
									WorkDate:   timeNow,
								},
							},
						},
					}, {
						Information: mcom.ListStationSitesInformation{
							UniqueSite: models.UniqueSite{
								SiteID: models.SiteID{
									Name:  testSiteB,
									Index: 0,
								},
								Station: testStationB,
							},
							Type:    sites.Type_CONTAINER,
							SubType: sites.SubType_MATERIAL,
						},
						Content: models.SiteContent{
							Container: &models.Container{
								{
									Material: &models.MaterialSite{
										Material: models.Material{
											ID:    testResourceMaterialID,
											Grade: testResourceMaterialGrade,
										},
										Quantity:   types.Decimal.NewFromFloat32(123.456),
										ResourceID: testResourceID,
									},
								},
							},
						},
					}, {
						Information: mcom.ListStationSitesInformation{
							UniqueSite: models.UniqueSite{
								SiteID: models.SiteID{
									Name:  testSiteC,
									Index: 0,
								},
								Station: testStationB,
							},
							Type:    sites.Type_SLOT,
							SubType: sites.SubType_TOOL,
						},
						Content: models.SiteContent{
							Slot: &models.Slot{},
						},
					}, {
						Information: mcom.ListStationSitesInformation{
							UniqueSite: models.UniqueSite{
								SiteID: models.SiteID{
									Name:  testSiteD,
									Index: 0,
								},
								Station: testStationB,
							},
							Type:    sites.Type_COLLECTION,
							SubType: sites.SubType_MATERIAL,
						},
						Content: models.SiteContent{
							Collection: &models.Collection{},
						},
					}, {
						Information: mcom.ListStationSitesInformation{
							UniqueSite: models.UniqueSite{
								SiteID: models.SiteID{
									Name:  testSiteE,
									Index: 0,
								},
								Station: testStationB,
							},
							Type:    sites.Type_QUEUE,
							SubType: sites.SubType_MATERIAL,
						},
						Content: models.SiteContent{
							Queue: &models.Queue{},
						},
					}, {
						Information: mcom.ListStationSitesInformation{
							UniqueSite: models.UniqueSite{
								SiteID: models.SiteID{
									Name:  testSiteF,
									Index: 0,
								},
								Station: testStationB,
							},
							Type:    sites.Type_COLQUEUE,
							SubType: sites.SubType_MATERIAL,
						},
						Content: models.SiteContent{
							Colqueue: &models.Colqueue{},
						},
					}, {
						Information: mcom.ListStationSitesInformation{
							UniqueSite: models.UniqueSite{
								SiteID: models.SiteID{
									Name:  testSiteG,
									Index: 0,
								},
								Station: testStationB,
							},
							Type:    sites.Type_SLOT,
							SubType: sites.SubType_MATERIAL,
						},
						Content: models.SiteContent{
							Slot: &models.Slot{},
						},
					}},
					UpdatedBy:  "",
					UpdatedAt:  timeNow,
					InsertedBy: "",
					InsertedAt: timeNow,
				},
			},
		}, results)
	}

	assert.NoError(clearStationsData(db))

	{ // test: pagination and order.
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "B",
			DepartmentOID: "DEP",
			State:         stations.State_IDLE,
		}))
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "A",
			DepartmentOID: "DEP",
			State:         stations.State_IDLE,
		}))
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "D",
			DepartmentOID: "DEP",
			State:         stations.State_IDLE,
		}))
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "C",
			DepartmentOID: "DEP",
			State:         stations.State_IDLE,
		}))

		actual, err := dm.ListStations(ctx, mcom.ListStationsRequest{
			DepartmentOID: "DEP",
		}.WithPagination(mcom.PaginationRequest{
			PageCount:      2,
			ObjectsPerPage: 2,
		}).WithOrder(mcom.Order{
			Name:       "id",
			Descending: false,
		}))
		assert.NoError(err)

		expected := mcom.ListStationsReply{
			Stations: []mcom.Station{
				{
					ID:                 "C",
					AdminDepartmentOID: "DEP",
					Sites:              []mcom.ListStationSite{},
					State:              stations.State_IDLE,
					UpdatedAt:          actual.Stations[0].UpdatedAt,
					InsertedAt:         actual.Stations[0].InsertedAt,
				},
				{
					ID:                 "D",
					AdminDepartmentOID: "DEP",
					Sites:              []mcom.ListStationSite{},
					State:              stations.State_IDLE,
					UpdatedAt:          actual.Stations[1].UpdatedAt,
					InsertedAt:         actual.Stations[1].InsertedAt,
				},
			},
			PaginationReply: mcom.PaginationReply{
				AmountOfData: 4,
			},
		}

		assert.Equal(expected, actual)
	}
}

func TestDataManager_GetStation(t *testing.T) {
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
	cm := newClearMaster(db, &models.Station{}, &models.Site{}, &models.SiteContents{}, &models.MaterialResource{}, &models.WarehouseStock{})
	assert.NoError(cm.Clear())
	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            "AAA",
		DepartmentOID: "AAA",
		State:         stations.State_IDLE,
	}))
	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            "BBB",
		DepartmentOID: "BBB",
		State:         stations.State_IDLE,
	}))
	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            "CCC",
		DepartmentOID: "AAA",
		State:         stations.State_IDLE,
	}))
	{ // case insufficient request (should not pass)
		_, err := dm.GetStation(ctx, mcom.GetStationRequest{})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST})
	}
	{ // case not found (should not pass)
		_, err := dm.GetStation(ctx, mcom.GetStationRequest{ID: "FFF"})
		assert.ErrorIs(err, mcomErr.Error{
			Code:    mcomErr.Code_STATION_NOT_FOUND,
			Details: fmt.Sprintf("station not found, id: %v", "FFF"),
		})
	}
	{ // normal case (should pass)
		strVar := "AAA"
		actual, err := dm.GetStation(ctx, mcom.GetStationRequest{ID: strVar})
		assert.NoError(err)
		expected := mcom.GetStationReply{
			ID:                 strVar,
			AdminDepartmentOID: strVar,
			Sites:              []mcom.ListStationSite{},
			State:              stations.State_IDLE,
			Information:        mcom.StationInformation{},
			UpdatedBy:          actual.UpdatedBy,
			UpdatedAt:          actual.UpdatedAt,
			InsertedBy:         actual.InsertedBy,
			InsertedAt:         actual.InsertedAt,
		}
		assert.Equal(expected, actual)

		strVar = "BBB"
		actual, err = dm.GetStation(ctx, mcom.GetStationRequest{ID: strVar})
		assert.NoError(err)
		expected = mcom.GetStationReply{
			ID:                 strVar,
			AdminDepartmentOID: strVar,
			Sites:              []mcom.ListStationSite{},
			State:              stations.State_IDLE,
			Information:        mcom.StationInformation{},
			UpdatedBy:          actual.UpdatedBy,
			UpdatedAt:          actual.UpdatedAt,
			InsertedBy:         actual.InsertedBy,
			InsertedAt:         actual.InsertedAt,
		}
		assert.Equal(expected, actual)

		strVar = "CCC"
		actual, err = dm.GetStation(ctx, mcom.GetStationRequest{ID: strVar})
		assert.NoError(err)
		expected = mcom.GetStationReply{
			ID:                 strVar,
			AdminDepartmentOID: "AAA",
			Sites:              []mcom.ListStationSite{},
			State:              stations.State_IDLE,
			Information:        mcom.StationInformation{},
			UpdatedBy:          actual.UpdatedBy,
			UpdatedAt:          actual.UpdatedAt,
			InsertedBy:         actual.InsertedBy,
			InsertedAt:         actual.InsertedAt,
		}
		assert.Equal(expected, actual)
	}
	{ // deprecated station (should not pass)
		assert.NoError(dm.DeleteStation(ctx, mcom.DeleteStationRequest{StationID: "BBB"}))
		_, err := dm.GetStation(ctx, mcom.GetStationRequest{
			ID: "BBB",
		})
		assert.ErrorIs(err, mcomErr.Error{
			Code:    mcomErr.Code_STATION_NOT_FOUND,
			Details: fmt.Sprintf("station not found, id: %v", "BBB"),
		})
	}
	{ // unspecified quantity of products in slot and queue.
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "station",
			DepartmentOID: "d",
			Sites: []mcom.SiteInformation{
				{
					Name:       "slot",
					Index:      0,
					Type:       sites.Type_SLOT,
					SubType:    sites.SubType_MATERIAL,
					Limitation: []string{},
				},
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

		now := time.Now()
		nowNano := types.ToTimeNano(now)
		_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
			Materials: []mcom.CreateMaterialResourcesRequestDetail{{
				Type:       "A",
				ID:         "A",
				Grade:      "A",
				Status:     resources.MaterialStatus_AVAILABLE,
				Quantity:   decimal.NewFromInt(100),
				ExpiryTime: now,
				ResourceID: "A",
			}}})
		assert.NoError(err)

		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "station",
			Details: []mcom.MaterialBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "slot",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{
						{
							Material: models.Material{
								ID:    "A",
								Grade: "A",
							},
							Quantity:    nil,
							ResourceID:  "A",
							ProductType: "A",
							Status:      resources.MaterialStatus_AVAILABLE,
							ExpiryTime:  nowNano,
							Warehouse:   mcom.Warehouse{},
						},
					},
					Option: models.BindOption{},
				},
				{
					Type: bindtype.BindType_RESOURCE_BINDING_QUEUE_BIND,
					Site: models.SiteID{
						Name:  "queue",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{
						{
							Material: models.Material{
								ID:    "A",
								Grade: "A",
							},
							Quantity:    nil,
							ResourceID:  "A",
							ProductType: "A",
							Status:      resources.MaterialStatus_AVAILABLE,
							ExpiryTime:  nowNano,
							Warehouse:   mcom.Warehouse{},
						},
					},
					Option: models.BindOption{Head: true},
				},
			},
		}))

		actual, err := dm.GetStation(ctx, mcom.GetStationRequest{
			ID: "station",
		})
		assert.NoError(err)
		expected := mcom.GetStationReply{
			AdminDepartmentOID: "d",
			Sites: []mcom.ListStationSite{
				{
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID: models.SiteID{
								Name:  "queue",
								Index: 0,
							},
							Station: "station",
						},
						Type:    sites.Type_QUEUE,
						SubType: sites.SubType_MATERIAL,
					},
					Content: models.SiteContent{
						Queue: &models.Queue{{
							Material: &models.MaterialSite{
								Material: models.Material{
									ID:    "A",
									Grade: "A",
								},
								Quantity:    nil,
								ResourceID:  "A",
								ProductType: "A",
								Status:      resources.MaterialStatus_AVAILABLE,
								ExpiryTime:  nowNano,
							},
						}},
					},
				},
				{
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID: models.SiteID{
								Name:  "slot",
								Index: 0,
							},
							Station: "station",
						},
						Type:    sites.Type_SLOT,
						SubType: sites.SubType_MATERIAL,
					},
					Content: models.SiteContent{
						Slot: &models.Slot{
							Material: &models.MaterialSite{
								Material: models.Material{
									ID:    "A",
									Grade: "A",
								},
								Quantity:    nil,
								ResourceID:  "A",
								ProductType: "A",
								Status:      resources.MaterialStatus_AVAILABLE,
								ExpiryTime:  nowNano,
							},
						},
					},
				},
			},
			Information: mcom.StationInformation{},
			UpdatedBy:   "",
			UpdatedAt:   actual.UpdatedAt,
			InsertedBy:  "",
			InsertedAt:  actual.InsertedAt,
			ID:          "station",
			State:       stations.State_IDLE,
		}
		assert.Equal(expected, actual)
	}

	{ // get station with an empty-name site
		const (
			stationID    = "emptySiteName"
			departmentID = "emptySiteName"
		)
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            stationID,
			DepartmentOID: departmentID,
			Sites: []mcom.SiteInformation{
				{
					Name:       "",
					Index:      0,
					Type:       sites.Type_SLOT,
					SubType:    sites.SubType_MATERIAL,
					Limitation: []string{},
				},
			},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
		}))

		actual, err := dm.GetStation(ctx, mcom.GetStationRequest{
			ID: stationID,
		})
		assert.NoError(err)

		assert.Equal(mcom.GetStationReply{
			ID:                 stationID,
			AdminDepartmentOID: departmentID,
			Sites: []mcom.ListStationSite{
				{
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID:  models.SiteID{},
							Station: stationID,
						},
						Type:    sites.Type_SLOT,
						SubType: sites.SubType_MATERIAL,
					},
					Content: models.SiteContent{
						Slot: &models.Slot{},
					},
				},
			},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
			UpdatedBy:   actual.UpdatedBy,
			UpdatedAt:   actual.UpdatedAt,
			InsertedBy:  actual.InsertedBy,
			InsertedAt:  actual.InsertedAt,
		}, actual)
	}
	{ // unspecified quantity of products not exist in the database.
		assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
			Station: "station",
			Details: []mcom.MaterialBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "slot",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{
						{
							Material: models.Material{
								ID:    "not found",
								Grade: "A",
							},
							Quantity:    nil,
							ResourceID:  "not found",
							ProductType: "not found",
							Status:      resources.MaterialStatus_AVAILABLE,
							ExpiryTime:  0,
						},
					},
					Option: models.BindOption{},
				},
				{
					Type: bindtype.BindType_RESOURCE_BINDING_QUEUE_CLEAR,
					Site: models.SiteID{
						Name:  "queue",
						Index: 0,
					},
					Resources: []mcom.BindMaterialResource{},
					Option:    models.BindOption{},
				},
			},
		}))
		actual, err := dm.GetStation(ctx, mcom.GetStationRequest{
			ID: "station",
		})
		assert.NoError(err)

		assert.Equal(mcom.GetStationReply{
			ID:                 "station",
			AdminDepartmentOID: "d",
			Sites: []mcom.ListStationSite{
				{
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID: models.SiteID{
								Name:  "queue",
								Index: 0,
							},
							Station: "station",
						},
						Type:    sites.Type_QUEUE,
						SubType: sites.SubType_MATERIAL,
					},
					Content: models.SiteContent{
						Queue: &models.Queue{},
					},
				},
				{
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID: models.SiteID{
								Name:  "slot",
								Index: 0,
							},
							Station: "station",
						},
						Type:    sites.Type_SLOT,
						SubType: sites.SubType_MATERIAL,
					},
					Content: models.SiteContent{
						Slot: &models.Slot{
							Material: &models.MaterialSite{
								Material: models.Material{
									ID:    "not found",
									Grade: "A",
								},
								Quantity:    nil,
								ResourceID:  "not found",
								ProductType: "not found",
								Status:      resources.MaterialStatus_AVAILABLE,
								ExpiryTime:  0,
							},
						},
					},
				},
			},
			Information: mcom.StationInformation{},
			UpdatedBy:   "",
			UpdatedAt:   actual.UpdatedAt,
			InsertedBy:  "",
			InsertedAt:  actual.InsertedAt,
			State:       stations.State_IDLE,
		}, actual)
	}
	assert.NoError(cm.Clear())
}

func TestDataManager_CreateStation(t *testing.T) {
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
	assert.NoError(clearStationsData(db))

	// 統一時間
	timeNow := time.Now().Local()
	monkey.Patch(time.Now, func() time.Time {
		return timeNow
	})
	defer monkey.UnpatchAll()

	{ // insufficient request: id.
		err := dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "",
			DepartmentOID: "T1234",
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "the ID or departmentOID is empty",
		}, err)
	}
	{ // insufficient request: department oid.
		err := dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "T1234",
			DepartmentOID: "",
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "the ID or departmentOID is empty",
		}, err)
	}
	{ // create station without sites: good case.
		err := dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            testStationA,
			DepartmentOID: testDepartmentA,
		})
		assert.NoError(err)

		station, err := dm.GetStation(ctx, mcom.GetStationRequest{ID: testStationA})
		assert.NoError(err)
		assert.Equal(mcom.GetStationReply{
			ID:                 testStationA,
			AdminDepartmentOID: testDepartmentA,
			Sites:              []mcom.ListStationSite{},
			State:              stations.State_SHUTDOWN,
			UpdatedAt:          timeNow,
			UpdatedBy:          testUser,
			InsertedAt:         timeNow,
			InsertedBy:         testUser,
		}, station)
	}
	{ // associate to a site.
		existedSiteStation := "associate to a site"

		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            existedSiteStation,
			DepartmentOID: testDepartmentA,
			Sites: []mcom.SiteInformation{
				{
					Name:    testSiteA,
					Index:   0,
					Station: existedSiteStation,
					Type:    sites.Type_SLOT,
					SubType: sites.SubType_MATERIAL,
				},
			},
		}))
		assert.NoError(err)

		createdStation := "new station"
		err = dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            createdStation,
			DepartmentOID: testDepartmentA,
			Sites: []mcom.SiteInformation{{
				Name:    testSiteA,
				Index:   0,
				Station: existedSiteStation,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_MATERIAL,
			}}},
		)
		assert.NoError(err)

		rep, err := dm.GetStation(ctx, mcom.GetStationRequest{
			ID: createdStation,
		})
		assert.NoError(err)

		assert.Equal(mcom.GetStationReply{
			ID:                 createdStation,
			AdminDepartmentOID: testDepartmentA,
			Sites: []mcom.ListStationSite{
				{
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID: models.SiteID{
								Name:  testSiteA,
								Index: 0,
							},
							Station: existedSiteStation,
						},
						Type:    sites.Type_SLOT,
						SubType: sites.SubType_MATERIAL,
					},
					Content: models.SiteContent{
						Slot: &models.Slot{},
					},
				},
			},
			State:      stations.State_SHUTDOWN,
			UpdatedAt:  rep.UpdatedAt,
			UpdatedBy:  testUser,
			InsertedAt: rep.InsertedAt,
			InsertedBy: testUser,
		}, rep)

		assert.NoError(err)
	}
	{ // create station with sites: good case.
		err := dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            testStationB,
			DepartmentOID: testDepartmentA,
			Sites: []mcom.SiteInformation{{
				Name:       testSiteC,
				Index:      0,
				Type:       sites.Type_COLLECTION,
				SubType:    sites.SubType_MATERIAL,
				Limitation: []string{"Hello"},
			}},
		})
		assert.NoError(err)

		station := models.Station{}
		err = db.Where(`id = ?`, testStationB).Find(&station).Error
		assert.NoError(err)
		assert.Equal(models.Station{
			ID:                testStationB,
			AdminDepartmentID: testDepartmentA,
			Sites: []models.UniqueSite{{
				SiteID: models.SiteID{
					Name:  testSiteC,
					Index: 0,
				},
				Station: testStationB,
			}},
			State:       stations.State_SHUTDOWN,
			Information: models.StationInformation{},
			UpdatedAt:   types.ToTimeNano(timeNow),
			UpdatedBy:   testUser,
			CreatedAt:   types.ToTimeNano(timeNow),
			CreatedBy:   testUser,
		}, station)

		rep, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: testStationB,
			SiteName:  testSiteC,
			SiteIndex: 0,
		})
		assert.NoError(err)

		assert.NoError(rep.Attributes.LimitHandler("Hello"))
		assert.ErrorIs(
			rep.Attributes.LimitHandler("Bonjour"),
			mcomErr.Error{Code: mcomErr.Code_PRODUCT_ID_MISMATCH, Details: "product id: Bonjour"},
		)
	}
	{ // create station with sites missing type and subtype: bad case.
		err := dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            testStationB,
			DepartmentOID: testDepartmentA,
			Sites: []mcom.SiteInformation{{
				Name:  testSiteD,
				Index: 0,
			}}})
		assert.ErrorIs(err, mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: fmt.Sprintf("missing type or subtype while adding sites, name: %v, index: %d", testSiteD, 0),
		})
	}
	{ // station already exists.
		err := dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            testStationA,
			DepartmentOID: testDepartmentA,
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_STATION_ALREADY_EXISTS,
		}, err)
	}
	{ // create after delete should pass.
		assert.NoError(dm.DeleteStation(ctx, mcom.DeleteStationRequest{
			StationID: testStationA,
		}))
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            testStationA,
			DepartmentOID: testDepartmentA,
		}))
	}
	{ // create site with empty site name.
		const (
			stationID    = "emptySiteName"
			departmentID = "emptySiteName"
		)
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            stationID,
			DepartmentOID: departmentID,
			Sites: []mcom.SiteInformation{
				{
					Name:       "",
					Index:      0,
					Type:       sites.Type_SLOT,
					SubType:    sites.SubType_MATERIAL,
					Limitation: []string{},
				},
			},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
		}))

		actual, err := dm.GetStation(ctx, mcom.GetStationRequest{
			ID: stationID,
		})
		assert.NoError(err)

		assert.Equal(mcom.GetStationReply{
			ID:                 stationID,
			AdminDepartmentOID: departmentID,
			Sites: []mcom.ListStationSite{
				{
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID:  models.SiteID{},
							Station: stationID,
						},
						Type:    sites.Type_SLOT,
						SubType: sites.SubType_MATERIAL,
					},
					Content: models.SiteContent{
						Slot: &models.Slot{},
					},
				},
			},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
			UpdatedBy:   actual.UpdatedBy,
			UpdatedAt:   actual.UpdatedAt,
			InsertedBy:  actual.InsertedBy,
			InsertedAt:  actual.InsertedAt,
		}, actual)
	}

	assert.NoError(clearStationsData(db))
}

func TestDataManager_DeleteStation(t *testing.T) {
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
	assert.NoError(clearStationsData(db))

	// 統一時間
	timeNow := time.Now().Local()
	monkey.Patch(time.Now, func() time.Time {
		return timeNow
	})
	defer monkey.UnpatchAll()

	{ // insufficient request.
		err := dm.DeleteStation(ctx, mcom.DeleteStationRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // station not found.
		err := dm.DeleteStation(ctx, mcom.DeleteStationRequest{StationID: testStationA})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_STATION_NOT_FOUND,
			Details: fmt.Sprintf("station not found, id: %v", testStationA),
		}, err)
	}
	{ // good case.
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            testStationA,
			DepartmentOID: testDepartmentA,
			Sites:         []mcom.SiteInformation{},
			State:         stations.State_IDLE,
			Information:   mcom.StationInformation{},
		}))
		err = dm.DeleteStation(ctx, mcom.DeleteStationRequest{StationID: testStationA})
		assert.NoError(err)
		// check List function.
		rep, err := dm.ListStations(ctx, mcom.ListStationsRequest{
			DepartmentOID: testDepartmentA,
		})
		assert.NoError(err)
		for _, elem := range rep.Stations {
			assert.NotEqual(elem.ID, testStationA)
		}
	}
	{ // delete sites that belong to the station (should pass)
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "deleteTest",
			DepartmentOID: "321",
			Sites: []mcom.SiteInformation{{
				Name:    "deleteTestSite",
				Index:   0,
				Type:    sites.Type_CONTAINER,
				SubType: sites.SubType_OPERATOR,
			}},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
		}))
		_, err := dm.GetStation(ctx, mcom.GetStationRequest{
			ID: "deleteTest",
		})
		assert.NoError(err)
		_, err = dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "deleteTest",
			SiteName:  "deleteTestSite",
			SiteIndex: 0,
		})
		assert.NoError(err)
		assert.NoError(dm.DeleteStation(ctx, mcom.DeleteStationRequest{
			StationID: "deleteTest",
		}))
		_, err = dm.GetStation(ctx, mcom.GetStationRequest{
			ID: "deleteTest",
		})
		assert.ErrorIs(err, mcomErr.Error{
			Code:    mcomErr.Code_STATION_NOT_FOUND,
			Details: fmt.Sprintf("station not found, id: %v", "deleteTest"),
		})
		_, err = dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "deleteTest",
			SiteName:  "deleteTestSite",
			SiteIndex: 0,
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND})
	}
	{ // delete sites that belong to the station with remaining objects(should not pass)
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "deleteTest2",
			DepartmentOID: "123",
			Sites: []mcom.SiteInformation{{
				Name:    "deleteTest2Site",
				Index:   0,
				Type:    sites.Type_CONTAINER,
				SubType: sites.SubType_OPERATOR,
			}},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
		}))
		db.Where(models.SiteContents{
			Name:  "deleteTest2Site",
			Index: 0,
		}).Updates(&models.SiteContents{
			Content: models.SiteContent{
				Container: &models.Container{{
					Operator: &models.OperatorSite{EmployeeID: "hahaha"},
				}},
			},
		})
		_, err := dm.GetStation(ctx, mcom.GetStationRequest{
			ID: "deleteTest2",
		})
		assert.NoError(err)
		_, err = dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "deleteTest2",
			SiteName:  "deleteTest2Site",
			SiteIndex: 0,
		})
		assert.NoError(err)
		assert.ErrorIs(dm.DeleteStation(ctx, mcom.DeleteStationRequest{
			StationID: "deleteTest2",
		}), mcomErr.Error{Code: mcomErr.Code_STATION_SITE_REMAINING_OBJECTS, Details: "remaining objects in [station: deleteTest2, site name: deleteTest2Site, site index: 0]"})
		_, err = dm.GetStation(ctx, mcom.GetStationRequest{
			ID: "deleteTest2",
		})
		assert.NoError(err)
		_, err = dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "deleteTest2",
			SiteName:  "deleteTest2Site",
			SiteIndex: 0,
		})
		assert.NoError(err)
	}
	{ // dissociate sites that belong to the specified station but have been associated to other stations (should not pass)
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "sharedSiteStation2",
			DepartmentOID: "321",
			Sites: []mcom.SiteInformation{{
				Name:    "deleteTestSite",
				Index:   0,
				Type:    sites.Type_CONTAINER,
				SubType: sites.SubType_OPERATOR,
			}},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
		}))
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "deleteTest3",
			DepartmentOID: "321",
			Sites: []mcom.SiteInformation{{
				Station: "sharedSiteStation2",
				Name:    "deleteTestSite",
				Index:   0,
				Type:    sites.Type_CONTAINER,
				SubType: sites.SubType_OPERATOR,
			}},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
		}))

		assert.EqualError(dm.DeleteStation(ctx, mcom.DeleteStationRequest{
			StationID: "sharedSiteStation2",
		}), "please dissociate the stations first")
	}

	assert.NoError(clearStationsData(db))
}

func TestDataManager_UpdateStation(t *testing.T) {
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
	assert.NoError(clearStationsData(db))

	// 統一時間
	timeNow := time.Now().Local()
	monkey.Patch(time.Now, func() time.Time {
		return timeNow
	})
	defer monkey.UnpatchAll()

	{ // insufficient request.
		err := dm.UpdateStation(ctx, mcom.UpdateStationRequest{
			ID: "",
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // station not found.
		err := dm.UpdateStation(ctx, mcom.UpdateStationRequest{
			ID:    testStationA,
			Sites: []mcom.UpdateStationSite{},
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_STATION_NOT_FOUND,
			Details: fmt.Sprintf("station not found, id: %v", testStationA),
		}, err)
	}

	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            testStationA,
		DepartmentOID: testDepartmentA,
		Sites:         []mcom.SiteInformation{},
		State:         stations.State_IDLE,
		Information:   mcom.StationInformation{},
	}))
	{ // good case: update sites add
		err = dm.UpdateStation(ctx, mcom.UpdateStationRequest{
			ID: testStationA,
			Sites: []mcom.UpdateStationSite{{
				ActionMode: sites.ActionType_ADD,
				Information: mcom.SiteInformation{
					Name:    testSiteC,
					Index:   0,
					Type:    sites.Type_COLLECTION,
					SubType: sites.SubType_MATERIAL,
				},
			}},
		})
		assert.NoError(err)

		station, err := dm.GetStation(ctx, mcom.GetStationRequest{ID: testStationA})
		assert.NoError(err)
		assert.Equal(mcom.GetStationReply{
			ID:                 testStationA,
			AdminDepartmentOID: testDepartmentA,
			Sites: []mcom.ListStationSite{{
				Information: mcom.ListStationSitesInformation{
					UniqueSite: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  testSiteC,
							Index: 0,
						},
						Station: testStationA,
					},

					Type:    sites.Type_COLLECTION,
					SubType: sites.SubType_MATERIAL,
				},
				Content: models.SiteContent{
					Collection: &models.Collection{},
				},
			}},
			State:      stations.State_IDLE,
			UpdatedAt:  timeNow,
			UpdatedBy:  testUser,
			InsertedAt: timeNow,
			InsertedBy: testUser,
		}, station)
	}
	{ // good case: update with empty department.
		assert.NoError(dm.UpdateStation(ctx, mcom.UpdateStationRequest{
			ID:            testStationA,
			DepartmentOID: "",
		}))
		station, err := dm.GetStation(ctx, mcom.GetStationRequest{ID: testStationA})
		assert.NoError(err)
		assert.Equal(testDepartmentA, station.AdminDepartmentOID)

		rep, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: testStationA,
			SiteName:  testSiteC,
			SiteIndex: 0,
		})
		assert.NoError(err)
		assert.Equal(rep.AdminDepartmentOID, testDepartmentA)
	}
	{ // good case: update sites delete.
		_, err = dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: testStationA,
			SiteName:  testSiteC,
			SiteIndex: 0,
		})
		assert.NoError(err)

		err = dm.UpdateStation(ctx, mcom.UpdateStationRequest{
			ID: testStationA,
			Sites: []mcom.UpdateStationSite{{
				ActionMode: sites.ActionType_REMOVE,
				Information: mcom.SiteInformation{
					Name:  testSiteC,
					Index: 0,
				},
			}},
		})
		assert.NoError(err)

		station, err := dm.GetStation(ctx, mcom.GetStationRequest{ID: testStationA})
		assert.NoError(err)
		assert.Equal(mcom.GetStationReply{
			ID:                 testStationA,
			AdminDepartmentOID: testDepartmentA,
			Sites:              []mcom.ListStationSite{},
			State:              stations.State_IDLE,
			UpdatedAt:          timeNow,
			UpdatedBy:          testUser,
			InsertedAt:         timeNow,
			InsertedBy:         testUser,
		}, station)

		_, err = dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: testStationA,
			SiteName:  testSiteC,
			SiteIndex: 0,
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND})
	}
	{ // good case: update state.
		err = dm.UpdateStation(ctx, mcom.UpdateStationRequest{
			ID:    testStationA,
			State: stations.State_RUNNING,
		})
		assert.NoError(err)

		station, err := dm.GetStation(ctx, mcom.GetStationRequest{ID: testStationA})
		assert.NoError(err)
		assert.Equal(mcom.GetStationReply{
			ID:                 testStationA,
			AdminDepartmentOID: testDepartmentA,
			Sites:              []mcom.ListStationSite{},
			State:              stations.State_RUNNING,
			UpdatedAt:          timeNow,
			UpdatedBy:          testUser,
			InsertedAt:         timeNow,
			InsertedBy:         testUser,
		}, station)
	}
	{ // good case: update information.
		err = dm.UpdateStation(ctx, mcom.UpdateStationRequest{
			ID: testStationA,
			Information: mcom.StationInformation{
				Code:        "test",
				Description: "test",
			},
		})
		assert.NoError(err)

		station, err := dm.GetStation(ctx, mcom.GetStationRequest{ID: testStationA})
		assert.NoError(err)
		assert.Equal(mcom.GetStationReply{
			ID:                 testStationA,
			AdminDepartmentOID: testDepartmentA,
			Sites:              []mcom.ListStationSite{},
			State:              stations.State_RUNNING,
			Information: mcom.StationInformation{
				Code:        "test",
				Description: "test",
			},
			UpdatedAt:  timeNow,
			UpdatedBy:  testUser,
			InsertedAt: timeNow,
			InsertedBy: testUser,
		}, station)
	}
	{ // test update : should not update wrong stations.
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "AAA",
			DepartmentOID: "AAA",
			State:         stations.State_IDLE,
		}))
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "BBB",
			DepartmentOID: "BBB",
			State:         stations.State_IDLE,
		}))
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "CCC",
			DepartmentOID: "AAA",
			State:         stations.State_IDLE,
		}))
		rep, err := dm.ListStations(ctx, mcom.ListStationsRequest{
			DepartmentOID: "AAA",
		})
		assert.NoError(err)
		for _, elem := range rep.Stations {
			switch {
			case elem.ID == "AAA":
				assert.Equal("", elem.Information.Description)
			case elem.ID == "CCC":
				assert.Equal("", elem.Information.Description)
			}
		}
		rep, err = dm.ListStations(ctx, mcom.ListStationsRequest{
			DepartmentOID: "BBB",
		})
		assert.NoError(err)
		assert.Equal("", rep.Stations[0].Information.Description)
		assert.NoError(dm.UpdateStation(ctx, mcom.UpdateStationRequest{
			ID:            "AAA",
			DepartmentOID: "AAA",
			State:         stations.State_IDLE,
			Information: mcom.StationInformation{
				Code:        "AAA",
				Description: "AAA",
			},
		}))
		rep, err = dm.ListStations(ctx, mcom.ListStationsRequest{
			DepartmentOID: "AAA",
		})
		assert.NoError(err)
		for _, elem := range rep.Stations {
			switch {
			case elem.ID == "AAA":
				assert.Equal("AAA", elem.Information.Description)
			case elem.ID == "CCC":
				assert.Equal("", elem.Information.Description)
			}
		}
		rep, err = dm.ListStations(ctx, mcom.ListStationsRequest{
			DepartmentOID: "BBB",
		})
		assert.NoError(err)
		assert.Equal("", rep.Stations[0].Information.Description)
	}
	{ // add a site with empty name.
		assert.NoError(dm.UpdateStation(ctx, mcom.UpdateStationRequest{
			ID:            "BBB",
			DepartmentOID: "BBB",
			Sites: []mcom.UpdateStationSite{
				{
					ActionMode: sites.ActionType_ADD,
					Information: mcom.SiteInformation{
						Name:       "",
						Index:      0,
						Type:       sites.Type_SLOT,
						SubType:    sites.SubType_OPERATOR,
						Limitation: []string{},
					},
				},
			},
			State: stations.State_IDLE,
		}))

		actual, err := dm.GetStation(ctx, mcom.GetStationRequest{
			ID: "BBB",
		})
		assert.NoError(err)
		assert.Equal(mcom.GetStationReply{
			ID:                 "BBB",
			AdminDepartmentOID: "BBB",
			Sites: []mcom.ListStationSite{
				{
					Information: mcom.ListStationSitesInformation{
						UniqueSite: models.UniqueSite{
							SiteID:  models.SiteID{},
							Station: "BBB",
						},
						Type:    sites.Type_SLOT,
						SubType: sites.SubType_OPERATOR,
					},
					Content: models.SiteContent{
						Slot: &models.Slot{},
					},
				},
			},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
			UpdatedBy:   actual.UpdatedBy,
			UpdatedAt:   actual.UpdatedAt,
			InsertedBy:  actual.InsertedBy,
			InsertedAt:  actual.InsertedAt,
		}, actual)
	}

	assert.NoError(clearStationsData(db))

}

func TestDataManager_ListStationState(t *testing.T) {
	assert := assert.New(t)
	ctx := commonsCtx.WithUserID(context.Background(), testUser)

	dm, _, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}
	defer dm.Close()

	reply, err := dm.ListStationState(ctx)
	assert.NoError(err)
	expected := mcom.ListStationStateReply{
		{
			Name:  "SHUTDOWN",
			Value: 1,
		},
		{
			Name:  "IDLE",
			Value: 2,
		},
		{
			Name:  "RUNNING",
			Value: 3,
		},
		{
			Name:  "MAINTENANCE",
			Value: 4,
		},
		{
			Name:  "REPAIRING",
			Value: 5,
		},
		{
			Name:  "MALFUNCTION",
			Value: 6,
		},
		{
			Name:  "DISPOSAL",
			Value: 7,
		},
	}
	assert.Equal(expected, reply)
}

func TestDataManager_ListStationIDs(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &models.Station{})
	assert.NoError(cm.Clear())
	// #region build test data
	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            "A1",
		DepartmentOID: "A",
	}))
	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            "A2",
		DepartmentOID: "A",
	}))
	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            "B1",
		DepartmentOID: "B",
	}))
	// #endregion build test data
	{ // list by department (should pass)
		actual, err := dm.ListStationIDs(ctx, mcom.ListStationIDsRequest{DepartmentOID: "A"})
		assert.NoError(err)
		assert.Equal(mcom.ListStationIDsReply{Stations: []string{"A1", "A2"}}, actual)
	}
	{ // list all stations (should pass)
		actual, err := dm.ListStationIDs(ctx, mcom.ListStationIDsRequest{})
		assert.NoError(err)
		assert.Equal(mcom.ListStationIDsReply{Stations: []string{"A1", "A2", "B1"}}, actual)
	}
	assert.NoError(cm.Clear())
}
