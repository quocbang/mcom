package mcom

import (
	"github.com/shopspring/decimal"

	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// CreatePackRecordsRequest definition.
type CreatePackRecordsRequest struct {
	Station string
	Packs   []PackReq
}

type PackReq struct {
	Packing      string
	PackNumber   int
	Quantity     int
	ActualWeight decimal.Decimal
	Timestamp    types.TimeNano
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreatePackRecordsRequest) CheckInsufficiency() error {
	return nil
}

type ListPackRecordsReply struct {
	Packs []Pack
}

type Pack struct {
	SerialNumber int64
	Packing      string
	PackNumber   int
	Quantity     int
	ActualWeight decimal.Decimal
	Station      string
	CreatedAt    int64
	CreatedBy    string
}
