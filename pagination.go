package mcom

import (
	"github.com/go-playground/validator/v10"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
)

type Sliceable interface {
	// WithPagination(PaginationRequest)self.
	NeedPagination() bool
	GetOffset() int
	GetLimit() int
	ValidatePagination() error
}

type PaginationRequest struct {
	PageCount      uint `validate:"required,min=1"`
	ObjectsPerPage uint `validate:"required,min=1"`
}

func (req PaginationRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: err.Error(),
		}
	}
	return nil
}

type PaginationReply struct {
	AmountOfData int64
}
