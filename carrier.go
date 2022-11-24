package mcom

import (
	"fmt"
	"time"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

// CarrierInfo Definition.
type CarrierInfo struct {
	ID              string
	AllowedMaterial string
	Contents        []string
	UpdateBy        string
	UpdateAt        time.Time
}

func NewCarrierInfo(carrier models.Carrier) CarrierInfo {
	return CarrierInfo{
		ID:              carrier.IDPrefix + fmt.Sprintf("%04d", carrier.SerialNumber),
		AllowedMaterial: carrier.AllowedMaterial,
		Contents:        carrier.Contents,
		UpdateBy:        carrier.UpdatedBy,
		UpdateAt:        carrier.UpdatedAt.Time(),
	}
}

// ListCarriersReply definition.
type ListCarriersReply struct {
	Info []CarrierInfo
	PaginationReply
}

type GetCarrierReply CarrierInfo

// ListCarriersRequest definition.
type ListCarriersRequest struct {
	DepartmentOID string

	paginationRequest PaginationRequest
	orderRequest      OrderRequest
}

func (req ListCarriersRequest) WithPagination(p PaginationRequest) ListCarriersRequest {
	req.paginationRequest = p
	return req
}

// NeedPagination implements gitlab.kenda.com.tw/kenda/mcom Paginationable interface.
func (req ListCarriersRequest) NeedPagination() bool {
	return req.paginationRequest != PaginationRequest{}
}

// GetOffset implements gitlab.kenda.com.tw/kenda/mcom Paginationable interface.
func (req ListCarriersRequest) GetOffset() int {
	return (int(req.paginationRequest.PageCount) - 1) * int(req.paginationRequest.ObjectsPerPage)
}

// GetLimit implements gitlab.kenda.com.tw/kenda/mcom Paginationable interface.
func (req ListCarriersRequest) GetLimit() int {
	return int(req.paginationRequest.ObjectsPerPage)
}

// ValidatePagination implements gitlab.kenda.com.tw/kenda/mcom Paginationable interface.
func (req ListCarriersRequest) ValidatePagination() error {
	return req.paginationRequest.Validate()
}

func (req ListCarriersRequest) WithOrder(o ...Order) ListCarriersRequest {
	req.orderRequest.OrderBy = o
	return req
}

// NeedOrder implements gitlab.kenda.com.tw/kenda/mcom Orderable interface.
func (req ListCarriersRequest) NeedOrder() bool {
	return len(req.orderRequest.OrderBy) != 0
}

// ValidateOrder implements gitlab.kenda.com.tw/kenda/mcom Orderable interface.
func (req ListCarriersRequest) ValidateOrder() error {

	for _, order := range req.orderRequest.OrderBy {
		if !isValidName[models.Carrier](order.Name) {
			return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "order: invalid field name"}
		}
	}
	return nil
}

// GetOrderFunc implements gitlab.kenda.com.tw/kenda/mcom Orderable interface.
func (req ListCarriersRequest) GetOrder() []Order {
	return req.orderRequest.OrderBy
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListCarriersRequest) CheckInsufficiency() error {
	if req.DepartmentOID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "department OID is required"}
	}
	return nil
}

type GetCarrierRequest struct {
	ID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetCarrierRequest) CheckInsufficiency() error {
	if req.ID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "carrier ID is required"}
	}
	return nil
}

// UpdateCarrierRequest Definition.
type UpdateCarrierRequest struct {
	ID     string
	Action UpdateCarrierAction
}

type UpdateCarrierAction interface {
	implementUpdateCarrierAction()
}

type UpdateProperties struct {
	AllowedMaterial string
	DepartmentOID   string
}

func (um UpdateProperties) implementUpdateCarrierAction() {}

type BindResources struct {
	ResourcesID []string
}

func (br BindResources) implementUpdateCarrierAction() {}

type RemoveResources struct {
	ResourcesID []string
}

func (rr RemoveResources) implementUpdateCarrierAction() {}

type ClearResources struct{}

func (cr ClearResources) implementUpdateCarrierAction() {}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req UpdateCarrierRequest) CheckInsufficiency() error {
	if req.ID == "" {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "id is required",
		}
	}
	if req.Action == nil {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "missing action",
		}
	}
	return nil
}

// CreateCarrierRequest Definition.
type CreateCarrierRequest struct {
	DepartmentOID string
	// leading chars of ID.
	IDPrefix        string
	Quantity        int32
	AllowedMaterial string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreateCarrierRequest) CheckInsufficiency() error {
	if req.DepartmentOID == "" || req.Quantity == 0 || len(req.IDPrefix) != 2 {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "illegal CreateCarrierRequest"}
	}
	return nil
}

// DeleteCarrierRequest definition.
type DeleteCarrierRequest struct {
	ID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req DeleteCarrierRequest) CheckInsufficiency() error {
	if req.ID == "" {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "id is required",
		}
	}
	return nil
}
