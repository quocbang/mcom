package models

import (
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

type ToolResource struct {
	// resource id
	ID          string         `gorm:"primaryKey"`
	ToolID      string         `gorm:"not null"`
	BindingSite UniqueSite     `gorm:"column:binding_site;type:jsonb;not null"`
	UpdatedAt   types.TimeNano `gorm:"autoUpdateTime:nano;not null"`
	UpdatedBy   string         `gorm:"not null"`
	CreatedAt   types.TimeNano `gorm:"autoCreateTime:nano;not null"`
	CreatedBy   string         `gorm:"not null"`
}

// Model implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (*ToolResource) Model() {}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (*ToolResource) TableName() string {
	return "tool_resource"
}

func (t *ToolResource) BindTo(site UniqueSite) {
	t.BindingSite = site
}

func (t *ToolResource) Unbind() {
	t.BindingSite = UniqueSite{}
}
