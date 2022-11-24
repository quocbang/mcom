package impl

import (
	"context"
	"fmt"
	"testing"
	"time"

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

const (
	testSiteContainer = "container"
	testSiteSlot      = "slot"
	testSiteColqueue  = "colqueue"
)

func deleteAllListSiteMaterialsData(db *gorm.DB) error {
	return newClearMaster(db, &models.Station{}, &models.Site{}, &models.SiteContents{}).Clear()
}

func Test_ListSiteMaterials(t *testing.T) {
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

	assert.NoError(deleteAllListSiteMaterialsData(db))

	sDB, err := db.DB()
	if !assert.NoError(err) {
		return
	}
	defer sDB.Close()
	{ // station not found (with empty site name)
		_, err = dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
			Station: testStationA,
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_STATION_NOT_FOUND,
		}, err)
	}
	{ // station not found.
		_, err = dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
			Station: testStationA,
			Site:    models.SiteID{Name: testSiteContainer},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_STATION_NOT_FOUND,
		}, err)
	}
	{ // station does not contain specified site.
		err = db.Create(&models.Station{
			ID:                testStationA,
			AdminDepartmentID: testDepartmentA,
			Sites: []models.UniqueSite{
				{
					SiteID: models.SiteID{
						Name:  testSiteContainer,
						Index: 0,
					},
					Station: testStationA,
				},
			},
			State:     stations.State_RUNNING,
			UpdatedBy: testUser,
			CreatedBy: testUser,
		}).Error
		assert.NoError(err)

		_, err = dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
			Station: testStationA,
			Site:    models.SiteID{Name: "bad_site"},
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND})
	}
	{ // site not found.
		_, err = dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
			Station: testStationA,
			Site:    models.SiteID{Name: testSiteContainer},
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_STATION_SITE_NOT_FOUND,
			Details: fmt.Sprintf("site not found, station: %s, name: %s, index: %d", testStationA, testSiteContainer, 0),
		}, err)
	}
	{ // site content not found.
		err = db.Create(&models.Site{
			Name:              testSiteContainer,
			Index:             0,
			AdminDepartmentID: testDepartmentA,
			Attributes: models.SiteAttributes{
				Type:    sites.Type_CONTAINER,
				SubType: sites.SubType_MATERIAL,
			},
			Station:   testStationA,
			UpdatedBy: testUser,
			CreatedBy: testUser,
		}).Error
		assert.NoError(err)

		_, err = dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
			Station: testStationA,
			Site:    models.SiteID{Name: testSiteContainer},
		})
		assert.EqualError(fmt.Errorf("site content not found, station: %s, name: %s, index: %d", testStationA, testSiteContainer, 0), err.Error())
	}
	{ // good case, empty site.
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "EmptySiteStation",
			DepartmentOID: testDepartmentA,
			Sites: []mcom.SiteInformation{{
				Name:    "EmptyContainer",
				Index:   0,
				Type:    sites.Type_CONTAINER,
				SubType: sites.SubType_MATERIAL,
			}},
			State: stations.State_IDLE,
		}))
		results, err := dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
			Station: "EmptySiteStation",
			Site: models.SiteID{
				Name:  "EmptyContainer",
				Index: 0,
			},
		})
		assert.NoError(err)
		assert.Equal(mcom.ListSiteMaterialsReply{}, results)
	}
	{ // good case, return container material.
		err = db.Create(&models.SiteContents{
			Name:    testSiteContainer,
			Station: testStationA,
			Content: models.SiteContent{
				Container: &models.Container{
					{
						Material: &models.MaterialSite{
							Material: models.Material{
								ID:    testMaterialName,
								Grade: testMaterialGrade,
							},
							Quantity:   types.Decimal.NewFromFloat32(123.456),
							ResourceID: testResourceID,
						},
					}, {
						Material: &models.MaterialSite{
							Material: models.Material{
								ID:    testMaterialName,
								Grade: testMaterialGrade,
							},
							Quantity:   types.Decimal.NewFromFloat32(654.321),
							ResourceID: testResourceID,
						},
					},
				},
			},
			UpdatedBy: testUser,
		}).Error
		assert.NoError(err)

		results, err := dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
			Station: testStationA,
			Site:    models.SiteID{Name: testSiteContainer},
		})
		assert.NoError(err)
		assert.Equal(mcom.ListSiteMaterialsReply{{
			ID:         testMaterialName,
			Grade:      testMaterialGrade,
			ResourceID: testResourceID,
			Quantity:   types.Decimal.NewFromFloat32(123.456),
			ExpiryTime: types.TimeNano(0).Time(),
		}, {
			ID:         testMaterialName,
			Grade:      testMaterialGrade,
			ResourceID: testResourceID,
			Quantity:   types.Decimal.NewFromFloat32(654.321),
			ExpiryTime: types.TimeNano(0).Time(),
		}}, results)
	}
	{ // good case, return slot material.
		err = db.Create(&models.Station{
			ID:                testStationB,
			AdminDepartmentID: testDepartmentA,
			Sites: []models.UniqueSite{{
				SiteID:  models.SiteID{Name: testSiteSlot},
				Station: testStationB,
			}},
			State:     stations.State_RUNNING,
			UpdatedBy: testUser,
			CreatedBy: testUser,
		}).Error
		assert.NoError(err)

		err = db.Create(&models.Site{
			Name:              testSiteSlot,
			AdminDepartmentID: testWorkOrderDepartmentOID,
			Attributes: models.SiteAttributes{
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_MATERIAL,
			},
			Station:   testStationB,
			UpdatedBy: testUser,
			CreatedBy: testUser,
		}).Error
		assert.NoError(err)

		err = db.Create(&models.SiteContents{
			Name:    testSiteSlot,
			Station: testStationB,
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    testMaterialName,
							Grade: testMaterialGrade,
						},
						Quantity:   types.Decimal.NewFromFloat32(123.456),
						ResourceID: testResourceID,
					},
				},
			},
			UpdatedBy: testUser,
		}).Error
		assert.NoError(err)

		results, err := dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
			Station: testStationB,
			Site:    models.SiteID{Name: testSiteSlot},
		})
		assert.NoError(err)
		assert.Equal(mcom.ListSiteMaterialsReply{{
			ID:         testMaterialName,
			Grade:      testMaterialGrade,
			ResourceID: testResourceID,
			Quantity:   types.Decimal.NewFromFloat32(123.456),
			ExpiryTime: types.TimeNano(0).Time(),
		}}, results)
	}
	{ // good case, return colqueue material.
		err = db.Create(&models.Station{
			ID:                testStationC,
			AdminDepartmentID: testDepartmentA,
			Sites: []models.UniqueSite{{
				SiteID:  models.SiteID{Name: testSiteColqueue},
				Station: testStationC,
			}},
			State:     stations.State_RUNNING,
			UpdatedBy: testUser,
			CreatedBy: testUser,
		}).Error
		assert.NoError(err)

		err = db.Create(&models.Site{
			Name:              testSiteColqueue,
			AdminDepartmentID: testWorkOrderDepartmentOID,
			Attributes: models.SiteAttributes{
				Type:    sites.Type_COLQUEUE,
				SubType: sites.SubType_MATERIAL,
			},
			Station:   testStationC,
			UpdatedBy: testUser,
			CreatedBy: testUser,
		}).Error
		assert.NoError(err)

		err = db.Create(&models.SiteContents{
			Name:    testSiteColqueue,
			Station: testStationC,
			Content: models.SiteContent{
				Colqueue: &models.Colqueue{
					{
						{
							Material: &models.MaterialSite{
								Material: models.Material{
									ID:    "A1",
									Grade: "A",
								},
								Quantity:   types.Decimal.NewFromInt32(123),
								ResourceID: "A1001",
							},
						}, {
							Material: &models.MaterialSite{
								Material: models.Material{
									ID:    "A2",
									Grade: "A",
								},
								Quantity:   types.Decimal.NewFromInt32(456),
								ResourceID: "A2001",
							},
						},
					}, {
						{
							Material: &models.MaterialSite{
								Material: models.Material{
									ID:    "B1",
									Grade: "B",
								},
								Quantity:   types.Decimal.NewFromInt32(123),
								ResourceID: "B1001",
							},
						}, {
							Material: &models.MaterialSite{
								Material: models.Material{
									ID:    "B2",
									Grade: "B",
								},
								Quantity:   types.Decimal.NewFromInt32(456),
								ResourceID: "B2001",
							},
						},
					},
				},
			},
			UpdatedBy: testUser,
		}).Error
		assert.NoError(err)

		results, err := dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
			Station: testStationC,
			Site:    models.SiteID{Name: testSiteColqueue},
		})
		assert.NoError(err)
		assert.Equal(mcom.ListSiteMaterialsReply{{
			ID:         "A1",
			Grade:      "A",
			ResourceID: "A1001",
			Quantity:   types.Decimal.NewFromFloat32(123),
			ExpiryTime: types.TimeNano(0).Time(),
		}, {
			ID:         "A2",
			Grade:      "A",
			ResourceID: "A2001",
			Quantity:   types.Decimal.NewFromFloat32(456),
			ExpiryTime: types.TimeNano(0).Time(),
		}, {
			ID:         "B1",
			Grade:      "B",
			ResourceID: "B1001",
			Quantity:   types.Decimal.NewFromFloat32(123),
			ExpiryTime: types.TimeNano(0).Time(),
		}, {
			ID:         "B2",
			Grade:      "B",
			ResourceID: "B2001",
			Quantity:   types.Decimal.NewFromFloat32(456),
			ExpiryTime: types.TimeNano(0).Time(),
		}}, results)
	}
	{ // good case, empty slot.
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "emptySlot",
			DepartmentOID: testDepartmentA,
			Sites: []mcom.SiteInformation{{
				Name:    "emptySlot",
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_MATERIAL,
			}},
			State: stations.State_IDLE,
		}))

		actual, err := dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
			Station: "emptySlot",
			Site: models.SiteID{
				Name:  "emptySlot",
				Index: 0,
			},
		})
		assert.NoError(err)

		assert.Equal(mcom.ListSiteMaterialsReply{}, actual)
	}
	{ // good case, empty site name
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "emptySiteName",
			DepartmentOID: "emptySiteNameOID",
			Sites: []mcom.SiteInformation{
				{
					Station:    "emptySiteName",
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

		assert.NoError(dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{
				{
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
						Quantity:    types.Decimal.NewFromInt32(10),
						ResourceID:  "A",
						ProductType: "A",
						Status:      resources.MaterialStatus_AVAILABLE,
						ExpiryTime:  0,
					}},
					Option: models.BindOption{},
				},
			},
		}))

		rep, err := dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
			Station: "emptySiteName",
			Site: models.SiteID{
				Name:  "",
				Index: 0,
			},
		})

		assert.NoError(err)

		assert.Equal(mcom.ListSiteMaterialsReply{
			{
				ID:         "A",
				Grade:      "A",
				Quantity:   types.Decimal.NewFromInt32(10),
				ResourceID: "A",
				ExpiryTime: time.UnixMilli(0),
			},
		}, rep)
	}
	assert.NoError(deleteAllListSiteMaterialsData(db))
}

func Test_SiteMaterialExpiryTime(t *testing.T) {
	ctx, dm, db := initializeDB(t)
	assert := assert.New(t)
	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            "testStation",
		DepartmentOID: "testDepartment",
		Sites: []mcom.SiteInformation{{
			Name:    "testSite",
			Index:   0,
			Type:    sites.Type_SLOT,
			SubType: sites.SubType_MATERIAL,
		}},
		State: stations.State_IDLE,
	}))

	_, err := dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{{
			Type:           "abc",
			ID:             "TestMaterial",
			Status:         resources.MaterialStatus_AVAILABLE,
			Quantity:       decimal.NewFromInt(123),
			ProductionTime: types.TimeNano(0).Time(),
			ExpiryTime:     types.TimeNano(12345).Time(),
			ResourceID:     "testResource",
		}},
	})
	assert.NoError(err)

	assert.NoError(dm.MaterialResourceBind(ctx, mcom.MaterialResourceBindRequest{
		Station: "testStation",
		Details: []mcom.MaterialBindRequestDetail{
			{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{
					Name:  "testSite",
					Index: 0,
				},
				Resources: []mcom.BindMaterialResource{
					{
						Material: models.Material{
							ID:    "TestMaterial",
							Grade: "A",
						},
						Quantity:    types.Decimal.NewFromInt32(10),
						ResourceID:  "testResource",
						ProductType: "abc",
						Warehouse:   mcom.Warehouse{},
						ExpiryTime:  types.TimeNano(12345),
					},
				},
				Option: models.BindOption{},
			},
		},
	}))

	actual, err := dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
		Station: "testStation",
		Site: models.SiteID{
			Name:  "testSite",
			Index: 0,
		},
	})
	assert.NoError(err)
	assert.Equal(types.TimeNano(12345).Time(), actual[0].ExpiryTime)

	assert.NoError(clearMaterialResourceBindTestDB(db))
}
