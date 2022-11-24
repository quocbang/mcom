package mcom

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"

	pbWorkorder "gitlab.kenda.com.tw/kenda/commons/v2/proto/golang/mes/v2/workorder"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/workorder"
)

// GetWorkOrderReply definition.
type GetWorkOrderReply struct {
	ID       string
	Product  Product
	Process  WorkOrderProcess
	RecipeID string

	Status        pbWorkorder.Status
	DepartmentOID string
	Station       string
	Sequence      int32
	Date          time.Time
	Unit          string
	Note          string
	Parent        string

	models.BatchQuantityDetails

	CurrentBatch      int
	CollectedSequence int
	Abnormality       workorder.Abnormality

	CollectedQuantity decimal.Decimal

	UpdatedBy  string
	UpdatedAt  time.Time
	InsertedBy string
	InsertedAt time.Time
}

// GetWorkOrderRequest definition.
type GetWorkOrderRequest struct {
	ID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetWorkOrderRequest) CheckInsufficiency() error {
	if req.ID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// WorkOrderProcess definition.
type WorkOrderProcess struct {
	OID  string
	Name string
	Type string
}

// CreateWorkOrdersReply : work order ids.
type CreateWorkOrdersReply struct {
	IDs []string
}

// CreateWorkOrder definition.
type CreateWorkOrder struct {
	ProcessOID  string
	RecipeID    string
	ProcessName string
	ProcessType string

	Status          pbWorkorder.Status
	DepartmentOID   string
	Station         string
	Sequence        int32
	BatchesQuantity BatchQuantity
	Parent          string

	// Date is the reserved date when to produce.
	Date time.Time
	Unit string
	Note string
}

// CreateWorkOrdersRequest definition.
type CreateWorkOrdersRequest struct {
	WorkOrders []CreateWorkOrder
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreateWorkOrdersRequest) CheckInsufficiency() error {
	if len(req.WorkOrders) == 0 {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	for _, value := range req.WorkOrders {
		if value.ProcessOID == "" || value.RecipeID == "" || value.DepartmentOID == "" ||
			value.Date.IsZero() || value.ProcessName == "" || value.ProcessType == "" || value.BatchesQuantity == nil {
			return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
		}
	}
	return nil
}

// UpdateWorkOrder definition.
type UpdateWorkOrder struct {
	ID      string `validate:"required"`
	Status  pbWorkorder.Status
	Station string

	RecipeID    string
	ProcessOID  string
	ProcessName string
	ProcessType string

	// Date is the reserved date when to produce.
	Date     time.Time
	Sequence int32

	BatchesQuantity BatchQuantity

	Abnormality workorder.Abnormality
}

// UpdateWorkOrdersRequest definition.
type UpdateWorkOrdersRequest struct {
	Orders []UpdateWorkOrder `validate:"min=1,dive"`
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req UpdateWorkOrdersRequest) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

// ListWorkOrdersRequest definition.
type ListWorkOrdersRequest struct {
	ID string
	// Date is reserved date.
	Date    time.Time
	Station string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListWorkOrdersRequest) CheckInsufficiency() error {
	if req.Station == "" || req.Date.IsZero() {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// ListWorkOrdersReply definition.
type ListWorkOrdersReply struct {
	WorkOrders []GetWorkOrderReply
}

type ListWorkOrdersByDurationRequest struct {
	ID           string
	Since        time.Time `validate:"required"`
	Until        time.Time
	Status       []pbWorkorder.Status
	Limit        int
	Station      string
	DepartmentID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req *ListWorkOrdersByDurationRequest) CheckInsufficiency() error {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	if req.Until.IsZero() {
		req.Until = time.Date(9999, 12, 31, 0, 0, 0, 0, time.Local)
	}
	if req.Until.Before(req.Since) {
		return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "invalid time value"}
	}
	if req.DepartmentID == "" && req.Station == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// ListWorkOrdersByDurationReply definition.
type ListWorkOrdersByDurationReply struct {
	Contents []GetWorkOrderReply
}

type ListWorkOrdersByIDsRequest struct {
	IDs     []string `validate:"required"`
	Station string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListWorkOrdersByIDsRequest) CheckInsufficiency() error {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

// ListWorkOrdersByIDsReply definition.
type ListWorkOrdersByIDsReply struct {
	// where the key means work order id.
	Contents map[string]GetWorkOrderReply
}
