package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/shopspring/decimal"

	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

type StationConfiguration struct {
	StationID  string                         `gorm:"type:varchar(32);not null;primaryKey"`
	Production StationConfigProductionSetting `gorm:"type:jsonb;not null"`
	UI         StationConfigUISetting         `gorm:"type:jsonb;not null"`
	UpdatedAt  types.TimeNano                 `gorm:"autoUpdateTime:nano;not null"`
	UpdatedBy  string                         `gorm:"type:text;not null"`
}

// Model implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (*StationConfiguration) Model() {}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (*StationConfiguration) TableName() string {
	return "station_configuration"
}

type StationConfigUISetting struct {
	SplitFeedAndCollect bool `json:"split_feed_and_collect"`

	FeedQuantitySource   stations.FeedQuantitySource `json:"feed_quantity_source"`
	NeedMaterialResource bool                        `json:"need_material_resource"`
	FeedingOperatorSites []UniqueSite                `json:"feeding_operator_sites"`

	CollectQuantitySource   stations.CollectQuantitySource `json:"collect_quantity_source"`
	DefaultCollectQuantity  decimal.Decimal                `json:"default_collect_quantity"`
	NeedCollectResource     bool                           `json:"need_collect_resource"`
	NeedCarrierResource     bool                           `json:"need_carrier_resource"`
	CollectingOperatorSites []UniqueSite                   `json:"collecting_operator_sites"`
}

// Scan implements database/sql Scanner interface.
func (info *StationConfigUISetting) Scan(src interface{}) error {
	return ScanJSON(src, info)
}

// Value implements database/sql/driver Valuer interface.
func (info StationConfigUISetting) Value() (driver.Value, error) {
	return json.Marshal(info)
}

type StationConfigProductionSetting struct {
	// ProductTypes that the station is required when binding or feeding in production.
	// It is a hint to let MES know what resources are being bound or fed if there are many product identities for this resource id.
	ProductTypes []string
}

// Scan implements database/sql Scanner interface.
func (info *StationConfigProductionSetting) Scan(src interface{}) error {
	return ScanJSON(src, info)
}

// Value implements database/sql/driver Valuer interface.
func (info StationConfigProductionSetting) Value() (driver.Value, error) {
	return json.Marshal(info)
}
