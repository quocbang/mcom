package mcom

import (
	"time"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/roles"
)

// GetTokenInfoReply definition.
type GetTokenInfoReply struct {
	User        string
	Valid       bool
	ExpiryTime  time.Time
	CreatedTime time.Time
	Roles       []roles.Role
}

// GetTokenInfoRequest definition.
type GetTokenInfoRequest struct {
	Token string
}

// Department definition.
type Department struct {
	OID string
	ID  string
}

// SignInReply definition.
type SignInReply struct {
	Token              string
	TokenExpiry        time.Time
	Departments        []Department
	Roles              []roles.Role
	MustChangePassword bool
}

// SignInOptions definition.
type SignInOptions struct {
	ExpiredAfter time.Duration
}

func newSignInOptions() SignInOptions {
	return SignInOptions{
		ExpiredAfter: 8 * time.Hour,
	}
}

// ParseSignInOptions returns options for sign-in.
func ParseSignInOptions(opts []SignInOption) SignInOptions {
	o := newSignInOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// SignInOption definition.
type SignInOption func(*SignInOptions)

// WithTokenExpiredAfter defines the token expired after d.
func WithTokenExpiredAfter(d time.Duration) SignInOption {
	return func(o *SignInOptions) {
		o.ExpiredAfter = d
	}
}

// User definition.
type User struct {
	ID           string
	Account      string
	DepartmentID string
}

// CreateUsersRequest definition.
type CreateUsersRequest struct {
	Users []User
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreateUsersRequest) CheckInsufficiency() error {
	if len(req.Users) == 0 {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "there is no user to create",
		}
	}
	for _, user := range req.Users {
		if user.ID == "" || user.DepartmentID == "" {
			return mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "the user id or the department id is empty",
			}
		}
	}
	return nil
}

// UpdateUserRequest defintion.
type UpdateUserRequest struct {
	ID           string
	Account      string
	DepartmentID string
	LeaveDate    time.Time
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req UpdateUserRequest) CheckInsufficiency() error {
	if req.ID == "" ||
		(req.DepartmentID == "" && req.LeaveDate.IsZero() && req.Account == "") {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "the user id is empty or one of the department id or leave date or account is empty",
		}
	}
	return nil
}

// UpdateDepartmentRequest definition.
type UpdateDepartmentRequest struct {
	OldID string
	NewID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req UpdateDepartmentRequest) CheckInsufficiency() error {
	if req.OldID == "" || req.NewID == "" {
		return mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}
	}
	return nil
}

// UserRoles definition.
type UserRoles struct {
	// account id.
	ID    string
	Roles []roles.Role
}

// ListUserRolesReply definition.
type ListUserRolesReply struct {
	Users []UserRoles
}

// ListUserRolesRequest definition.
type ListUserRolesRequest struct {
	DepartmentOID string
}

// Role definition.
type Role struct {
	Name  string
	Value roles.Role
}

// ListRolesReply definition.
type ListRolesReply struct {
	Roles []Role
}

// DeleteUserRequest definition.
type DeleteUserRequest struct {
	ID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req DeleteUserRequest) CheckInsufficiency() error {
	if req.ID == "" {
		return mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}
	}
	return nil
}

// AuthUserRoleRequest definition.
type AuthUserRoleRequest map[string][]roles.Role

// ListUnauthorizedUsersRequest definition.
type ListUnauthorizedUsersRequest struct {
	DepartmentOID string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListUnauthorizedUsersRequest) CheckInsufficiency() error {
	if req.DepartmentOID == "" {
		return mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}
	}
	return nil
}

// ListUnauthorizedUsersReply definition.
type ListUnauthorizedUsersReply []string

// ResetUserPasswordRequest definition.
type ResetUserPasswordRequest struct {
	ID string
}

// ListUnauthorizedUsersOption definition.
type ListUnauthorizedUsersOption func(*ListUnauthorizedUsersOptions)

// ListUnauthorizedUsersOptions definition.
type ListUnauthorizedUsersOptions struct {
	ExcludeUsers []string
}

// ParseListUnauthorizedUsersOptions definition.
func ParseListUnauthorizedUsersOptions(opts []ListUnauthorizedUsersOption) ListUnauthorizedUsersOptions {
	var opt ListUnauthorizedUsersOptions
	for _, optFunc := range opts {
		optFunc(&opt)
	}
	return opt
}

// ExcludeUsers excludes the specified user IDs while listing unauthorized users.
func ExcludeUsers(users []string) ListUnauthorizedUsersOption {
	return func(opt *ListUnauthorizedUsersOptions) {
		opt.ExcludeUsers = users
	}
}

// SignInStationRequest definition.
type SignInStationRequest struct {
	Station  string
	Site     models.SiteID
	Group    int32
	WorkDate time.Time
	// default 8 hours. Deprcated.
	ExpiryDuration time.Duration
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req SignInStationRequest) CheckInsufficiency() error {
	if req.Station == "" {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "missing station",
		}
	}
	return nil
}

type SignInStationOptions struct {
	VerifyWorkDate        func(time.Time) bool
	Force                 bool
	CreateSiteIfNotExists bool
}

func newSignInStationOptions() SignInStationOptions {
	return SignInStationOptions{
		VerifyWorkDate: alwaysPassVerifyWorkDateHandler,
	}
}

func ParseSignInStationOptions(opts []SignInStationOption) SignInStationOptions {
	o := newSignInStationOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

type SignInStationOption func(*SignInStationOptions)

func alwaysPassVerifyWorkDateHandler(time.Time) bool { return true }

func WithVerifyWorkDateHandler(f func(time.Time) bool) SignInStationOption {
	return func(o *SignInStationOptions) {
		o.VerifyWorkDate = f
	}
}

func ForceSignIn() SignInStationOption {
	return func(siso *SignInStationOptions) {
		siso.Force = true
	}
}

func CreateSiteIfNotExists() SignInStationOption {
	return func(siso *SignInStationOptions) {
		siso.CreateSiteIfNotExists = true
	}
}

// SignOutStationRequest definition.
type SignOutStationRequest struct {
	Station string
	Site    models.SiteID
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req SignOutStationRequest) CheckInsufficiency() error {
	if req.Station == "" {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "missing station",
		}
	}
	return nil
}

// ListAllDepartmentReply definition.
type ListAllDepartmentReply struct {
	IDs []string
}

type SignOutStationsRequest struct {
	Sites []models.UniqueSite `validate:"min=1"`
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req SignOutStationsRequest) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}
