package mcom

import (
	"time"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
)

// Station definition.
type Station struct {
	ID string
	// AdminDepartmentOID is the department to be charge of the station.
	AdminDepartmentOID string
	Sites              []ListStationSite
	State              stations.State
	Information        StationInformation

	UpdatedBy  string
	UpdatedAt  time.Time
	InsertedBy string
	InsertedAt time.Time
}

// ListStationsReply definition.
type ListStationsReply struct {
	Stations []Station
	PaginationReply
}

// ListStationsRequest definition.
type ListStationsRequest struct {
	DepartmentOID string

	paginationRequest PaginationRequest
	orderRequest      OrderRequest
}

func (req ListStationsRequest) WithPagination(p PaginationRequest) ListStationsRequest {
	req.paginationRequest = p
	return req
}

// NeedPagination implements gitlab.kenda.com.tw/kenda/mcom Paginationable interface.
func (req ListStationsRequest) NeedPagination() bool {
	return req.paginationRequest != PaginationRequest{}
}

// GetOffset implements gitlab.kenda.com.tw/kenda/mcom Paginationable interface.
func (req ListStationsRequest) GetOffset() int {
	return (int(req.paginationRequest.PageCount) - 1) * int(req.paginationRequest.ObjectsPerPage)
}

// GetLimit implements gitlab.kenda.com.tw/kenda/mcom Paginationable interface.
func (req ListStationsRequest) GetLimit() int {
	return int(req.paginationRequest.ObjectsPerPage)
}

// ValidatePagination implements gitlab.kenda.com.tw/kenda/mcom Paginationable interface.
func (req ListStationsRequest) ValidatePagination() error {
	return req.paginationRequest.Validate()
}

func (req ListStationsRequest) WithOrder(o ...Order) ListStationsRequest {
	req.orderRequest = OrderRequest{OrderBy: o}
	return req
}

// NeedOrder implements gitlab.kenda.com.tw/kenda/mcom Orderable interface.
func (req ListStationsRequest) NeedOrder() bool {
	return len(req.orderRequest.OrderBy) != 0
}

// ValidateOrder implements gitlab.kenda.com.tw/kenda/mcom Orderable interface.
func (req ListStationsRequest) ValidateOrder() error {
	for _, order := range req.orderRequest.OrderBy {
		if !isValidName[models.Station](order.Name) {
			return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "order: invalid field name"}
		}
	}
	return nil
}

// GetOrderFunc implements gitlab.kenda.com.tw/kenda/mcom Orderable interface.
func (req ListStationsRequest) GetOrder() []Order {
	return req.orderRequest.OrderBy
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListStationsRequest) CheckInsufficiency() error {
	if req.DepartmentOID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// GetStationRequest definition.
type GetStationRequest struct {
	ID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetStationRequest) CheckInsufficiency() error {
	if req.ID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// GetStationReply definition.
type GetStationReply Station

// StationInformation definition.
type StationInformation struct {
	Code        string
	Description string
}

// StationGroupRequest definition.
type StationGroupRequest struct {
	ID       string
	Stations []string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req StationGroupRequest) CheckInsufficiency() error {
	if req.ID == "" || len(req.Stations) == 0 {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// DeleteStationGroupRequest definition.
type DeleteStationGroupRequest struct {
	GroupID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req DeleteStationGroupRequest) CheckInsufficiency() error {
	if req.GroupID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

type ListAssociatedStationsRequest struct {
	Site models.UniqueSite `validate:"required"`
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListAssociatedStationsRequest) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

// ListAssociatedStationsReply definition.
type ListAssociatedStationsReply struct {
	StationIDs []string
}

// GetSiteRequest definition.
type GetSiteRequest struct {
	StationID string `validate:"required"`
	SiteName  string
	SiteIndex int16
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetSiteRequest) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

// GetSiteReply definition.
type GetSiteReply struct {
	Name               string
	Index              int16
	AdminDepartmentOID string
	Attributes         SiteAttributes
	Content            models.SiteContent
}

type SiteAttributes struct {
	Type    sites.Type
	SubType sites.SubType

	// LimitHandler returns whether the product can be bound into this site or not.
	// The returned USER_ERROR would be as below:
	//  - mcomErr.Code_PRODUCT_ID_MISMATCH
	LimitHandler func(productID string) error
}

func NewSiteAttributes(sa models.SiteAttributes) SiteAttributes {
	limitationMapping := make(map[string]bool, len(sa.Limitation))
	for _, limit := range sa.Limitation {
		limitationMapping[limit] = true
	}

	return SiteAttributes{
		Type:    sa.Type,
		SubType: sa.SubType,
		LimitHandler: func(productID string) error {
			if len(limitationMapping) == 0 {
				return nil
			}
			_, ok := limitationMapping[productID]
			if !ok {
				return mcomErr.Error{Code: mcomErr.Code_PRODUCT_ID_MISMATCH, Details: "product id: " + productID}
			}
			return nil
		},
	}
}

// SiteInformation definition.
type SiteInformation struct {
	// default: CreateStationRequest.ID / UpdateStationRequest.ID
	Station    string
	Name       string
	Index      int
	Type       sites.Type
	SubType    sites.SubType
	Limitation []string
}

// UpdateStationSite definition.
type UpdateStationSite struct {
	ActionMode  sites.ActionType
	Information SiteInformation
}

// ListStationSite definition.
type ListStationSite struct {
	Information ListStationSitesInformation
	Content     models.SiteContent
}

type ListStationSitesInformation struct {
	models.UniqueSite
	Type       sites.Type
	SubType    sites.SubType
	Limitation []string
}

// CreateStationRequest definition.
//
// CreateStationRequest.Sites[i].Station has a default value CreateStationRequest.ID.
type CreateStationRequest struct {
	ID            string
	DepartmentOID string
	Sites         []SiteInformation
	State         stations.State
	Information   StationInformation
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreateStationRequest) CheckInsufficiency() error {
	if req.ID == "" || req.DepartmentOID == "" {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "the ID or departmentOID is empty",
		}
	}
	return nil
}

func (req *CreateStationRequest) Correct() {
	for i := range req.Sites {
		if req.Sites[i].Station == "" {
			req.Sites[i].Station = req.ID
		}
	}
}

// UpdateStationRequest definition.
//
// UpdateStationRequest.Sites[i].Station has a default value UpdateStationRequest.ID.
type UpdateStationRequest struct {
	ID            string
	DepartmentOID string
	Sites         []UpdateStationSite
	State         stations.State
	Information   StationInformation
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req UpdateStationRequest) CheckInsufficiency() error {
	if req.ID == "" {
		return mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}
	}
	return nil
}

func (req *UpdateStationRequest) Correct() {
	for i := range req.Sites {
		if req.Sites[i].Information.Station == "" {
			req.Sites[i].Information.Station = req.ID
		}
	}
}

// DeleteStationRequest : id.
type DeleteStationRequest struct {
	StationID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req DeleteStationRequest) CheckInsufficiency() error {
	if req.StationID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// ListSiteTypeReply definition.
type ListSiteTypeReply []SiteType

// ListSiteSubTypeReply definition.
type ListSiteSubTypeReply []SiteSubType

// SiteType definition.
type SiteType struct {
	Name  string
	Value sites.Type
}

// SiteSubType definition.
type SiteSubType struct {
	Name  string
	Value sites.SubType
}

// StationState definition.
type StationState struct {
	Name  string
	Value stations.State
}

// ListStationStateReply definition.
type ListStationStateReply []StationState

type ListStationIDsRequest struct {
	DepartmentOID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListStationIDsRequest) CheckInsufficiency() error {
	return nil
}

// ListStationIDsReply definition.
type ListStationIDsReply struct {
	Stations []string
}
