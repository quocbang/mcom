package models

import (
	"github.com/shopspring/decimal"
)

// PackRecord definition.
type PackRecord struct {
	SerialNumber int64           `gorm:"type:bigserial;primaryKey"`
	Packing      string          `gorm:"type:text;default:'';not null"`
	PackNumber   int             `gorm:"default:0;not null"`
	Quantity     int             `gorm:"default:0;not null"`
	ActualWeight decimal.Decimal `gorm:"type:numeric(14,6);default:0;not null"`
	// The station where the pack was weighed.
	Station   string `gorm:"type:varchar(32);default:'';not null"`
	CreatedAt int64  `gorm:"not null"`
	CreatedBy string `gorm:"type:text;not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (PackRecord) TableName() string {
	return "pack_record"
}
