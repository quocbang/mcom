package cloud

import (
	"github.com/lib/pq"

	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

type Blob struct {
	ContainerName string         `gorm:"primaryKey;not null"`
	BlobURI       string         `gorm:"primaryKey;not null"`
	Resources     pq.StringArray `gorm:"type:text[];not null"`
	Station       string         `gorm:"not null"`
	DateTime      types.TimeNano `gorm:"not null"`
}

// Model implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface
func (Blob) Model() {}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface
func (Blob) TableName() string {
	return "blob"
}
