package impl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

func TestDataManager_BindRecord(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, models.BindRecords{})
	assert.NoError(cm.Clear())

	tx := dm.(*DataManager).beginTx(ctx)
	assert.NoError(tx.addBindRecord(mcom.MaterialResourceBindRequest{
		Station: "station",
		Details: []mcom.MaterialBindRequestDetail{{
			Site: models.SiteID{
				Name:  "site",
				Index: 0,
			},
			Resources: []mcom.BindMaterialResource{
				{
					ResourceID:  "A",
					ProductType: "A",
				}, {
					ResourceID:  "B",
					ProductType: "B",
				}},
		}},
	}))

	assert.NoError(tx.addBindRecord(mcom.MaterialResourceBindRequest{
		Station: "station",
		Details: []mcom.MaterialBindRequestDetail{{
			Site: models.SiteID{
				Name:  "site",
				Index: 0,
			},
			Resources: []mcom.BindMaterialResource{
				{
					ResourceID:  "F",
					ProductType: "F",
				}},
		}},
	}))
	assert.NoError(tx.Commit())
	{ // no such site.
		assert.ErrorIs(dm.BindRecordsCheck(ctx, mcom.BindRecordsCheckRequest{
			Site: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "not_found",
					Index: 0,
				},
				Station: "station",
			},
			Resources: []models.UniqueMaterialResource{{
				ResourceID:  "A",
				ProductType: "A",
			}, {
				ResourceID:  "B",
				ProductType: "B",
			}},
		}), mcomErr.Error{Code: mcomErr.Code_STATION_SITE_BIND_RECORD_NOT_FOUND, Details: "no records in the specified site"})
	}
	{ // record not found.
		assert.ErrorIs(dm.BindRecordsCheck(ctx, mcom.BindRecordsCheckRequest{
			Site: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "site",
					Index: 0,
				},
				Station: "station",
			},
			Resources: []models.UniqueMaterialResource{{
				ResourceID:  "C",
				ProductType: "C",
			}, {
				ResourceID:  "B",
				ProductType: "B",
			}},
		}), mcomErr.Error{
			Code:    mcomErr.Code_STATION_SITE_BIND_RECORD_NOT_FOUND,
			Details: "not found resources: resource id:C, product type:C",
		})
	}
	{ // records not found.
		assert.ErrorIs(dm.BindRecordsCheck(ctx, mcom.BindRecordsCheckRequest{
			Site: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "site",
					Index: 0,
				},
				Station: "station",
			},
			Resources: []models.UniqueMaterialResource{{
				ResourceID:  "C",
				ProductType: "C",
			}, {
				ResourceID:  "D",
				ProductType: "D",
			}},
		}), mcomErr.Error{
			Code:    mcomErr.Code_STATION_SITE_BIND_RECORD_NOT_FOUND,
			Details: "not found resources: resource id:C, product type:C;resource id:D, product type:D",
		})
	}
	{ // good case.
		assert.NoError(dm.BindRecordsCheck(ctx, mcom.BindRecordsCheckRequest{
			Site: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "site",
					Index: 0,
				},
				Station: "station",
			},
			Resources: []models.UniqueMaterialResource{{
				ResourceID:  "A",
				ProductType: "A",
			}, {
				ResourceID:  "B",
				ProductType: "B",
			}, {
				ResourceID:  "F",
				ProductType: "F",
			}},
		}))
	}
}
