package models

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	pbWorkOrder "gitlab.kenda.com.tw/kenda/commons/v2/proto/golang/mes/v2/workorder"

	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
	"gitlab.kenda.com.tw/kenda/mcom/utils/workorder"
)

// WorkOrderInformation definition.
type WorkOrderInformation struct {
	BatchQuantityDetails
	Unit        string                `json:"unit"`
	Abnormality workorder.Abnormality `json:"abnormality"`
	Parent      string                `json:"parent"`

	ProductID   string `json:"product_id"`
	ProductType string `json:"product_type"`
}

type BatchQuantityDetails struct {
	BatchQuantityType workorder.BatchSize `json:"batch_quantity_type"`
	// QuantityForBatches are an estimated quantity for batches. Index 0 for batch.
	// no.1, Index 1 for batch no.2, and so on.
	//
	// BatchQuantityType => BatchQuantityType_PER_BATCH
	// todo rename json name
	QuantityForBatches []decimal.Decimal `json:"batches,omitempty"`
	// BatchQuantityType => BatchQuantityType_FIXED
	FixedQuantity *FixedQuantity `json:"fixed_quantity,omitempty"`
	// BatchQuantityType => BatchQuantityType_PLAN
	PlanQuantity *PlanQuantity `json:"plan_quantity,omitempty"`
}

type FixedQuantity struct {
	BatchCount   uint            `json:"batch_count,omitempty"`
	PlanQuantity decimal.Decimal `json:"plan_quantity,omitempty"`
}

type PlanQuantity struct {
	BatchCount   uint            `json:"batch_count,omitempty"`
	PlanQuantity decimal.Decimal `json:"plan_quantity,omitempty"`
}

// Scan implements database/sql Scanner interface.
func (info *WorkOrderInformation) Scan(src interface{}) error {
	return ScanJSON(src, info)
}

// Value implements database/sql/driver Valuer interface.
func (info WorkOrderInformation) Value() (driver.Value, error) {
	return json.Marshal(info)
}

// WorkOrder definition.
type WorkOrder struct {
	// ID is automatically generated before creating.
	ID string `gorm:"type:char(20);primaryKey"`

	// RecipeID is relative to Recipe.ID.
	RecipeID string `gorm:"type:text;not null"`

	// ProcessOID is relative to RecipeProcessDefinition.OID.
	ProcessOID string `gorm:"column:process_oid;type:uuid;not null"`

	ProcessName string `gorm:"type:text;not null"`

	ProcessType string `gorm:"type:text;not null"`

	// DepartmentID the work order is for. The value is relative to Department.ID.
	DepartmentID string `gorm:"column:department_id;type:text;not null;index:idx_work_order_department"`

	Status pbWorkOrder.Status `gorm:"default:0;not null"`

	// Station is where the work order is executed. The value is relative to Station.ID.
	Station string `gorm:"type:varchar(32);default:'';not null;index:idx_work_order_station"`

	// ReservedDate is the date when to produce.
	ReservedDate     time.Time `gorm:"type:date;not null;index:idx_work_order_date"`
	ReservedSequence int32     `gorm:"type:integer;default:0;not null"`

	Information WorkOrderInformation `gorm:"type:jsonb;default:'{}';not null"`

	// UpdatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	UpdatedAt types.TimeNano `gorm:"autoUpdateTime:nano;not null"`
	UpdatedBy string         `gorm:"type:text;not null"`
	// CreatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	CreatedAt types.TimeNano `gorm:"autoCreateTime:nano;not null"`
	CreatedBy string         `gorm:"type:text;not null"`
}

// BeforeCreate gorm hook.
func (wo *WorkOrder) BeforeCreate(*gorm.DB) error {
	wo.ID = strings.ToUpper(xid.New().String())
	return nil
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (WorkOrder) TableName() string {
	return "work_order"
}

type FeedRecord struct {
	ID string `gorm:"type:text;not null;primaryKey"`

	// OperatorID is who feeds.
	OperatorID string      `gorm:"not null"`
	Materials  FeedDetails `gorm:"type:jsonb;default:'[]';not null"`

	// Time is when to feed.
	Time time.Time `gorm:"not null"`
}

// Model implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface
func (FeedRecord) Model() {}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface
func (FeedRecord) TableName() string {
	return "feed_record"
}

type FeedDetails []FeedDetail

// Scan implements database/sql Scanner interface.
func (details *FeedDetails) Scan(src interface{}) error {
	return ScanJSON(src, details)
}

// Value implements database/sql/driver Valuer interface.
func (details FeedDetails) Value() (driver.Value, error) {
	return json.Marshal(details)
}

// FeedDetail definition.
type FeedDetail struct {
	Resources []FeedResource `json:"feed_resources"`

	// The site that materials are fed from.
	UniqueSite
}

type FeedResource struct {
	// ResourceID is the ID of the resource which is fed.
	// It is relative to MaterialResource.ID.
	ResourceID  string                   `json:"resource_id"`
	ProductID   string                   `json:"product_id"`
	ProductType string                   `json:"product_type"`
	Grade       string                   `json:"grade"`
	Status      resources.MaterialStatus `json:"status"`
	ExpiryTime  time.Time                `json:"expiry_time"`

	// Quantity is how many/much the operator feeds.
	Quantity decimal.Decimal `json:"quantity"`
}

// Batch is a batch table that divide the operation to the work order into smaller.
// operations.
type Batch struct {
	// WorkOrder is relative to WorkOrder.ID.
	WorkOrder string `gorm:"type:char(20);primaryKey"`

	// Number is batch number, it starts from 1.
	Number int16 `gorm:"primaryKey;check:number > 0"`

	Status pbWorkOrder.BatchStatus `gorm:"default:0;not null"`

	// fed record id
	RecordsID pq.StringArray `gorm:"type:text[];default array[]::text[];not null"`

	Note string `gorm:"type:varchar(50);default:'';not null"`

	// UpdatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	UpdatedAt types.TimeNano `gorm:"autoUpdateTime:nano;not null"`
	UpdatedBy string         `gorm:"type:text;not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (Batch) TableName() string {
	return "batch"
}

// CollectRecordDetail definition.
type CollectRecordDetail struct {
	// OperatorID is who collects products.
	OperatorID string `json:"operator_id"`

	// Quantity is how many product the operator collects.
	Quantity decimal.Decimal `json:"quantity"`

	// BatchCount is the total of collected batch.
	BatchCount int16 `json:"batch_count"`
}

// Scan implements database/sql Scanner interface.
func (info *CollectRecordDetail) Scan(src interface{}) error {
	return ScanJSON(src, info)
}

// Value implements database/sql/driver Valuer interface.
func (info CollectRecordDetail) Value() (driver.Value, error) {
	return json.Marshal(info)
}

// CollectRecord definition.
type CollectRecord struct {
	// WorkOrder is relative to WorkOrder.ID.
	WorkOrder string `gorm:"type:char(20);primaryKey"`
	Sequence  int16  `gorm:"primaryKey"`
	LotNumber string `gorm:"type:varchar(8);"`

	// Station is relative to Station.ID.
	Station string `gorm:"type:varchar(32);not null"`

	// ResourceOID is relative to Resource.OID.
	ResourceOID string              `gorm:"column:resource_oid;type:uuid;not null"`
	Detail      CollectRecordDetail `gorm:"type:jsonb;default:'{}';not null"`

	// CreatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	CreatedAt types.TimeNano `gorm:"autoCreateTime:nano;not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (CollectRecord) TableName() string {
	return "collect_record"
}
