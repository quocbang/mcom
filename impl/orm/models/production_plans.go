package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// ProductionPlanProduct definition.
type ProductionPlanProduct struct {
	// ID is relative to RecipeProcessDefinition.OutputProduct.ID.
	ID string `gorm:"column:product_id;type:text;not null;uniqueIndex:idx_plan_production"`
	// Type is relative to RecipeProcessDefinition.OutputProduct.Type.
	Type string `gorm:"column:product_type;type:text;not null;uniqueIndex:idx_plan_production"`
}

// ProductionPlan is a table defines plan production for departments.
type ProductionPlan struct {
	// OID is automatically generated as an UUID before creating.
	OID string `gorm:"column:oid;type:uuid;primaryKey"`

	// DepartmentID is relative to Department.ID.
	DepartmentID string    `gorm:"column:department_id;type:text;not null;uniqueIndex:idx_plan_production"`
	PlanDate     time.Time `gorm:"type:date;not null;uniqueIndex:idx_plan_production"`

	ProductionPlanProduct

	Quantity decimal.Decimal `gorm:"type:numeric(16, 6);not null"`

	// UpdatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	UpdatedAt types.TimeNano `gorm:"autoUpdateTime:nano;not null"`
	UpdatedBy string         `gorm:"type:text;not null"`
	// CreatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	CreatedAt types.TimeNano `gorm:"autoCreateTime:nano;not null"`
	CreatedBy string         `gorm:"type:text;not null"`
}

// BeforeCreate gorm hook.
func (pp *ProductionPlan) BeforeCreate(*gorm.DB) error {
	pp.OID = uuid.NewV4().String()
	return nil
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" models interface.
func (ProductionPlan) TableName() string {
	return "production_plan"
}
