package impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func Test_ListSiteType(t *testing.T) {
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

	{
		types, err := dm.ListSiteType(ctx)
		assert.NoError(err)
		assert.Equal(mcom.ListSiteTypeReply{
			{
				Name:  "CONTAINER",
				Value: 1,
			}, {
				Name:  "SLOT",
				Value: 2,
			}, {
				Name:  "COLLECTION",
				Value: 3,
			}, {
				Name:  "QUEUE",
				Value: 4,
			}, {
				Name:  "COLQUEUE",
				Value: 5,
			},
		}, types)
	}
}

func Test_ListSiteSubType(t *testing.T) {
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

	{
		types, err := dm.ListSiteSubType(ctx)
		assert.NoError(err)
		assert.Equal(mcom.ListSiteSubTypeReply{
			{
				Name:  "OPERATOR",
				Value: 1,
			}, {
				Name:  "MATERIAL",
				Value: 2,
			}, {
				Name:  "TOOL",
				Value: 3,
			},
		}, types)
	}
}

func TestDataManager_GetSite(t *testing.T) {
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
			Name:    "AAA0",
			Index:   0,
			Type:    sites.Type_COLLECTION,
			SubType: sites.SubType_MATERIAL,
		}, {
			Name:    "AAA1",
			Index:   1,
			Type:    sites.Type_SLOT,
			SubType: sites.SubType_MATERIAL,
		}},
		State:       stations.State_IDLE,
		Information: mcom.StationInformation{},
	}))
	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            "BBB",
		DepartmentOID: "BBB",
		Sites: []mcom.SiteInformation{{
			Name:    "BBB0",
			Index:   0,
			Type:    sites.Type_SLOT,
			SubType: sites.SubType_OPERATOR,
		}},
		State:       stations.State_IDLE,
		Information: mcom.StationInformation{},
	}))
	{ // not found (should not pass)
		_, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "AAA",
			SiteName:  "CCC0",
			SiteIndex: 0,
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND})
		_, err = dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "AAA",
			SiteName:  "BBB0",
			SiteIndex: 0,
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND})
	}
	{ // normal case (should pass)
		assert.NoError(db.Where(&models.SiteContents{Name: "AAA0"}).Updates(&models.SiteContents{
			Name:  "AAA0",
			Index: 0,
			Content: models.SiteContent{
				Collection: &models.Collection{{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "AAA0M",
							Grade: "C",
						},
						Quantity:   types.Decimal.NewFromFloat32(200),
						ResourceID: "AAA0M",
					},
				}},
			},
		}).Error)
		actual, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "AAA",
			SiteName:  "AAA0",
			SiteIndex: 0,
		})
		assert.NoError(err)
		expected := mcom.GetSiteReply{
			Name:               "AAA0",
			Index:              0,
			AdminDepartmentOID: "AAA",
			Attributes: mcom.SiteAttributes{
				Type:    sites.Type_COLLECTION,
				SubType: sites.SubType_MATERIAL,
			},
			Content: models.SiteContent{
				Collection: &models.Collection{{
					Material: &models.MaterialSite{
						Material: models.Material{
							ID:    "AAA0M",
							Grade: "C",
						},
						Quantity:   types.Decimal.NewFromInt32(200),
						ResourceID: "AAA0M",
					},
				}}},
		}

		actual.Attributes.LimitHandler = nil // function assertion will always fail.
		assert.Equal(expected, actual)
		actual, err = dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "BBB",
			SiteName:  "BBB0",
			SiteIndex: 0,
		})
		assert.NoError(err)
		expected = mcom.GetSiteReply{
			Name:               "BBB0",
			Index:              0,
			AdminDepartmentOID: "BBB",
			Attributes: mcom.SiteAttributes{
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_OPERATOR,
			},
			Content: models.SiteContent{Slot: &models.Slot{}},
		}

		handler := actual.Attributes.LimitHandler
		actual.Attributes.LimitHandler = nil // function assertion will always fail.
		assert.Equal(expected, actual)
		assert.NoError(handler("any"))
	}
	{ // test limitation.
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "testLimitation",
			DepartmentOID: "AAA",
			Sites: []mcom.SiteInformation{{
				Name:       "testLimitation",
				Index:      0,
				Type:       sites.Type_SLOT,
				SubType:    sites.SubType_MATERIAL,
				Limitation: []string{"apple", "pen"},
			}},
		}))

		actual, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "testLimitation",
			SiteName:  "testLimitation",
			SiteIndex: 0,
		})
		assert.NoError(err)

		handler := actual.Attributes.LimitHandler
		actual.Attributes.LimitHandler = nil

		expected := mcom.GetSiteReply{
			Name:               "testLimitation",
			Index:              0,
			AdminDepartmentOID: "AAA",
			Attributes: mcom.SiteAttributes{
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_MATERIAL,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{},
			},
		}
		assert.Equal(expected, actual)

		assert.NoError(handler("apple"))
		assert.ErrorIs(handler("apple pen"), mcomErr.Error{Code: mcomErr.Code_PRODUCT_ID_MISMATCH, Details: "product id: apple pen"})
	}
	{ // should not get a wrong site
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "CCC",
			DepartmentOID: "CCC",
			Sites: []mcom.SiteInformation{
				{
					Station:    "BBB",
					Name:       "BBB0",
					Index:      0,
					Type:       sites.Type_SLOT,
					SubType:    sites.SubType_OPERATOR,
					Limitation: []string{},
				},
				{
					Name:       "BBB0",
					Index:      0,
					Type:       sites.Type_CONTAINER,
					SubType:    sites.SubType_MATERIAL,
					Limitation: []string{},
				},
			},
			State:       stations.State_IDLE,
			Information: mcom.StationInformation{},
		}))

		actual, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "CCC",
			SiteName:  "BBB0",
			SiteIndex: 0,
		})
		assert.NoError(err)

		actual.Attributes.LimitHandler = nil
		assert.Equal(mcom.GetSiteReply{
			Name:               "BBB0",
			Index:              0,
			AdminDepartmentOID: "CCC",
			Attributes: mcom.SiteAttributes{
				Type:         sites.Type_CONTAINER,
				SubType:      sites.SubType_MATERIAL,
				LimitHandler: nil,
			},
			Content: models.SiteContent{
				Container: &models.Container{},
			},
		}, actual)
	}
}

func TestSiteContents_AfterFind(t *testing.T) {
	assert := assert.New(t)
	_, _, db := initializeDB(t)

	cm := newClearMaster(db, models.SiteContents{})
	assert.NoError(cm.Clear())

	{ // replace the contents in container, collection and colqueue sites.
		originSiteContents := models.SiteContents{
			Name:    "test",
			Index:   0,
			Station: "test",
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				},
				Container: &models.Container{{
					Material: &models.MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				}},
				Collection: &models.Collection{{
					Material: &models.MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				}},
				Queue: &models.Queue{{
					Material: &models.MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				}},
				Colqueue: &models.Colqueue{{{
					Material: &models.MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				}}},
			},
		}

		assert.NoError(db.Create(&originSiteContents).Error)

		var foundSiteContents models.SiteContents
		assert.NoError(db.Model(&models.SiteContents{}).Take(&foundSiteContents).Error)

		expected := models.SiteContents{
			Name:    "test",
			Index:   0,
			Station: "test",
			Content: models.SiteContent{
				Slot: &models.Slot{
					Material: &models.MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				},
				Container: &models.Container{{
					Material: &models.MaterialSite{
						Quantity:   types.Decimal.NewFromInt32(0),
						ResourceID: "id",
					},
				}},
				Collection: &models.Collection{{
					Material: &models.MaterialSite{
						Quantity:   types.Decimal.NewFromInt32(0),
						ResourceID: "id",
					},
				}},
				Queue: &models.Queue{{
					Material: &models.MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				}},
				Colqueue: &models.Colqueue{{{
					Material: &models.MaterialSite{
						Quantity:   types.Decimal.NewFromInt32(0),
						ResourceID: "id",
					},
				}}},
			},
			UpdatedAt: foundSiteContents.UpdatedAt,
		}

		assert.Equal(expected, foundSiteContents)
	}
	{ // do not replace the contents in operator and tool sites.
		originSiteContents := models.SiteContents{
			Name: "test2",
			Content: models.SiteContent{
				Container: &models.Container{{
					Tool: &models.ToolSite{
						ResourceID: "t",
						ToolID:     "t",
					},
				}},
			},
		}

		assert.NoError(db.Create(&originSiteContents).Error)

		var foundSiteContents models.SiteContents
		assert.NoError(db.Model(&models.SiteContents{}).Where(&models.SiteContents{Name: "test2"}).Take(&foundSiteContents).Error)

		expected := models.SiteContents{
			Name: "test2",
			Content: models.SiteContent{
				Container: &models.Container{{
					Tool: &models.ToolSite{
						ResourceID: "t",
						ToolID:     "t",
					},
				}},
			},
			UpdatedAt: foundSiteContents.UpdatedAt,
		}

		assert.Equal(expected, foundSiteContents)
	}
}

func TestDataManager_ListAssociatedStations(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, models.Station{}, models.SiteContents{}, models.Site{})
	assert.NoError(cm.Clear())
	{ // good case
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "station1",
			DepartmentOID: "dep",
			Sites: []mcom.SiteInformation{
				{
					Name:       "site1",
					Index:      0,
					Type:       sites.Type_SLOT,
					SubType:    sites.SubType_TOOL,
					Limitation: []string{},
				},
			},
			State: stations.State_IDLE,
		}))

		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "station2",
			DepartmentOID: "dep",
			Sites: []mcom.SiteInformation{
				{
					Station:    "station1",
					Name:       "site1",
					Index:      0,
					Type:       sites.Type_SLOT,
					SubType:    sites.SubType_TOOL,
					Limitation: []string{},
				},
			},
			State: stations.State_IDLE,
		}))

		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "station3",
			DepartmentOID: "dep",
			State:         stations.State_IDLE,
		}))

		actual, err := dm.ListAssociatedStations(ctx, mcom.ListAssociatedStationsRequest{Site: models.UniqueSite{
			SiteID: models.SiteID{
				Name:  "site1",
				Index: 0,
			},
			Station: "station1",
		}})
		assert.NoError(err)

		assert.Equal(mcom.ListAssociatedStationsReply{
			StationIDs: []string{"station1", "station2"},
		}, actual)
	}
}
