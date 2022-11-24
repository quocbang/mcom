package impl

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
)

func TestDataManager_Set_Get_StationConfiguration(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &models.StationConfiguration{})
	assert.NoError(cm.Clear())
	{ // create a config.
		assert.NoError(dm.SetStationConfiguration(ctx, mcom.SetStationConfigurationRequest{
			StationID: "station",
			Feed: mcom.StationFeedConfigs{
				ProductTypes:         []string{"A"},
				NeedMaterialResource: false,
				QuantitySource:       stations.FeedQuantitySource_FROM_RECIPE,
			},
			Collect: mcom.StationCollectConfigs{
				NeedCarrierResource: true,
				NeedCollectResource: false,
				QuantitySource:      stations.CollectQuantitySource_FROM_STATION_CONFIGS,
				DefaultQuantity:     decimal.NewFromInt(15),
			},
		}))

		actual, err := dm.GetStationConfiguration(ctx, mcom.GetStationConfigurationRequest{StationID: "station"})
		assert.NoError(err)
		expected := mcom.GetStationConfigurationReply{
			Feed: mcom.StationFeedConfigs{
				ProductTypes:         []string{"A"},
				NeedMaterialResource: false,
				QuantitySource:       stations.FeedQuantitySource_FROM_RECIPE,
			},
			Collect: mcom.StationCollectConfigs{
				NeedCarrierResource: true,
				NeedCollectResource: false,
				QuantitySource:      stations.CollectQuantitySource_FROM_STATION_CONFIGS,
				DefaultQuantity:     decimal.NewFromInt(15),
			},
			UpdatedAt: actual.UpdatedAt,
		}
		assert.Equal(expected, actual)
	}
	{ // update a config.
		assert.NoError(dm.SetStationConfiguration(ctx, mcom.SetStationConfigurationRequest{
			StationID:           "station",
			SplitFeedAndCollect: true,
			Feed: mcom.StationFeedConfigs{
				ProductTypes:         []string{"B", "C"},
				NeedMaterialResource: true,
				OperatorSites: []models.UniqueSite{{
					SiteID: models.SiteID{
						Name: "site",
					},
					Station: "station",
				}},
			},
			Collect: mcom.StationCollectConfigs{
				NeedCarrierResource: false,
				NeedCollectResource: true,
				QuantitySource:      stations.CollectQuantitySource_FROM_STATION_PARAMS,
				DefaultQuantity:     decimal.NewFromInt(50),
			},
		}))

		actual, err := dm.GetStationConfiguration(ctx, mcom.GetStationConfigurationRequest{StationID: "station"})
		assert.NoError(err)
		expected := mcom.GetStationConfigurationReply{
			SplitFeedAndCollect: true,

			Feed: mcom.StationFeedConfigs{
				ProductTypes:         []string{"B", "C"},
				NeedMaterialResource: true,
				OperatorSites: []models.UniqueSite{{
					SiteID: models.SiteID{
						Name: "site",
					},
					Station: "station",
				}},
			},
			Collect: mcom.StationCollectConfigs{
				NeedCarrierResource: false,
				NeedCollectResource: true,
				QuantitySource:      stations.CollectQuantitySource_FROM_STATION_PARAMS,
				DefaultQuantity:     decimal.NewFromInt(50),
			},

			UpdatedAt: actual.UpdatedAt,
		}
		assert.Equal(expected, actual)
	}
	{ // record not found : no error.
		rep, err := dm.GetStationConfiguration(ctx, mcom.GetStationConfigurationRequest{StationID: "not_found"})
		assert.NoError(err)
		assert.Empty(rep)
	}
	assert.NoError(cm.Clear())
}
