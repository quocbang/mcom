package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models/cloud"
)

// Model definition.
type Model interface {
	schema.Tabler
}

// Function definition.
type Function interface {
	MigrateFunction(db *gorm.DB) error
}

// Trigger definition.
type Trigger interface {
	MigrateTrigger(db *gorm.DB) error
}

// GetModelList returns a list of gorm models.
func GetModelList() []Model {
	return []Model{
		&Token{},
		&Account{},
		&User{},
		&Department{},

		&Site{},
		&SiteContents{},

		&ProductionPlan{},
		&WorkOrder{},
		&Batch{},
		&CollectRecord{},
		&FeedRecord{},

		&Warehouse{},
		&WarehouseStock{},
		&MaterialResource{},
		&ResourceTransportRecord{},

		&Station{},
		&StationGroup{},
		&StationConfiguration{},
		&BindRecords{},

		&Recipe{},
		&RecipeProcessDefinition{},

		&ProductGroup{},

		&CarrierSerial{},
		&Carrier{},

		&SubstitutionMapping{},
		&PackRecord{},

		&ToolResource{},

		&LimitaryHour{},
	}
}

// GetCloudModelList returns a list of gorm models
func GetCloudModelList() []Model {
	return []Model{
		&cloud.Blob{},
	}
}
