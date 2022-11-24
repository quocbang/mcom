package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

func (dm *DataManager) SetStationConfiguration(ctx context.Context, req mcom.SetStationConfigurationRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	return session.setStationConfiguration(req, commonsCtx.UserID(ctx))
}

func (session *session) setStationConfiguration(req mcom.SetStationConfigurationRequest, updater string) error {
	assignments := clause.Assignments(map[string]interface{}{
		"production": models.StationConfigProductionSetting{
			ProductTypes: req.Feed.ProductTypes,
		},
		"ui": models.StationConfigUISetting{
			SplitFeedAndCollect:     req.SplitFeedAndCollect,
			NeedMaterialResource:    req.Feed.NeedMaterialResource,
			FeedQuantitySource:      req.Feed.QuantitySource,
			FeedingOperatorSites:    req.Feed.OperatorSites,
			NeedCarrierResource:     req.Collect.NeedCarrierResource,
			NeedCollectResource:     req.Collect.NeedCollectResource,
			CollectQuantitySource:   req.Collect.QuantitySource,
			DefaultCollectQuantity:  req.Collect.DefaultQuantity,
			CollectingOperatorSites: req.Collect.OperatorSites,
		},
		"updated_by": updater,
	})

	config := models.StationConfiguration{
		StationID: req.StationID,
		Production: models.StationConfigProductionSetting{
			ProductTypes: req.Feed.ProductTypes,
		},
		UI: models.StationConfigUISetting{
			SplitFeedAndCollect:     req.SplitFeedAndCollect,
			NeedMaterialResource:    req.Feed.NeedMaterialResource,
			FeedQuantitySource:      req.Feed.QuantitySource,
			FeedingOperatorSites:    req.Feed.OperatorSites,
			NeedCarrierResource:     req.Collect.NeedCarrierResource,
			NeedCollectResource:     req.Collect.NeedCollectResource,
			CollectQuantitySource:   req.Collect.QuantitySource,
			DefaultCollectQuantity:  req.Collect.DefaultQuantity,
			CollectingOperatorSites: req.Collect.OperatorSites,
		},
		UpdatedBy: updater,
	}

	return session.db.Model(&models.StationConfiguration{}).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "station_id"}},
			DoUpdates: assignments,
		}).Create(&config).Error
}

func (dm *DataManager) GetStationConfiguration(ctx context.Context, req mcom.GetStationConfigurationRequest) (mcom.GetStationConfigurationReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetStationConfigurationReply{}, err
	}

	session := dm.newSession(ctx)
	return session.getStationConfiguration(req)
}

func (session session) getStationConfiguration(req mcom.GetStationConfigurationRequest) (mcom.GetStationConfigurationReply, error) {
	var configs models.StationConfiguration
	if err := session.db.Where(models.StationConfiguration{StationID: req.StationID}).Take(&configs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcom.GetStationConfigurationReply{}, nil
		}
		return mcom.GetStationConfigurationReply{}, err
	}
	return mcom.GetStationConfigurationReply{
		Feed: mcom.StationFeedConfigs{
			ProductTypes:         configs.Production.ProductTypes,
			NeedMaterialResource: configs.UI.NeedMaterialResource,
			QuantitySource:       configs.UI.FeedQuantitySource,
			OperatorSites:        configs.UI.FeedingOperatorSites,
		},
		Collect: mcom.StationCollectConfigs{
			NeedCarrierResource: configs.UI.NeedCarrierResource,
			NeedCollectResource: configs.UI.NeedCollectResource,
			QuantitySource:      configs.UI.CollectQuantitySource,
			DefaultQuantity:     configs.UI.DefaultCollectQuantity,
			OperatorSites:       configs.UI.CollectingOperatorSites,
		},
		SplitFeedAndCollect: configs.UI.SplitFeedAndCollect,
		UpdatedAt:           configs.UpdatedAt,
		UpdatedBy:           configs.UpdatedBy,
	}, nil
}
