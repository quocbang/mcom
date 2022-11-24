package impl

import (
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/bindtype"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func TestDataManager_GetToolResource(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &models.ToolResource{})
	assert.NoError(cm.Clear())

	// #region create test data
	db.Create(&models.ToolResource{
		ID:        "resource id",
		ToolID:    "tool id",
		UpdatedBy: "rockefeller",
		CreatedBy: "rockefeller",
	})
	// #endregion create test data

	{ // not found.
		_, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{ResourceID: "not found"})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_RESOURCE_NOT_FOUND}, err)
	}
	{ // good case.
		actual, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{ResourceID: "resource id"})
		assert.NoError(err)

		expected := mcom.GetToolResourceReply{
			ToolID:    "tool id",
			CreatedBy: "rockefeller",
			CreatedAt: actual.CreatedAt,
		}

		assert.Equal(expected, actual)
	}

	assert.NoError(cm.Clear())
}

func TestDataManager_ToolResourceBind(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &models.Station{}, &models.Site{}, &models.SiteContents{}, &models.ToolResource{})
	assert.NoError(cm.Clear())

	// #region create test data
	// #region station and sites
	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            "station",
		DepartmentOID: "doid",
		Sites: []mcom.SiteInformation{
			{
				Name:    "slot0",
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_TOOL,
			},
			{
				Name:    "slot1",
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_TOOL,
			},
			{
				Name:    "slot1",
				Index:   1,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_TOOL,
			},
			{
				Name:    "wrong type",
				Index:   0,
				Type:    sites.Type_QUEUE,
				SubType: sites.SubType_TOOL,
			},
			{
				Name:    "wrong sub type",
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_MATERIAL,
			},
		},
		State: stations.State_IDLE,
	}))
	// #endregion station and sites
	// #region tools
	tools := []models.ToolResource{
		{
			ID:     "A",
			ToolID: "A",
		},
		{
			ID:     "B",
			ToolID: "B",
		},
		{
			ID:     "C",
			ToolID: "C",
		},
		{
			ID:     "D",
			ToolID: "D",
		},
		{
			ID:     "E",
			ToolID: "E",
		},
		{
			ID:     "F",
			ToolID: "F",
		},
	}
	assert.NoError(db.Create(&tools).Error)
	// #endregion tools
	// #endregion create test data

	// #region bad cases
	{ // bind without a resource.
		assert.ErrorIs(dm.ToolResourceBind(ctx, mcom.ToolResourceBindRequest{
			Station: "station",
			Details: []mcom.ToolBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "slot0",
						Index: 0,
					},
				},
			},
		}), mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "Key: 'ToolResource.ResourceID' Error:Field validation for 'ResourceID' failed on the 'required' tag",
		})
	}
	{ // clear with resources.
		assert.ErrorIs(dm.ToolResourceBind(ctx, mcom.ToolResourceBindRequest{
			Station: "station",
			Details: []mcom.ToolBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR,
					Site: models.SiteID{
						Name:  "slot0",
						Index: 0,
					},
					Resource: mcom.ToolResource{
						ResourceID: "A",
						ToolID:     "A",
					},
				},
			},
		}), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "there should be no resources while clearing a slot.",
		})
	}
	{ // illegal site type.
		assert.EqualError(dm.ToolResourceBind(ctx, mcom.ToolResourceBindRequest{
			Station: "station",
			Details: []mcom.ToolBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "wrong type",
						Index: 0,
					},
					Resource: mcom.ToolResource{
						ResourceID: "A",
						ToolID:     "A",
					},
				},
			},
		}), "STATION_SITE_SUB_TYPE_MISMATCH: illegal site type or sub type, site name: wrong type, site index: 0")
	}
	{ // illegal site sub type.
		assert.EqualError(dm.ToolResourceBind(ctx, mcom.ToolResourceBindRequest{
			Station: "station",
			Details: []mcom.ToolBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "wrong sub type",
						Index: 0,
					},
					Resource: mcom.ToolResource{
						ResourceID: "A",
						ToolID:     "A",
					},
				},
			},
		}), "STATION_SITE_SUB_TYPE_MISMATCH: illegal site type or sub type, site name: wrong sub type, site index: 0")
	}
	{ // site not found.
		assert.ErrorIs(dm.ToolResourceBind(ctx, mcom.ToolResourceBindRequest{
			Station: "station",
			Details: []mcom.ToolBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "not found",
						Index: 0,
					},
					Resource: mcom.ToolResource{
						ResourceID: "A",
						ToolID:     "A",
					},
				},
			},
		}), mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND})
	}
	// #endregion bad cases

	// #region good cases
	{ // bind tool.
		assert.NoError(dm.ToolResourceBind(ctx, mcom.ToolResourceBindRequest{
			Station: "station",
			Details: []mcom.ToolBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "slot0",
						Index: 0,
					},
					Resource: mcom.ToolResource{
						ResourceID: "A",
						ToolID:     "A",
					},
				},
			},
		}))

		actualSite, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "slot0",
			SiteIndex: 0,
		})
		assert.NoError(err)
		actualSite.Attributes.LimitHandler = nil
		expectedSite := mcom.GetSiteReply{
			Name:               "slot0",
			Index:              0,
			AdminDepartmentOID: "doid",
			Attributes: mcom.SiteAttributes{
				Type:         sites.Type_SLOT,
				SubType:      sites.SubType_TOOL,
				LimitHandler: nil,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{
					Tool: &models.ToolSite{
						ResourceID:    "A",
						ToolID:        "A",
						InstalledTime: actualSite.Content.Slot.Tool.InstalledTime,
					},
				},
			},
		}
		assert.Equal(expectedSite, actualSite)

		actualTool, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
			ResourceID: "A",
		})
		assert.NoError(err)
		expectedTool := mcom.GetToolResourceReply{
			ToolID: "A",
			BindingSite: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "slot0",
					Index: 0,
				},
				Station: "station",
			},
			CreatedBy: "",
			CreatedAt: actualTool.CreatedAt,
		}
		assert.Equal(expectedTool, actualTool)
	}
	{ // clear.
		assert.NoError(dm.ToolResourceBind(ctx, mcom.ToolResourceBindRequest{
			Station: "station",
			Details: []mcom.ToolBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR,
					Site: models.SiteID{
						Name:  "slot0",
						Index: 0,
					},
				},
			},
		}))

		actualSite, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "slot0",
			SiteIndex: 0,
		})
		assert.NoError(err)
		actualSite.Attributes.LimitHandler = nil
		expectedSite := mcom.GetSiteReply{
			Name:               "slot0",
			Index:              0,
			AdminDepartmentOID: "doid",
			Attributes: mcom.SiteAttributes{
				Type:         sites.Type_SLOT,
				SubType:      sites.SubType_TOOL,
				LimitHandler: nil,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{},
			},
		}
		assert.Equal(expectedSite, actualSite)

		actualTool, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
			ResourceID: "A",
		})
		assert.NoError(err)
		expectedTool := mcom.GetToolResourceReply{
			ToolID:    "A",
			CreatedBy: "",
			CreatedAt: actualTool.CreatedAt,
		}
		assert.Equal(expectedTool, actualTool)
	}
	{ // multiple bind
		assert.NoError(dm.ToolResourceBind(ctx, mcom.ToolResourceBindRequest{
			Station: "station",
			Details: []mcom.ToolBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "slot1",
						Index: 0,
					},
					Resource: mcom.ToolResource{
						ResourceID: "B",
						ToolID:     "B",
					},
				},
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "slot1",
						Index: 1,
					},
					Resource: mcom.ToolResource{
						ResourceID: "C",
						ToolID:     "C",
					},
				},
			},
		}))

		actualSite, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "slot1",
			SiteIndex: 0,
		})
		assert.NoError(err)
		actualSite.Attributes.LimitHandler = nil
		expectedSite := mcom.GetSiteReply{
			Name:               "slot1",
			Index:              0,
			AdminDepartmentOID: "doid",
			Attributes: mcom.SiteAttributes{
				Type:         sites.Type_SLOT,
				SubType:      sites.SubType_TOOL,
				LimitHandler: nil,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{
					Tool: &models.ToolSite{
						ResourceID:    "B",
						ToolID:        "B",
						InstalledTime: actualSite.Content.Slot.Tool.InstalledTime,
					},
				},
			},
		}
		assert.Equal(expectedSite, actualSite)

		actualSite2, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "slot1",
			SiteIndex: 0,
		})
		assert.NoError(err)
		actualSite2.Attributes.LimitHandler = nil
		expectedSite2 := mcom.GetSiteReply{
			Name:               "slot1",
			Index:              0,
			AdminDepartmentOID: "doid",
			Attributes: mcom.SiteAttributes{
				Type:         sites.Type_SLOT,
				SubType:      sites.SubType_TOOL,
				LimitHandler: nil,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{
					Tool: &models.ToolSite{
						ResourceID:    "B",
						ToolID:        "B",
						InstalledTime: actualSite2.Content.Slot.Tool.InstalledTime,
					},
				},
			},
		}
		assert.Equal(expectedSite2, actualSite2)

		actualTool, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
			ResourceID: "B",
		})
		assert.NoError(err)
		expectedTool := mcom.GetToolResourceReply{
			ToolID: "B",
			BindingSite: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "slot1",
					Index: 0,
				},
				Station: "station",
			},
			CreatedBy: "",
			CreatedAt: actualTool.CreatedAt,
		}
		assert.Equal(expectedTool, actualTool)

		actualTool2, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
			ResourceID: "C",
		})
		assert.NoError(err)
		expectedTool2 := mcom.GetToolResourceReply{
			ToolID: "C",
			BindingSite: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "slot1",
					Index: 1,
				},
				Station: "station",
			},
			CreatedBy: "",
			CreatedAt: actualTool2.CreatedAt,
		}
		assert.Equal(expectedTool2, actualTool2)
	}
	{ // bind to a site which is not empty
		assert.NoError(dm.ToolResourceBind(ctx, mcom.ToolResourceBindRequest{
			Station: "station",
			Details: []mcom.ToolBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{
					Name:  "slot1",
					Index: 1,
				},
				Resource: mcom.ToolResource{
					ResourceID:  "F",
					ToolID:      "F",
					BindingSite: models.UniqueSite{},
				},
			}},
		}))

		actualTool, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
			ResourceID: "C",
		})
		assert.NoError(err)
		expectedTool := mcom.GetToolResourceReply{
			ToolID:      "C",
			BindingSite: models.UniqueSite{},
			CreatedBy:   "",
			CreatedAt:   actualTool.CreatedAt,
		}
		assert.Equal(expectedTool, actualTool)

		actualTool2, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
			ResourceID: "F",
		})
		assert.NoError(err)
		expectedTool2 := mcom.GetToolResourceReply{
			ToolID: "F",
			BindingSite: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "slot1",
					Index: 1,
				},
				Station: "station",
			},
			CreatedBy: "",
			CreatedAt: actualTool2.CreatedAt,
		}
		assert.Equal(expectedTool2, actualTool2)
	}
	// #endregion good cases
	assert.NoError(cm.Clear())
}

func TestDataManager_ToolResourceBindV2(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &models.Station{}, &models.Site{}, &models.SiteContents{}, &models.ToolResource{})
	assert.NoError(cm.Clear())

	// #region create test data
	// #region station and sites
	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            "station",
		DepartmentOID: "doid",
		Sites: []mcom.SiteInformation{
			{
				Name:    "slot0",
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_TOOL,
			},
			{
				Name:    "slot1",
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_TOOL,
			},
			{
				Name:    "slot1",
				Index:   1,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_TOOL,
			},
			{
				Name:    "wrong type",
				Index:   0,
				Type:    sites.Type_QUEUE,
				SubType: sites.SubType_TOOL,
			},
			{
				Name:    "wrong sub type",
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_MATERIAL,
			},
		},
		State: stations.State_IDLE,
	}))
	// #endregion station and sites
	// #region tools
	tools := []models.ToolResource{
		{
			ID:     "A",
			ToolID: "A",
		},
		{
			ID:     "B",
			ToolID: "B",
		},
		{
			ID:     "C",
			ToolID: "C",
		},
		{
			ID:     "D",
			ToolID: "D",
		},
		{
			ID:     "E",
			ToolID: "E",
		},
		{
			ID:     "F",
			ToolID: "F",
		},
	}
	assert.NoError(db.Create(&tools).Error)
	// #endregion tools
	// #endregion create test data

	// #region bad cases
	{ // bind without a resource
		assert.ErrorIs(dm.ToolResourceBindV2(ctx, mcom.ToolResourceBindRequestV2{
			Details: []mcom.ToolBindRequestDetailV2{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "slot0",
							Index: 0,
						},
						Station: "station",
					},
				},
			},
		}), mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "Key: 'ToolResource.ResourceID' Error:Field validation for 'ResourceID' failed on the 'required' tag",
		})
	}
	{ // clear with resources
		assert.ErrorIs(dm.ToolResourceBindV2(ctx, mcom.ToolResourceBindRequestV2{
			Details: []mcom.ToolBindRequestDetailV2{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR,
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "slot0",
							Index: 0,
						},
						Station: "station",
					},
					Resource: mcom.ToolResource{
						ResourceID: "A",
						ToolID:     "A",
					},
				},
			},
		}), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "there should be no resources while clearing a slot.",
		})
	}
	{ // illegal site type
		assert.EqualError(dm.ToolResourceBindV2(ctx, mcom.ToolResourceBindRequestV2{
			Details: []mcom.ToolBindRequestDetailV2{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "wrong type",
							Index: 0,
						},
						Station: "station",
					},
					Resource: mcom.ToolResource{
						ResourceID: "A",
						ToolID:     "A",
					},
				},
			},
		}), "STATION_SITE_SUB_TYPE_MISMATCH: illegal site type or sub type, station: station, site name: wrong type, site index: 0")
	}
	{ // illegal site sub type
		assert.EqualError(dm.ToolResourceBindV2(ctx, mcom.ToolResourceBindRequestV2{
			Details: []mcom.ToolBindRequestDetailV2{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "wrong sub type",
							Index: 0,
						},
						Station: "station",
					},
					Resource: mcom.ToolResource{
						ResourceID: "A",
						ToolID:     "A",
					},
				},
			},
		}), "STATION_SITE_SUB_TYPE_MISMATCH: illegal site type or sub type, station: station, site name: wrong sub type, site index: 0")
	}
	{ // site not found
		assert.ErrorIs(dm.ToolResourceBindV2(ctx, mcom.ToolResourceBindRequestV2{
			Details: []mcom.ToolBindRequestDetailV2{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "not found",
							Index: 0,
						},
						Station: "station",
					},
					Resource: mcom.ToolResource{
						ResourceID: "A",
						ToolID:     "A",
					},
				},
			},
		}), mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND})
	}
	// #endregion bad cases

	// #region good cases
	{ // bind tool
		assert.NoError(dm.ToolResourceBindV2(ctx, mcom.ToolResourceBindRequestV2{
			Details: []mcom.ToolBindRequestDetailV2{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "slot0",
							Index: 0,
						},
						Station: "station",
					},
					Resource: mcom.ToolResource{
						ResourceID: "A",
						ToolID:     "A",
					},
				},
			},
		}))

		actualSite, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "slot0",
			SiteIndex: 0,
		})
		assert.NoError(err)
		actualSite.Attributes.LimitHandler = nil
		expectedSite := mcom.GetSiteReply{
			Name:               "slot0",
			Index:              0,
			AdminDepartmentOID: "doid",
			Attributes: mcom.SiteAttributes{
				Type:         sites.Type_SLOT,
				SubType:      sites.SubType_TOOL,
				LimitHandler: nil,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{
					Tool: &models.ToolSite{
						ResourceID:    "A",
						ToolID:        "A",
						InstalledTime: actualSite.Content.Slot.Tool.InstalledTime,
					},
				},
			},
		}
		assert.Equal(expectedSite, actualSite)

		actualTool, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
			ResourceID: "A",
		})
		assert.NoError(err)
		expectedTool := mcom.GetToolResourceReply{
			ToolID: "A",
			BindingSite: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "slot0",
					Index: 0,
				},
				Station: "station",
			},
			CreatedBy: "",
			CreatedAt: actualTool.CreatedAt,
		}
		assert.Equal(expectedTool, actualTool)
	}
	{ // clear
		assert.NoError(dm.ToolResourceBindV2(ctx, mcom.ToolResourceBindRequestV2{
			Details: []mcom.ToolBindRequestDetailV2{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR,
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "slot0",
							Index: 0,
						},
						Station: "station",
					},
				},
			},
		}))

		actualSite, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "slot0",
			SiteIndex: 0,
		})
		assert.NoError(err)
		actualSite.Attributes.LimitHandler = nil
		expectedSite := mcom.GetSiteReply{
			Name:               "slot0",
			Index:              0,
			AdminDepartmentOID: "doid",
			Attributes: mcom.SiteAttributes{
				Type:         sites.Type_SLOT,
				SubType:      sites.SubType_TOOL,
				LimitHandler: nil,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{},
			},
		}
		assert.Equal(expectedSite, actualSite)

		actualTool, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
			ResourceID: "A",
		})
		assert.NoError(err)
		expectedTool := mcom.GetToolResourceReply{
			ToolID:    "A",
			CreatedBy: "",
			CreatedAt: actualTool.CreatedAt,
		}
		assert.Equal(expectedTool, actualTool)
	}
	{ // multiple bind
		assert.NoError(dm.ToolResourceBindV2(ctx, mcom.ToolResourceBindRequestV2{
			Details: []mcom.ToolBindRequestDetailV2{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "slot1",
							Index: 0,
						},
						Station: "station",
					},
					Resource: mcom.ToolResource{
						ResourceID: "B",
						ToolID:     "B",
					},
				},
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.UniqueSite{
						SiteID: models.SiteID{
							Name:  "slot1",
							Index: 1,
						},
						Station: "station",
					},
					Resource: mcom.ToolResource{
						ResourceID: "C",
						ToolID:     "C",
					},
				},
			},
		}))

		actualSite, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "slot1",
			SiteIndex: 0,
		})
		assert.NoError(err)
		actualSite.Attributes.LimitHandler = nil
		expectedSite := mcom.GetSiteReply{
			Name:               "slot1",
			Index:              0,
			AdminDepartmentOID: "doid",
			Attributes: mcom.SiteAttributes{
				Type:         sites.Type_SLOT,
				SubType:      sites.SubType_TOOL,
				LimitHandler: nil,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{
					Tool: &models.ToolSite{
						ResourceID:    "B",
						ToolID:        "B",
						InstalledTime: actualSite.Content.Slot.Tool.InstalledTime,
					},
				},
			},
		}
		assert.Equal(expectedSite, actualSite)

		actualSite2, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "slot1",
			SiteIndex: 1,
		})
		assert.NoError(err)
		actualSite2.Attributes.LimitHandler = nil
		expectedSite2 := mcom.GetSiteReply{
			Name:               "slot1",
			Index:              1,
			AdminDepartmentOID: "doid",
			Attributes: mcom.SiteAttributes{
				Type:         sites.Type_SLOT,
				SubType:      sites.SubType_TOOL,
				LimitHandler: nil,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{
					Tool: &models.ToolSite{
						ResourceID:    "C",
						ToolID:        "C",
						InstalledTime: actualSite2.Content.Slot.Tool.InstalledTime,
					},
				},
			},
		}
		assert.Equal(expectedSite2, actualSite2)

		actualTool, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
			ResourceID: "B",
		})
		assert.NoError(err)
		expectedTool := mcom.GetToolResourceReply{
			ToolID: "B",
			BindingSite: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "slot1",
					Index: 0,
				},
				Station: "station",
			},
			CreatedBy: "",
			CreatedAt: actualTool.CreatedAt,
		}
		assert.Equal(expectedTool, actualTool)

		actualTool2, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
			ResourceID: "C",
		})
		assert.NoError(err)
		expectedTool2 := mcom.GetToolResourceReply{
			ToolID: "C",
			BindingSite: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "slot1",
					Index: 1,
				},
				Station: "station",
			},
			CreatedBy: "",
			CreatedAt: actualTool2.CreatedAt,
		}
		assert.Equal(expectedTool2, actualTool2)
	}
	{ // bind to a site which is not empty
		assert.NoError(dm.ToolResourceBindV2(ctx, mcom.ToolResourceBindRequestV2{
			Details: []mcom.ToolBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "slot1",
						Index: 1,
					},
					Station: "station",
				},
				Resource: mcom.ToolResource{
					ResourceID:  "F",
					ToolID:      "F",
					BindingSite: models.UniqueSite{},
				},
			}},
		}))

		actualTool, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
			ResourceID: "C",
		})
		assert.NoError(err)
		expectedTool := mcom.GetToolResourceReply{
			ToolID:      "C",
			BindingSite: models.UniqueSite{},
			CreatedBy:   "",
			CreatedAt:   actualTool.CreatedAt,
		}
		assert.Equal(expectedTool, actualTool)

		actualTool2, err := dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
			ResourceID: "F",
		})
		assert.NoError(err)
		expectedTool2 := mcom.GetToolResourceReply{
			ToolID: "F",
			BindingSite: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "slot1",
					Index: 1,
				},
				Station: "station",
			},
			CreatedBy: "",
			CreatedAt: actualTool2.CreatedAt,
		}
		assert.Equal(expectedTool2, actualTool2)
	}
	{ // bind with empty tool ID
		const resourceID = "not found"
		assert.NoError(dm.ToolResourceBindV2(ctx, mcom.ToolResourceBindRequestV2{
			Details: []mcom.ToolBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "slot1",
						Index: 1,
					},
					Station: "station",
				},
				Resource: mcom.ToolResource{
					ResourceID:  resourceID,
					BindingSite: models.UniqueSite{},
				},
			}},
		}))

		actual, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "station",
			SiteName:  "slot1",
			SiteIndex: 1,
		})
		assert.NoError(err)

		actual.Attributes.LimitHandler = nil
		expected := mcom.GetSiteReply{
			Name:               "slot1",
			Index:              1,
			AdminDepartmentOID: "doid",
			Attributes: mcom.SiteAttributes{
				Type:         sites.Type_SLOT,
				SubType:      sites.SubType_TOOL,
				LimitHandler: nil,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{
					Tool: &models.ToolSite{
						ResourceID:    resourceID,
						InstalledTime: actual.Content.Slot.Tool.InstalledTime,
					},
				},
			},
		}

		assert.Equal(expected, actual)
	}
	// #endregion good cases
	assert.NoError(cm.Clear())
}

func TestDataManager_ListToolResources(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &models.ToolResource{})
	assert.NoError(cm.Clear())

	now := time.Now()
	patch := monkey.Patch(time.Now, func() time.Time { return now })
	// #region create test data
	testResources := []models.ToolResource{
		{
			ID:     "A",
			ToolID: "A",
		},
		{
			ID:     "B",
			ToolID: "B",
		},
		{
			ID:     "C",
			ToolID: "C",
		},
	}
	assert.NoError(db.Create(testResources).Error)
	// #endregion create test data

	{
		actual, err := dm.ListToolResources(ctx, mcom.ListToolResourcesRequest{
			ResourcesID: []string{"A", "C", "D"},
		})
		assert.NoError(err)

		expected := mcom.ListToolResourcesReply{
			Resources: map[string]mcom.GetToolResourceReply{
				"A": {
					ToolID:    "A",
					CreatedAt: types.ToTimeNano(now),
				},
				"C": {
					ToolID:    "C",
					CreatedAt: types.ToTimeNano(now),
				},
			},
		}

		assert.Equal(expected, actual)
	}
	patch.Unpatch()
	assert.NoError(cm.Clear())
}
