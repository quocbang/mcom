package mcom

import (
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/utils/roles"
)

// CreateAccountRequest is a request for DataManager.CreateAccount()
//
// Notice that you must call at least one of the following methods.
//  - WithDefaultPassword => the default password is the same as the user id.
//  - WithSpecifiedPassword
type CreateAccountRequest struct {
	ID       string
	Roles    Roles
	password string
}

// CreateAccountsRequest definition.
//
// read the comments of CreateAccountRequest for details.
type CreateAccountsRequest []CreateAccountRequest

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreateAccountsRequest) CheckInsufficiency() error {
	if len(req) == 0 {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "CreateAccounts: empty request",
		}
	}
	for _, elem := range req {
		if err := elem.CheckInsufficiency(); err != nil {
			return err
		}
	}
	return nil
}

// WithDefaultPassword definition.
func (req CreateAccountRequest) WithDefaultPassword() CreateAccountRequest {
	req.password = req.ID
	return req
}

// WithSpecifiedPassword definition.
func (req CreateAccountRequest) WithSpecifiedPassword(pwd string) CreateAccountRequest {
	req.password = pwd
	return req
}

// GetPassword definition.
func (req CreateAccountRequest) GetPassword() string {
	return req.password
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreateAccountRequest) CheckInsufficiency() error {
	if req.ID == "" || len(req.Roles) == 0 || req.password == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// Roles definition.
type Roles []roles.Role

// CheckPermission checks the permission of roles while creating or updating accounts.
// Notice that the Administrator and Leader is not allow to mes accounts.
//
// The returned USER_ERROR would be as below:
//  - Code_ACCOUNT_ROLES_NOT_PERMIT
func (r Roles) CheckPermission() error {
	for _, role := range r {
		if role == roles.Role_ADMINISTRATOR || role == roles.Role_LEADER {
			return mcomErr.Error{
				Code:    mcomErr.Code_ACCOUNT_ROLES_NOT_PERMIT,
				Details: "Administrator and Leader are not allowed to add",
			}
		}
	}
	return nil
}

// UpdateAccountRequest definition.
type UpdateAccountRequest struct {
	UserID         string
	ChangePassword *struct {
		NewPassword string
		OldPassword string
	}
	Roles []roles.Role
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req UpdateAccountRequest) CheckInsufficiency() error {
	if req.UserID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

func (req UpdateAccountRequest) IsChangingPassword() bool {
	return req.ChangePassword != nil
}

func (req UpdateAccountRequest) IsModifyingRoles() bool {
	return len(req.Roles) > 0
}

// UpdateAccountOption definition.
type UpdateAccountOption func(*UpdateAccountOptions)

// UpdateAccountOptions definition.
type UpdateAccountOptions struct {
	ResetPassword bool
}

// ParseUpdateAccountOptions definition.
func ParseUpdateAccountOptions(opts []UpdateAccountOption) UpdateAccountOptions {
	var opt UpdateAccountOptions
	for _, optFunc := range opts {
		optFunc(&opt)
	}
	return opt
}

// ResetPassword definition.
func ResetPassword() UpdateAccountOption {
	return func(uao *UpdateAccountOptions) {
		uao.ResetPassword = true
	}
}

// SignInRequest definition.
type SignInRequest struct {
	Account  string
	Password string

	ADUser bool
}

// SignOutRequest definition.
type SignOutRequest struct {
	Token string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req SignOutRequest) CheckInsufficiency() error {
	if req.Token == "" {
		return mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}
	}
	return nil
}

// DeleteAccountRequest definition.
type DeleteAccountRequest struct {
	ID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req DeleteAccountRequest) CheckInsufficiency() error {
	if req.ID == "" {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "missing user id",
		}
	}
	return nil
}
