package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// Warehouse defines the warehouse information.
type Warehouse struct {
	ID string `gorm:"type:char(1);primaryKey"`
	// DepartmentID means the warehouse belongs to what department.
	// The value is relative to Department.ID.
	DepartmentID string `gorm:"column:department_id;type:text;not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (Warehouse) TableName() string {
	return "warehouse"
}

// WarehouseStock define stocks in the warehouse.
type WarehouseStock struct {
	// ID is relative to Warehouse.ID.
	ID        string          `gorm:"type:char(1);not null;primaryKey"`
	Location  string          `gorm:"type:varchar(2);not null;primaryKey"`
	ProductID string          `gorm:"not null;primaryKey"`
	Quantity  decimal.Decimal `gorm:"type:numeric(16, 6);not null"`
}

// Describe returns information containing resource id and product type.
func (w WarehouseStock) Describe() string {
	return fmt.Sprintln("WarehouseID: " + w.ID + ",Location: " + w.Location + ",Product ID:" + w.ProductID)
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (WarehouseStock) TableName() string {
	return "warehouse_stock"
}

// MigrateTrigger implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
// Trigger interface.
func (*WarehouseStock) MigrateTrigger(db *gorm.DB) error {
	if err := db.Exec("DROP TRIGGER IF EXISTS delete_non_positive_value_after_update on warehouse_stock;").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE TRIGGER delete_non_positive_value_after_update AFTER UPDATE ON warehouse_stock FOR EACH ROW EXECUTE PROCEDURE delete_non_positive_value();").Error; err != nil {
		return err
	}
	if err := db.Exec("DROP TRIGGER IF EXISTS delete_non_positive_value_after_insert on warehouse_stock;").Error; err != nil {
		return err
	}
	return db.Exec("CREATE TRIGGER delete_non_positive_value_after_insert AFTER INSERT ON warehouse_stock FOR EACH ROW EXECUTE PROCEDURE delete_non_positive_value();").Error
}

// MigrateFunction implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
// Function interface.
func (*WarehouseStock) MigrateFunction(db *gorm.DB) error {
	sql := strings.Join([]string{
		"CREATE OR REPLACE FUNCTION delete_non_positive_value()",
		"	RETURNS trigger AS",
		"$$",
		"BEGIN",
		"	IF NEW.quantity <= 0 THEN",
		"	DELETE FROM",
		"		warehouse_stock",
		"	WHERE",
		"		id = NEW.id",
		"		AND location = NEW.location",
		"		AND product_id = NEW.product_id;",
		"	END IF;",
		"	RETURN NEW;",
		"END;",
		"$$ LANGUAGE plpgsql;",
	}, "\n")
	return db.Exec(sql).Error
}

// MaterialResource table definition.
type MaterialResource struct {
	// OID is automatically generated as an UUID before creating.
	OID string `gorm:"column:oid;type:uuid;primaryKey"`

	// resource id
	ID string `gorm:"type:text;not null;uniqueIndex:idx_material_resource_id"`
	// ProductID is relative to RecipeProcessDefinition.OutputProduct.ID.
	ProductID string `gorm:"not null"`
	// ProductType is relative to RecipeProcessDefinition.OutputProduct.Type.
	ProductType string                   `gorm:"not null;uniqueIndex:idx_material_resource_id"`
	Quantity    decimal.Decimal          `gorm:"type:numeric(16, 6);not null"`
	Status      resources.MaterialStatus `gorm:"default:1;not null"`
	ExpiryTime  types.TimeNano           `gorm:"type:bigint;not null"`

	Info MaterialInfo `gorm:"type:jsonb;not null"`

	// WarehouseID is relative to Warehouse.ID.
	WarehouseID       string `gorm:"type:varchar(1);not null"`
	WarehouseLocation string `gorm:"type:varchar(2);not null"`

	FeedRecordsID pq.StringArray `gorm:"type:text[];not null"`

	// the station that the resource was produced by
	Station string `gorm:"type:varchar(32);not null"`

	// UpdatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	UpdatedAt types.TimeNano `gorm:"type:bigint;autoUpdateTime:nano;not null"`
	UpdatedBy string         `gorm:"type:text;not null"`
	CreatedAt types.TimeNano `gorm:"autoCreateTime:nano;not null;index"`
	CreatedBy string         `gorm:"type:text;not null"`
}

// HasStockedIn returns true if resource has been stocked in.
func (mr MaterialResource) HasStockedIn() bool {
	return mr.WarehouseID != "" && mr.WarehouseLocation != ""
}

// Describe returns information containing resource id and product type.
func (r MaterialResource) Describe() string {
	return fmt.Sprintln("Resource ID: " + r.ID + ", Product Type: " + r.ProductType)
}

// BeforeCreate gorm hook.
func (mr *MaterialResource) BeforeCreate(*gorm.DB) error {
	mr.OID = uuid.NewV4().String()
	return nil
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (MaterialResource) TableName() string {
	return "material_resource"
}

// MaterialInfo definition.
type MaterialInfo struct {
	Grade           string          `json:"grade"`
	Unit            string          `json:"unit"`
	LotNumber       string          `json:"lot_number"`
	ProductionTime  time.Time       `json:"production_time"`
	MinDosage       decimal.Decimal `json:"min_dosage"`
	Inspections     Inspections     `json:"inspections"`
	PlannedQuantity decimal.Decimal `json:"planned_quantity"`
	Remark          string          `json:"remark"`
}

type Inspections []Inspection

func (irs *Inspections) Split(ids []int) (Inspections, error) {
	var old, new Inspections
	targets := make(map[int]struct{})
	for _, elem := range ids {
		targets[elem] = struct{}{}
	}
	for _, elem := range *irs {
		if _, ok := targets[elem.ID]; ok {
			new = append(new, elem)
		} else {
			old = append(old, elem)
		}
	}

	if len(new) != len(targets) {
		return nil, fmt.Errorf("inspection remark number not found")
	}
	if len(new) == 0 || len(old) == 0 {
		return nil, fmt.Errorf("the number of inspection remark must be > 0")
	}

	*irs = old
	return new, nil
}

type Inspection struct {
	ID     int    `json:"id"`
	Remark string `json:"remark"`
}

// Scan implements database/sql Scanner interface.
func (c *MaterialInfo) Scan(src interface{}) error {
	return ScanJSON(src, c)
}

// Value implements database/sql/driver Valuer interface.
func (c MaterialInfo) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// ResourceTransportRecord definition.
type ResourceTransportRecord struct {
	// ID is a serial number, it is automatically generated when creating.
	ID int64 `gorm:"type:bigserial;primaryKey"`

	// OldWarehouseID are relative to Warehouse.ID.
	OldWarehouseID string `gorm:"type:varchar(1);not null"`
	// OldLocation are relative to Warehouse.Location.
	OldLocation string `gorm:"type:varchar(2);not null"`
	// NewWarehouseID are relative to Warehouse.ID.
	NewWarehouseID string `gorm:"type:varchar(1);not null"`
	// NewLocation are relative to Warehouse.Location.
	NewLocation string `gorm:"type:varchar(2);not null"`
	// ResourceID is relative to Resource.ID.
	ResourceID string `gorm:"type:text;not null"`
	Note       string `gorm:"type:varchar(50);default:'';not null"`

	// CreatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	CreatedAt types.TimeNano `gorm:"autoCreateTime:nano;not null"`
	CreatedBy string         `gorm:"type:text;not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (ResourceTransportRecord) TableName() string {
	return "resource_transport_record"
}
