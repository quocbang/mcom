package mcom

import (
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/bindtype"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

type GetToolResourceRequest struct {
	ResourceID string `validate:"required"`
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetToolResourceRequest) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

// GetToolResourceReply definition.
type GetToolResourceReply struct {
	ToolID      string
	BindingSite models.UniqueSite
	CreatedBy   string
	CreatedAt   types.TimeNano
}

// Deprecated: use V2 instead
type ToolResourceBindRequest struct {
	Station string
	Details []ToolBindRequestDetail `validate:"gt=0,required,dive"`
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ToolResourceBindRequest) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

type ToolResourceBindRequestV2 struct {
	Details []ToolBindRequestDetailV2 `validate:"gt=0,required,dive"`
}

func (req ToolResourceBindRequestV2) GetDetails() []BindRequestDetail {
	res := make([]BindRequestDetail, len(req.Details))
	for i, d := range req.Details {
		res[i] = d
	}
	return res
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ToolResourceBindRequestV2) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

// Deprecated: use V2 instead
type ToolBindRequestDetail struct {
	Type     bindtype.BindType `validate:"required,eq=1101|eq=1110"`
	Site     models.SiteID
	Resource ToolResource `validate:"omitempty,nostructlevel"`
}

type ToolBindRequestDetailV2 struct {
	Type     bindtype.BindType `validate:"required,eq=1101|eq=1110"`
	Site     models.UniqueSite `validate:"required"`
	Resource ToolResource      `validate:"omitempty,nostructlevel"`
}

func (mbrd ToolBindRequestDetailV2) GetUniqueSite() models.UniqueSite {
	return mbrd.Site
}

type ToolResource struct {
	ResourceID  string `validate:"required"`
	ToolID      string
	BindingSite models.UniqueSite `validate:"omitempty,nostructlevel"`
}

func (tr ToolResource) ToModelResource() models.ToolResource {
	return models.ToolResource{
		ID:          tr.ResourceID,
		ToolID:      tr.ToolID,
		BindingSite: tr.BindingSite,
	}
}

func (tr ToolResource) Require() error {
	if err := validate.Struct(tr); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

func (tr ToolResource) Empty() error {
	zero := ToolResource{}
	if tr == zero {
		return nil
	}

	return mcomErr.Error{
		Code:    mcomErr.Code_BAD_REQUEST,
		Details: "there should be no resources while clearing a slot.",
	}
}

type ListToolResourcesRequest struct {
	ResourcesID []string `validate:"required,gt=0,dive,required"`
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListToolResourcesRequest) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

// ListToolResourcesReply definition.
type ListToolResourcesReply struct {
	Resources map[string]GetToolResourceReply
}
