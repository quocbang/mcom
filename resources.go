package mcom

import (
	"time"

	"github.com/shopspring/decimal"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// Material definition.
type Material struct {
	Type            string
	ID              string
	Grade           string
	Status          resources.MaterialStatus
	Quantity        decimal.Decimal
	PlannedQuantity decimal.Decimal
	Station         string
	Unit            string
	LotNumber       string
	ProductionTime  time.Time
	ExpiryTime      time.Time
	ResourceID      string
	MinDosage       decimal.Decimal
	Inspections     models.Inspections
	Remark          string
	CarrierID       string
	UpdatedAt       types.TimeNano
	UpdatedBy       string
	CreatedAt       types.TimeNano
	CreatedBy       string
	FeedRecordsID   []string
}

// CreateMaterialResourcesRequest definition.
type CreateMaterialResourcesRequest struct {
	Materials []CreateMaterialResourcesRequestDetail
}

type CreateMaterialResourcesRequestDetail struct {
	Type            string
	ID              string
	Grade           string
	Status          resources.MaterialStatus
	Quantity        decimal.Decimal
	PlannedQuantity decimal.Decimal
	Station         string
	Unit            string
	LotNumber       string
	ProductionTime  time.Time
	ExpiryTime      time.Time
	ResourceID      string
	MinDosage       decimal.Decimal
	Inspections     models.Inspections
	Remark          string
	CarrierID       string
	FeedRecordID    string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreateMaterialResourcesRequest) CheckInsufficiency() error {
	if len(req.Materials) == 0 {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "there is no resource to add",
		}
	}
	for _, resource := range req.Materials {
		if resource.Quantity.LessThan(decimal.Zero) ||
			resource.PlannedQuantity.LessThan(decimal.Zero) ||
			(resource.Quantity.Equal(decimal.Zero) && resource.PlannedQuantity.Equal(decimal.Zero)) {
			return mcomErr.Error{Code: mcomErr.Code_INVALID_NUMBER}
		}
	}
	return nil
}

// CreateMaterialResourcesOptions definition.
type CreateMaterialResourcesOptions struct {
	Warehouse Warehouse
}

// Warehouse definition.
type Warehouse struct {
	ID       string
	Location string
}

func newCreateMaterialResourcesOptions() CreateMaterialResourcesOptions {
	return CreateMaterialResourcesOptions{Warehouse: Warehouse{}}
}

// CreateMaterialResourcesOption definition.
type CreateMaterialResourcesOption func(*CreateMaterialResourcesOptions)

// ParseCreateMaterialResourcesOptions return options for create resource.
func ParseCreateMaterialResourcesOptions(opts []CreateMaterialResourcesOption) CreateMaterialResourcesOptions {
	o := newCreateMaterialResourcesOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// WithStockIn defines resource stock in action.
func WithStockIn(warehouse Warehouse) CreateMaterialResourcesOption {
	return func(o *CreateMaterialResourcesOptions) {
		o.Warehouse = Warehouse{
			ID:       warehouse.ID,
			Location: warehouse.Location,
		}
	}
}

// ListChangeableStatusRequest definition.
type ListChangeableStatusRequest struct {
	MaterialID string
}
