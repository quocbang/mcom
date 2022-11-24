package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"

	"gitlab.kenda.com.tw/kenda/mcom/utils/roles"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// Token table definition.
type Token struct {
	ID EncryptedData `gorm:"type:bytea;primaryKey"`

	// BoundUser means the token is bound to the user.
	BoundUser string `gorm:"type:text;not null"`

	// ExpiryTime is expiry time of the token. The number of nanoseconds
	// elapsed since January 1, 1970 UTC.
	ExpiryTime types.TimeNano `gorm:"not null"`

	// CreatedTime is the time that the token is created. The number of
	// nanoseconds elapsed since January 1, 1970 UTC.
	CreatedTime types.TimeNano `gorm:"not null;autoCreateTime:nano"`

	// Valid is whether the token is valid or not.
	Valid bool `gorm:"default:true;not null"`

	Info UserInfo `gorm:"type:json;default:'{}';not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (Token) TableName() string {
	return "token"
}

// UserInfo is the information of the user.
type UserInfo struct {
	Roles []roles.Role `json:"roles"`
}

// Scan implements database/sql Scanner interface.
func (v *UserInfo) Scan(src interface{}) error {
	return ScanJSON(src, v)
}

// Value implements database/sql/driver Valuer interface.
func (v UserInfo) Value() (driver.Value, error) {
	return json.Marshal(v)
}

// Account defines who can sign in.
type Account struct {
	// ID is relative to User.ID.
	ID       string        `gorm:"type:text;primaryKey"`
	Password EncryptedData `gorm:"type:bytea;not null"`
	// Roles are relative to gitlab.kenda.com.tw/kenda/mcom/utils/roles enumerations.
	Roles              pq.Int64Array `gorm:"type:smallint[];default:'{}';not null"`
	MustChangePassword bool          `gorm:"type:boolean;default:false;not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (Account) TableName() string {
	return "account"
}

// User is in-service staff.
type User struct {
	// ID is an employee id.
	ID string `gorm:"type:text;primaryKey"`
	// Active Directory Account.
	Account string `gorm:"type:text;index:idx_user_account"`
	// DepartmentID is relative to Department.ID.
	DepartmentID string       `gorm:"column:department_id;type:text;not null"`
	LeaveDate    sql.NullTime `gorm:"type:date"`
}

// Resigned returns true if the user has resigned.
func (user *User) Resigned() bool {
	return user.LeaveDate.Valid && user.LeaveDate.Time.Before(time.Now())
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (User) TableName() string {
	return "user"
}

// Department definition.
type Department struct {
	ID string `gorm:"column:id;type:char(5);not null;index:idx_department_id;primaryKey"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (Department) TableName() string {
	return "department"
}
