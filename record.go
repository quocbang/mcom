package mcom

import (
	"time"

	"github.com/shopspring/decimal"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// GetCollectRecordRequest definition.
type GetCollectRecordRequest struct {
	WorkOrder string
	Sequence  int16
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetCollectRecordRequest) CheckInsufficiency() error {
	if req.WorkOrder == "" || req.Sequence == 0 {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// ListRecordsRequest definition.
type ListRecordsRequest struct {
	Date         time.Time
	DepartmentID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListRecordsRequest) CheckInsufficiency() error {
	if req.Date.IsZero() || req.DepartmentID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// FeedMaterial definition.
type FeedMaterial struct {
	// material id.
	ID string
	// material grade.
	Grade string

	ResourceID  string
	ProductType string
	Status      resources.MaterialStatus
	ExpiryTime  types.TimeNano

	Quantity decimal.Decimal
	Site     models.SiteID
}

// SiteID definition.
type SiteID struct {
	Name  string
	Index int16
}

// FeedRecord definition.
type FeedRecord struct {
	WorkOrder string
	Batch     int32

	RecipeID    string
	ProcessName string
	ProcessType string
	StationID   string

	Materials []FeedMaterial
}

type GetCollectRecordReply CollectRecord

// CollectRecord definition.
type CollectRecord struct {
	ResourceID string
	LotNumber  string
	WorkOrder  string
	RecipeID   string
	ProductID  string
	BatchCount int16
	Quantity   decimal.Decimal
	OperatorID string
}

// ListFeedRecordReply definition.
type ListFeedRecordReply []FeedRecord

// ListCollectRecordsReply definition.
type ListCollectRecordsReply []CollectRecord

// CreateCollectRecordRequest definition.
type CreateCollectRecordRequest struct {
	WorkOrder   string
	Sequence    int16
	LotNumber   string
	Station     string
	ResourceOID string
	Quantity    decimal.Decimal
	BatchCount  int16
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreateCollectRecordRequest) CheckInsufficiency() error {
	if req.WorkOrder == "" || req.Sequence == 0 || req.LotNumber == "" ||
		req.Station == "" || req.ResourceOID == "" {
		return mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}
	}
	if req.Quantity.LessThanOrEqual(decimal.Zero) {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "the Quantity is <= 0",
		}
	}
	return nil
}

// CreatedMaterialResource definition.
type CreatedMaterialResource struct {
	ID  string
	OID string
}

// CreateMaterialResourcesReply definition.
type CreateMaterialResourcesReply []CreatedMaterialResource
