package mcom

import (
	"time"

	"github.com/shopspring/decimal"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
)

// CreateProductionPlanRequest definition.
type CreateProductionPlanRequest struct {
	Date    time.Time
	Product Product

	DepartmentOID string
	Quantity      decimal.Decimal
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreateProductionPlanRequest) CheckInsufficiency() error {
	if req.Date.IsZero() ||
		req.Product.ID == "" ||
		req.Product.Type == "" ||
		req.DepartmentOID == "" {
		return mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}
	}

	if !req.Quantity.IsPositive() {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "the Quantity is <= 0",
		}
	}
	return nil
}

// ListProductPlansRequest definition.
type ListProductPlansRequest struct {
	DepartmentOID string
	Date          time.Time
	ProductType   string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListProductPlansRequest) CheckInsufficiency() error {
	if req.Date.IsZero() || req.DepartmentOID == "" || req.ProductType == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// ProductPlanQuantity definition.
type ProductPlanQuantity struct {
	Daily    decimal.Decimal
	Week     decimal.Decimal
	Stock    decimal.Decimal
	Reserved decimal.Decimal
}

// ProductPlan definition.
type ProductPlan struct {
	ProductID string
	Quantity  ProductPlanQuantity
}

// ListProductPlansReply definition.
type ListProductPlansReply struct {
	ProductPlans []ProductPlan
}
