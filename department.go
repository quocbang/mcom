package mcom

import mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"

// CreateDepartmentsRequest : department ids.
type CreateDepartmentsRequest []string

// DeleteDepartmentRequest definition.
type DeleteDepartmentRequest struct {
	DepartmentID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req DeleteDepartmentRequest) CheckInsufficiency() error {
	if req.DepartmentID == "" {
		return mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}
	}
	return nil
}
