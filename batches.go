package mcom

import (
	"github.com/shopspring/decimal"

	"gitlab.kenda.com.tw/kenda/commons/v2/proto/golang/mes/v2/workorder"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// CreateBatchRequest definition.
type CreateBatchRequest struct {
	WorkOrder string
	Number    int16
	Status    workorder.BatchStatus
	Note      string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreateBatchRequest) CheckInsufficiency() error {
	switch {
	case req.WorkOrder == "":
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "empty work order is not allowed"}
	case req.Number <= 0:
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "number should be greater than zero"}
	}
	return nil
}

// UpdateBatchRequest definition.
type UpdateBatchRequest struct {
	WorkOrder string
	Number    int16
	Status    workorder.BatchStatus
	Note      string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req UpdateBatchRequest) CheckInsufficiency() error {
	if req.WorkOrder == "" || req.Number == 0 {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "empty work order or empty number"}
	}
	return nil
}

type GetBatchRequest struct {
	WorkOrder string
	Number    int16
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetBatchRequest) CheckInsufficiency() error {
	switch {
	case req.WorkOrder == "":
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "empty work order is not allowed"}
	case req.Number < 1:
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "number should be positive"}
	}
	return nil
}

type GetBatchReply struct {
	Info BatchInfo
}

// ListBatchesRequest definition.
type ListBatchesRequest struct {
	WorkOrder string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListBatchesRequest) CheckInsufficiency() error {
	if req.WorkOrder == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "empty work order is not allowed"}
	}
	return nil
}

// BatchInfo definition.
type BatchInfo struct {
	WorkOrder string
	Number    int16
	// pbWorkOrder.BatchStatus.
	Status    int32
	Records   []models.FeedRecord
	Note      string
	UpdatedAt types.TimeNano
	UpdatedBy string
}

// ListBatchesReply definition.
type ListBatchesReply []BatchInfo

// FeedRequest definition.
type FeedRequest struct {
	Batch       BatchID
	FeedContent []FeedPerSite
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req FeedRequest) CheckInsufficiency() error {
	if len(req.FeedContent) == 0 {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}

	for _, fc := range req.FeedContent {
		if err := fc.FeedPerSiteCheckInsufficiency(); err != nil {
			return err
		}
	}
	return nil
}

// FeedPerSite definition.
type FeedPerSite interface {
	FeedPerSiteCheckInsufficiency() error
}

type FeedPerSiteType1 struct {
	Site     models.UniqueSite
	FeedAll  bool
	Quantity decimal.Decimal
}

type FeedPerSiteType2 struct {
	Site        models.UniqueSite
	Quantity    decimal.Decimal `validate:"required"`
	ResourceID  string          `validate:"required"`
	ProductType string
}

type FeedPerSiteType3 struct {
	Quantity    decimal.Decimal `validate:"required"`
	ResourceID  string          `validate:"required"`
	ProductType string
}

func (fps FeedPerSiteType1) FeedPerSiteCheckInsufficiency() error {
	if err := validate.Struct(fps); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}

	if fps.Quantity.LessThan(decimal.Zero) {
		return mcomErr.Error{Code: mcomErr.Code_INVALID_NUMBER}
	}
	return nil
}

func (fps FeedPerSiteType2) FeedPerSiteCheckInsufficiency() error {
	if err := validate.Struct(fps); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}

	if fps.Site.SiteID.Name == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}

	if fps.Quantity.LessThan(decimal.Zero) {
		return mcomErr.Error{Code: mcomErr.Code_INVALID_NUMBER}
	}
	return nil
}

func (fps FeedPerSiteType3) FeedPerSiteCheckInsufficiency() error {
	if err := validate.Struct(fps); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}

	if fps.Quantity.LessThan(decimal.Zero) {
		return mcomErr.Error{Code: mcomErr.Code_INVALID_NUMBER}
	}
	return nil
}

type FeedReply struct {
	FeedRecordID string
}

// BatchID definition.
type BatchID struct {
	WorkOrder string
	Number    int16
}
