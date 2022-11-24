package models

import (
	"github.com/lib/pq"
)

// ProductGroup definition.
type ProductGroup struct {
	// ProductID is relative to RecipeProcessDefinition.OutputProduct.ID.
	ProductID string `gorm:"type:text;primaryKey"`
	// ProductType is relative to RecipeProcessDefinition.OutputProduct.Type.
	ProductType string `gorm:"not null;primaryKey"`
	// DepartmentID is relative to Department.ID.
	DepartmentID string `gorm:"column:department_id;type:text;not null;primaryKey"`
	// Children the value of the element is relative to RecipeProcessDefinition.OutputProduct.ID.
	Children pq.StringArray `gorm:"type:text[];default:'{}';not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (ProductGroup) TableName() string {
	return "product_group"
}
