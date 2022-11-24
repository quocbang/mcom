package mcom

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/bindtype"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// SiteMaterial definition.
type SiteMaterial struct {
	ID         string
	Grade      string
	Quantity   *decimal.Decimal
	ResourceID string
	ExpiryTime time.Time
}

// ListSiteMaterialsReply definition.
type ListSiteMaterialsReply []SiteMaterial

// MaterialResourceBindRequest definition.
// Deprecated: use V2 instead.
type MaterialResourceBindRequest struct {
	Station string
	Details []MaterialBindRequestDetail
}

// MaterialResourceBindRequestV2 definition
type MaterialResourceBindRequestV2 struct {
	Details []MaterialBindRequestDetailV2 `validate:"required,gt=0,dive"`
}

func (req MaterialResourceBindRequestV2) GetDetails() []BindRequestDetail {
	res := make([]BindRequestDetail, len(req.Details))
	for i, d := range req.Details {
		res[i] = d
	}
	return res
}

// Deprecated: use V2 instead
type MaterialBindRequestDetail struct {
	Type      bindtype.BindType
	Site      models.SiteID
	Resources []BindMaterialResource

	// Queue and ColQueue Only
	Option models.BindOption
}

type MaterialBindRequestDetailV2 struct {
	Type      bindtype.BindType `validate:"gt=0"`
	Site      models.UniqueSite `validate:"required"`
	Resources []BindMaterialResource

	// Queue and ColQueue Only
	Option models.BindOption
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
// also check invalid number and bad request here.
func (req MaterialResourceBindRequestV2) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}

	for _, d := range req.Details {
		if d.Type == bindtype.BindType_RESOURCE_BINDING_NONE {
			return mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing bind type",
			}
		}

		if err := d.checkUnspecifiedQuantityBadRequest(); err != nil {
			return err
		}

		if err := d.checkBadRequest(); err != nil {
			return err
		}

	}
	return nil
}

func (d MaterialBindRequestDetailV2) checkUnspecifiedQuantityBadRequest() error {
	allowToBeUnspecifiedTypes := d.Type == bindtype.BindType_RESOURCE_BINDING_SLOT_BIND ||
		d.Type == bindtype.BindType_RESOURCE_BINDING_QUEUE_BIND ||
		d.Type == bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSH ||
		d.Type == bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSHPOP
	for _, resource := range d.Resources {
		if resource.Quantity == nil && !allowToBeUnspecifiedTypes {
			return mcomErr.Error{
				Code:    mcomErr.Code_BAD_REQUEST,
				Details: "unspecified quantity only applicable to slot bind or queue bind",
			}
		}
	}

	return nil
}

func (mbrd MaterialBindRequestDetailV2) GetUniqueSite() models.UniqueSite {
	return mbrd.Site
}

type BindMaterialResource struct {
	Material models.Material
	// IF Quantity == nil
	// For Slot bind, queue bind, queue push and queue pushpop.
	// The quantity of the resource will not be mutated while binding.
	Quantity    *decimal.Decimal
	ResourceID  string
	ProductType string
	Status      resources.MaterialStatus
	ExpiryTime  types.TimeNano
	Warehouse   Warehouse
}

func (resource BindMaterialResource) ToBoundResource() models.BoundResource {
	return models.BoundResource{
		Material: &models.MaterialSite{
			Material:    resource.Material,
			Quantity:    resource.Quantity,
			ResourceID:  resource.ResourceID,
			ProductType: resource.ProductType,
			Status:      resource.Status,
			ExpiryTime:  resource.ExpiryTime,
		},
	}
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req MaterialResourceBindRequest) CheckInsufficiency() error {
	// only one shared site can be targeted.
	if req.Station == "" && len(req.Details) != 1 {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "missing station or only one shared site can be targeted",
		}
	}

	for _, d := range req.Details {
		if d.Type == bindtype.BindType_RESOURCE_BINDING_NONE {
			return mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing bind type",
			}
		}
		if d.Site.Name == "" {
			return mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing SiteID",
			}
		}

		if err := d.checkUnspecifiedQuantityBadRequest(); err != nil {
			return err
		}

		if err := d.checkBadRequest(); err != nil {
			return err
		}

		for _, resource := range d.Resources {
			if resource.Quantity != nil {
				if resource.Quantity.LessThanOrEqual(*types.Decimal.NewFromInt32(0)) {
					return mcomErr.Error{
						Code:    mcomErr.Code_BAD_REQUEST,
						Details: "quantity cannot be less than 0 or not equal to nil",
					}
				}
			}
		}
	}
	return nil
}

func (d MaterialBindRequestDetail) checkUnspecifiedQuantityBadRequest() error {
	allowToBeUnspecifiedTypes := d.Type == bindtype.BindType_RESOURCE_BINDING_SLOT_BIND ||
		d.Type == bindtype.BindType_RESOURCE_BINDING_QUEUE_BIND ||
		d.Type == bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSH ||
		d.Type == bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSHPOP
	for _, resource := range d.Resources {
		if resource.Quantity == nil && !allowToBeUnspecifiedTypes {
			return mcomErr.Error{
				Code:    mcomErr.Code_BAD_REQUEST,
				Details: "unspecified quantity only applicable to slot bind or queue bind",
			}
		}
	}
	return nil
}

func (detail MaterialBindRequestDetailV2) checkBadRequest() error {
	cbrPair := defaultBindBadRequestPair
	t, ok := cbrPair[detail.Type]
	if !ok {
		return fmt.Errorf("undefine bind type")
	}
	return t.check(detail)
}

type checkResourceDelegate func([]BindMaterialResource) error
type checkIndexDelegate func(detail MaterialBindRequestDetailV2) error

var indexCheck = struct {
	Required    checkIndexDelegate
	NotRequired checkIndexDelegate
}{
	Required: func(detail MaterialBindRequestDetailV2) error {
		if !detail.checkIndex() {
			return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "invalid option"}
		}
		return nil
	},
	NotRequired: func(detail MaterialBindRequestDetailV2) error { return nil },
}

type checkBindBadRequest struct {
	ResourcePart checkResourceDelegate
	IndexPart    checkIndexDelegate
}

func (cbbr checkBindBadRequest) check(detail MaterialBindRequestDetailV2) error {
	if err := cbbr.ResourcePart(detail.Resources); err != nil {
		return err
	}
	return cbbr.IndexPart(detail)
}

var defaultBindBadRequestPair = map[bindtype.BindType]checkBindBadRequest{
	bindtype.BindType_RESOURCE_BINDING_CONTAINER_BIND: {
		ResourcePart: resourcesCheck.NotEmpty,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_CONTAINER_ADD: {
		ResourcePart: resourcesCheck.NotEmpty,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_CONTAINER_CLEAR: {
		ResourcePart: resourcesCheck.Empty,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_CONTAINER_CLEAN_DEVIATION: {
		ResourcePart: resourcesCheck.Empty,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_SLOT_BIND: {
		ResourcePart: resourcesCheck.One,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR: {
		ResourcePart: resourcesCheck.Empty,
		IndexPart:    indexCheck.NotRequired,
	},

	// collection operations
	bindtype.BindType_RESOURCE_BINDING_COLLECTION_BIND: {
		ResourcePart: resourcesCheck.NotEmpty,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_COLLECTION_ADD: {
		ResourcePart: resourcesCheck.NotEmpty,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_COLLECTION_CLEAR: {
		ResourcePart: resourcesCheck.Empty,
		IndexPart:    indexCheck.NotRequired,
	},
	// collection queue operations
	bindtype.BindType_RESOURCE_BINDING_COLQUEUE_BIND: {
		ResourcePart: resourcesCheck.NotEmpty,
		IndexPart:    indexCheck.Required,
	},
	bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD: {
		ResourcePart: resourcesCheck.NotEmpty,
		IndexPart:    indexCheck.Required,
	},
	bindtype.BindType_RESOURCE_BINDING_COLQUEUE_CLEAR: {
		ResourcePart: resourcesCheck.Empty,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_COLQUEUE_PUSH: {
		ResourcePart: resourcesCheck.NotEmpty,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_COLQUEUE_PUSHPOP: {
		ResourcePart: resourcesCheck.NotEmpty,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_COLQUEUE_POP: {
		ResourcePart: resourcesCheck.Empty,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_COLQUEUE_REMOVE: {
		ResourcePart: resourcesCheck.Empty,
		IndexPart:    indexCheck.Required,
	},
	// queue operations
	bindtype.BindType_RESOURCE_BINDING_QUEUE_BIND: {
		ResourcePart: resourcesCheck.One,
		IndexPart:    indexCheck.Required,
	},
	bindtype.BindType_RESOURCE_BINDING_QUEUE_POP: {
		ResourcePart: resourcesCheck.Empty,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSH: {
		ResourcePart: resourcesCheck.One,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSHPOP: {
		ResourcePart: resourcesCheck.One,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_QUEUE_CLEAR: {
		ResourcePart: resourcesCheck.Empty,
		IndexPart:    indexCheck.NotRequired,
	},
	bindtype.BindType_RESOURCE_BINDING_QUEUE_REMOVE: {
		ResourcePart: resourcesCheck.Empty,
		IndexPart:    indexCheck.Required,
	},
}

func (detail MaterialBindRequestDetail) checkBadRequest() error {
	switch detail.Type {
	case bindtype.BindType_RESOURCE_BINDING_CONTAINER_BIND:
		if err := resourcesCheck.NotEmpty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_CONTAINER_ADD:
		if err := resourcesCheck.NotEmpty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_CONTAINER_CLEAR:
		if err := resourcesCheck.Empty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_CONTAINER_CLEAN_DEVIATION:
		if err := resourcesCheck.Empty(detail.Resources); err != nil {
			return err
		}
		// slot operations.
	case bindtype.BindType_RESOURCE_BINDING_SLOT_BIND:
		if err := resourcesCheck.One(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR:
		if err := resourcesCheck.Empty(detail.Resources); err != nil {
			return err
		}

		// collection operations.
	case bindtype.BindType_RESOURCE_BINDING_COLLECTION_BIND:
		if err := resourcesCheck.NotEmpty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_COLLECTION_ADD:
		if err := resourcesCheck.NotEmpty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_COLLECTION_CLEAR:
		if err := resourcesCheck.Empty(detail.Resources); err != nil {
			return err
		}
		// collection queue operations.
	case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_BIND:
		if !detail.checkIndex() {
			return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "invalid option"}
		}
		if err := resourcesCheck.NotEmpty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD:
		if !detail.checkIndex() {
			return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "invalid option"}
		}
		if err := resourcesCheck.NotEmpty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_CLEAR:
		if err := resourcesCheck.Empty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_PUSH:
		if err := resourcesCheck.NotEmpty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_PUSHPOP:
		if err := resourcesCheck.NotEmpty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_POP:
		if err := resourcesCheck.Empty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_REMOVE:
		if err := resourcesCheck.Empty(detail.Resources); err != nil {
			return err
		}
		if !detail.checkIndex() {
			return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "invalid option"}
		}
		// queue operations.
	case bindtype.BindType_RESOURCE_BINDING_QUEUE_BIND:
		if !detail.checkIndex() {
			return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "invalid option"}
		}
		if err := resourcesCheck.One(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_QUEUE_POP:
		if err := resourcesCheck.Empty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSH:
		if err := resourcesCheck.One(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSHPOP:
		if err := resourcesCheck.One(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_QUEUE_CLEAR:
		if err := resourcesCheck.Empty(detail.Resources); err != nil {
			return err
		}
	case bindtype.BindType_RESOURCE_BINDING_QUEUE_REMOVE:
		if err := resourcesCheck.Empty(detail.Resources); err != nil {
			return err
		}
		if !detail.checkIndex() {
			return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "invalid option"}
		}
	}
	return nil
}

func (detail MaterialBindRequestDetailV2) checkIndex() bool {
	isHead := detail.Option.Head
	isTail := detail.Option.Tail
	isIndex := detail.Option.QueueIndex != nil
	return types.Btoi(isHead)+types.Btoi(isTail)+types.Btoi(isIndex) == 1
}

func (detail MaterialBindRequestDetail) checkIndex() bool {
	isHead := detail.Option.Head
	isTail := detail.Option.Tail
	isIndex := detail.Option.QueueIndex != nil
	return types.Btoi(isHead)+types.Btoi(isTail)+types.Btoi(isIndex) == 1
}

var resourcesCheck = struct {
	Empty    checkResourceDelegate
	One      checkResourceDelegate
	NotEmpty checkResourceDelegate
}{
	Empty: func(br []BindMaterialResource) error {
		if len(br) == 0 {
			return nil
		}
		return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "should not contain any resources"}
	},
	One: func(br []BindMaterialResource) error {
		if len(br) == 1 {
			return nil
		}
		return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "should contain only one resource"}
	},
	NotEmpty: func(br []BindMaterialResource) error {
		if len(br) != 0 {
			return nil
		}
		return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "should contain at least one resource"}
	},
}

// ListSiteMaterialsRequest definition.
type ListSiteMaterialsRequest struct {
	Station string `validate:"required"`
	Site    models.SiteID
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListSiteMaterialsRequest) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

type CreateSharedSiteRequest struct {
	SiteInfo     SiteInformation
	Stations     []string
	DepartmentID string
}

func (req CreateSharedSiteRequest) CheckInsufficiency() error {
	if req.SiteInfo.Name == "" ||
		req.SiteInfo.Type == sites.Type_TYPE_UNSPECIFIED ||
		req.SiteInfo.SubType == sites.SubType_SUB_TYPE_UNSPECIFIED ||
		req.DepartmentID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

type UpdateSharedSiteRequest struct {
	TargetSite         SiteID
	AddToStations      []string
	RemoveFromStations []string
}

func (req UpdateSharedSiteRequest) CheckInsufficiency() error {
	if req.TargetSite.Name == "" ||
		(len(req.AddToStations) == 0 && len(req.RemoveFromStations) == 0) {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

type DeleteSharedSiteRequest struct {
	TargetSite SiteID
}

func (req DeleteSharedSiteRequest) CheckInsufficiency() error {
	if req.TargetSite.Name == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

type BindRecordsCheckRequest struct {
	Site      models.UniqueSite
	Resources []models.UniqueMaterialResource `validate:"required,gt=0,dive"`
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req BindRecordsCheckRequest) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}
