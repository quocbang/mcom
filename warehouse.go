package mcom

import (
	"github.com/shopspring/decimal"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// WarehousingStockRequest definition.
type WarehousingStockRequest struct {
	Warehouse   Warehouse
	ResourceIDs []string // must be material.
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req WarehousingStockRequest) CheckInsufficiency() error {
	if len(req.ResourceIDs) == 0 ||
		req.Warehouse.ID == "" ||
		req.Warehouse.Location == "" {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "missing required request",
		}
	}

	for _, resourceID := range req.ResourceIDs {
		if resourceID == "" {
			return mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "some resourceID is empty",
			}
		}
	}
	return nil
}

// GetResourceWarehouseReply definition.
type GetResourceWarehouseReply Warehouse

// GetResourceWarehouseRequest definition.
type GetResourceWarehouseRequest struct {
	ResourceID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetResourceWarehouseRequest) CheckInsufficiency() error {
	if req.ResourceID == "" {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "empty resourceID",
		}
	}
	return nil
}

// MaterialReply definition.
type MaterialReply struct {
	Material  Material
	Warehouse Warehouse
}

// GetMaterialResourceReply definition.
type GetMaterialResourceReply []MaterialReply

type ListMaterialResourceIdentitiesReply struct {
	Replies []*MaterialReply
}

// GetMaterialResourceRequest definition.
type GetMaterialResourceRequest struct {
	ResourceID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetMaterialResourceRequest) CheckInsufficiency() error {
	return nil
}

// GetMaterialResourceIdentityRequest definition.
type GetMaterialResourceIdentityRequest struct {
	ResourceID, ProductType string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetMaterialResourceIdentityRequest) CheckInsufficiency() error {
	return nil
}

// GetMaterialResourceIdentityReply definition.
type GetMaterialResourceIdentityReply MaterialReply

type ListMaterialResourceIdentitiesRequest struct {
	Details []GetMaterialResourceIdentityRequest
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListMaterialResourceIdentitiesRequest) CheckInsufficiency() error {
	return nil
}

type ListMaterialResourcesRequest struct {
	ProductID   string
	ProductType string
	Status      resources.MaterialStatus
	CreatedAt   types.TimeNano

	paginationRequest PaginationRequest
	orderRequest      OrderRequest
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListMaterialResourcesRequest) CheckInsufficiency() error {
	return nil
}

func (req ListMaterialResourcesRequest) WithPagination(p PaginationRequest) ListMaterialResourcesRequest {
	req.paginationRequest = p
	return req
}

// NeedPagination implements gitlab.kenda.com.tw/kenda/mcom Sliceable interface.
func (req ListMaterialResourcesRequest) NeedPagination() bool {
	return req.paginationRequest != PaginationRequest{}
}

// GetOffset implements gitlab.kenda.com.tw/kenda/mcom Sliceable interface.
func (req ListMaterialResourcesRequest) GetOffset() int {
	return (int(req.paginationRequest.PageCount) - 1) * int(req.paginationRequest.ObjectsPerPage)
}

// GetLimit implements gitlab.kenda.com.tw/kenda/mcom Sliceable interface.
func (req ListMaterialResourcesRequest) GetLimit() int {
	return int(req.paginationRequest.ObjectsPerPage)
}

// ValidatePagination implements gitlab.kenda.com.tw/kenda/mcom Sliceable interface.
func (req ListMaterialResourcesRequest) ValidatePagination() error {
	return req.paginationRequest.Validate()
}

func (req ListMaterialResourcesRequest) WithOrder(os ...Order) ListMaterialResourcesRequest {
	req.orderRequest.OrderBy = append(req.orderRequest.OrderBy, os...)
	return req
}

// NeedOrder implements gitlab.kenda.com.tw/kenda/mcom Orderable interface.
func (req ListMaterialResourcesRequest) NeedOrder() bool {
	return len(req.orderRequest.OrderBy) != 0
}

// ValidateOrder implements gitlab.kenda.com.tw/kenda/mcom Orderable interface.
func (req ListMaterialResourcesRequest) ValidateOrder() error {
	for _, order := range req.orderRequest.OrderBy {
		if !isValidName[models.MaterialResource](order.Name) {
			return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "order: invalid field name"}
		}
	}
	return nil
}

// GetOrderFunc implements gitlab.kenda.com.tw/kenda/mcom Orderable interface.
func (req ListMaterialResourcesRequest) GetOrder() []Order {
	return req.orderRequest.OrderBy
}

type ListMaterialResourcesReply struct {
	Resources []MaterialReply
	PaginationReply
}

type ListMaterialResourceStatusReply []string

type ListMaterialResourcesByIdRequest struct {
	ResourcesID []string `validate:"required,min=1"`
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListMaterialResourcesByIdRequest) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

type SplitMaterialResourceRequest struct {
	ResourceID    string
	ProductType   string
	Quantity      decimal.Decimal
	InspectionIDs []int
	Remark        string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req SplitMaterialResourceRequest) CheckInsufficiency() error {
	if req.ResourceID == "" || req.ProductType == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

type ListMaterialResourcesByIdReply struct {
	// where the keys are product types.
	TypeResourcePairs []map[string]MaterialReply
}

type SplitMaterialResourceReply struct {
	NewResourceID string
}
