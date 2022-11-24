package mcom

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
)

func Test_NewSiteAttributes(t *testing.T) {
	assert := assert.New(t)
	{
		param := models.SiteAttributes{
			Type:       sites.Type_SLOT,
			SubType:    sites.SubType_MATERIAL,
			Limitation: []string{"material"},
		}
		actual := NewSiteAttributes(param)
		expected := SiteAttributes{
			Type:    sites.Type_SLOT,
			SubType: sites.SubType_MATERIAL,
			LimitHandler: func(productID string) error {
				if productID != "material" {
					return mcomErr.Error{
						Code:    mcomErr.Code_PRODUCT_ID_MISMATCH,
						Details: "product id: " + productID,
					}
				}
				return nil
			},
		}

		assert.Equal(expected.Type, actual.Type)
		assert.Equal(expected.SubType, actual.SubType)
		assert.NoError(actual.LimitHandler("material"))
		assert.Equal(expected.LimitHandler("not found"), actual.LimitHandler("not found"))
	}
	{ // case : no limitation.
		param := models.SiteAttributes{
			Type:       sites.Type_SLOT,
			SubType:    sites.SubType_MATERIAL,
			Limitation: []string{},
		}
		actual := NewSiteAttributes(param)
		expected := SiteAttributes{
			Type:    sites.Type_SLOT,
			SubType: sites.SubType_MATERIAL,
			LimitHandler: func(productID string) error {
				return nil
			},
		}
		assert.Equal(expected.LimitHandler(""), actual.LimitHandler("")) //always nil
		assert.Nil(actual.LimitHandler(""))
	}

}

func TestCreateStationRequest_Correct(t *testing.T) {
	assert := assert.New(t)
	req := CreateStationRequest{
		ID: "A",
		Sites: []SiteInformation{
			{
				Station: "",
				Name:    "a",
			},
			{
				Station: "B",
				Name:    "b",
			},
		},
	}
	req.Correct()

	assert.Equal([]SiteInformation{
		{
			Station: "A",
			Name:    "a",
		},
		{
			Station: "B",
			Name:    "b",
		},
	}, req.Sites)
}

func TestUpdateStationRequest_Correct(t *testing.T) {
	assert := assert.New(t)
	req := UpdateStationRequest{
		ID: "A",
		Sites: []UpdateStationSite{
			{
				ActionMode: sites.ActionType_ADD,
				Information: SiteInformation{
					Station: "",
					Name:    "a",
				},
			},
			{
				ActionMode: sites.ActionType_REMOVE,
				Information: SiteInformation{
					Station: "B",
					Name:    "b",
				},
			},
		},
	}
	req.Correct()

	assert.Equal([]UpdateStationSite{
		{
			ActionMode: sites.ActionType_ADD,
			Information: SiteInformation{
				Station: "A",
				Name:    "a",
			},
		},
		{
			ActionMode: sites.ActionType_REMOVE,
			Information: SiteInformation{
				Station: "B",
				Name:    "b",
			},
		},
	}, req.Sites)
}
