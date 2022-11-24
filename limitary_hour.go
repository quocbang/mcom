package mcom

import (
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
)

type LimitaryHour struct {
	ProductType  string
	LimitaryHour LimitaryHourParameter
}

type LimitaryHourParameter struct {
	Min int32
	Max int32
}

type CreateLimitaryHourRequest struct {
	LimitaryHour []LimitaryHour
}

func (req CreateLimitaryHourRequest) CheckInsufficiency() error {
	if len(req.LimitaryHour) == 0 {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	// ProductType cant be empty, Min or Max never negative number,Max cannot be less than Min.
	for _, v := range req.LimitaryHour {
		if v.ProductType == "" ||
			v.LimitaryHour.Min < 0 ||
			v.LimitaryHour.Max < 0 ||
			v.LimitaryHour.Max < v.LimitaryHour.Min {
			return mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			}
		}
	}
	return nil
}

type GetLimitaryHourRequest struct {
	ProductType string
}

type GetLimitaryHourReply struct {
	LimitaryHour LimitaryHourParameter
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetLimitaryHourRequest) CheckInsufficiency() error {
	if req.ProductType == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "product type is required"}
	}
	return nil
}
