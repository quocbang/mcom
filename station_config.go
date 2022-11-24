package mcom

import (
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

type SetStationConfigurationRequest struct {
	StationID           string `validate:"required"`
	SplitFeedAndCollect bool

	Feed    StationFeedConfigs
	Collect StationCollectConfigs
}

type StationFeedConfigs struct {
	ProductTypes         []string
	NeedMaterialResource bool
	QuantitySource       stations.FeedQuantitySource
	OperatorSites        []models.UniqueSite
}

type StationCollectConfigs struct {
	NeedCollectResource bool
	NeedCarrierResource bool
	OperatorSites       []models.UniqueSite
	QuantitySource      stations.CollectQuantitySource
	DefaultQuantity     decimal.Decimal
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req SetStationConfigurationRequest) CheckInsufficiency() error {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

type GetStationConfigurationRequest struct {
	StationID string `validate:"required"`
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetStationConfigurationRequest) CheckInsufficiency() error {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

// GetStationConfigurationReply definition.
type GetStationConfigurationReply struct {
	SplitFeedAndCollect bool

	Feed StationFeedConfigs

	Collect StationCollectConfigs

	UpdatedAt types.TimeNano
	UpdatedBy string
}
