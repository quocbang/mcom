package mcom

import mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"

// ListProductIDsRequest definition.
type ListProductIDsRequest struct {
	Type          string
	IsLastProcess bool
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListProductIDsRequest) CheckInsufficiency() error {
	if req.Type == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// ListProductIDsReply definition.
type ListProductIDsReply []string

// ListProductGroupsRequest definition.
type ListProductGroupsRequest struct {
	DepartmentOID string
	Type          string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListProductGroupsRequest) CheckInsufficiency() error {
	if req.DepartmentOID == "" || req.Type == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// ListProductGroupsReply definition.
type ListProductGroupsReply struct {
	Products []ProductGroup
}

// ProductGroup definition.
type ProductGroup struct {
	ID       string
	Children ListProductIDsReply
}

// Product definition.
type Product struct {
	ID   string `validate:"required"`
	Type string `validate:"required"`
}

// ListProductTypesRequest definition.
type ListProductTypesRequest struct {
	DepartmentOID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListProductTypesRequest) CheckInsufficiency() error {
	// 允許不輸入任何部門OID
	return nil
}

// ListProductTypesReply : product types.
type ListProductTypesReply []string
