package mcom

import (
	"time"

	"github.com/shopspring/decimal"
)

// GetMaterialRequest definition.
type GetMaterialRequest struct {
	MaterialID string
}

// GetMaterialReply definition.
type GetMaterialReply struct {
	MaterialProductID string
	MaterialID        string
	MaterialType      string
	Sequence          string
	Status            string
	Quantity          decimal.Decimal
	Comment           string
	ExpireDate        time.Time
}

// UpdateMaterialRequest definition.
type UpdateMaterialRequest struct {
	MaterialID       string
	ExtendedDuration time.Duration
	User             string
	NewStatus        string
	Reason           string
	ProductCate      string
	ControlArea      string
}

// Code definition.
type Code struct {
	Code            string
	CodeDescription string
}

// ListControlReasonsReply definition.
type ListControlReasonsReply struct {
	Codes []*Code
}

// ListControlAreasReply definition.
type ListControlAreasReply struct {
	Codes []*Code
}

// ListChangeableStatusReply definition.
type ListChangeableStatusReply struct {
	Codes []*Code
}
