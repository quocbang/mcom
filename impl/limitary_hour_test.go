package impl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

func TestDataManager_CreateLimitaryHour(t *testing.T) {
	const (
		product_type = "CARBON RING"
		min          = 0
		max          = 168
	)
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &models.LimitaryHour{})
	assert.NoError(cm.Clear())

	{ // good case.
		assert.NoError(dm.CreateLimitaryHour(ctx, mcom.CreateLimitaryHourRequest{
			LimitaryHour: []mcom.LimitaryHour{{
				ProductType: product_type,
				LimitaryHour: mcom.LimitaryHourParameter{
					Min: min,
					Max: max,
				},
			},
			},
		}))
		var actual models.LimitaryHour
		assert.NoError(db.Where(`product_type = ?`, "CARBON RING").Take(&actual).Error)
		expected := models.LimitaryHour{
			ProductType: "CARBON RING",
			Min:         0,
			Max:         168,
		}
		assert.Equal(expected, actual)
	}
	{ //product-type already exist
		err := dm.CreateLimitaryHour(ctx, mcom.CreateLimitaryHourRequest{
			LimitaryHour: []mcom.LimitaryHour{{
				ProductType: product_type,
				LimitaryHour: mcom.LimitaryHourParameter{
					Min: min,
					Max: max,
				},
			},
			},
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_LIMITARY_HOUR_ALREADY_EXISTS,
			Details: "product-type already exist",
		}, err)
	}
}

func TestDataManager_GetLimitaryHour(t *testing.T) {
	ctx, dm, db := initializeDB(t)
	defer dm.Close()

	assert := assert.New(t)
	cm := newClearMaster(db, &models.LimitaryHour{})
	assert.NoError(cm.Clear())
	{ // LimitaryHour not found.
		_, err := dm.GetLimitaryHour(ctx, mcom.GetLimitaryHourRequest{
			ProductType: "not found",
		})
		assert.ErrorIs(mcomErr.Error{Code: mcomErr.Code_LIMITARY_HOUR_NOT_FOUND}, err)
	}

	{ // good case
		assert.NoError(dm.CreateLimitaryHour(ctx, mcom.CreateLimitaryHourRequest{
			LimitaryHour: []mcom.LimitaryHour{{
				ProductType: "test",
				LimitaryHour: mcom.LimitaryHourParameter{
					Min: 24,
					Max: 24,
				},
			}},
		}))

		actual, err := dm.GetLimitaryHour(ctx, mcom.GetLimitaryHourRequest{
			ProductType: "test",
		})
		assert.NoError(err)

		expected := mcom.GetLimitaryHourReply{
			LimitaryHour: mcom.LimitaryHourParameter{
				Min: 24,
				Max: 24,
			},
		}
		assert.Equal(expected, actual)
	}
}
