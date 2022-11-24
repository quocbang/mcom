package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/lib/pq"

	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// Station definition.
type Station struct {
	ID string `gorm:"type:varchar(32);primaryKey"`

	// AdminDepartmentID is the department ID to be charge of this station.
	// The value is relative to Department.ID.
	AdminDepartmentID string `gorm:"column:admin_department_id;type:text;not null;index:idx_station_department"`

	Sites StationSites `gorm:"type:jsonb;default:'[]';not null"`

	State stations.State `gorm:"default:1;not null"`

	// Information would NOT be filter conditions.
	Information StationInformation `gorm:"type:json;default:'{}';not null"`

	// UpdatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	UpdatedAt types.TimeNano `gorm:"autoUpdateTime:nano;not null"`
	UpdatedBy string         `gorm:"type:text;not null"`
	// CreatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	CreatedAt types.TimeNano `gorm:"autoCreateTime:nano;not null"`
	CreatedBy string         `gorm:"type:text;not null"`
}

func (station *Station) RemoveSharedSite(target SiteID) bool {
	sites := station.Sites
	res := make([]UniqueSite, 0, len(sites))
	for _, site := range sites {
		if site.Station == "" &&
			site.SiteID.Name == target.Name &&
			site.SiteID.Index == target.Index {
			continue
		}
		res = append(res, site)
	}
	station.Sites = res
	return len(res) < len(sites)
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (Station) TableName() string {
	return "station"
}

// StationInformation definition.
type StationInformation struct {
	// Code is represented as what station is. It is used for production batch.
	// number.
	Code string `json:"code"`

	Description string `json:"description,omitempty"`
}

// Scan implements database/sql Scanner interface.
func (info *StationInformation) Scan(src interface{}) error {
	return ScanJSON(src, info)
}

// Value implements database/sql/driver Valuer interface.
func (info StationInformation) Value() (driver.Value, error) {
	return json.Marshal(info)
}

// SitesID definition.
type SitesID []SiteID

type StationSites []UniqueSite

// Scan implements database/sql Scanner interface.
func (sites *StationSites) Scan(src interface{}) error {
	return ScanJSON(src, sites)
}

// Value implements database/sql/driver Valuer interface.
func (sites StationSites) Value() (driver.Value, error) {
	return json.Marshal(sites)
}

// UniqueSite specifies a site in the station.
// The "Station" field in shared sites is "".
type UniqueSite struct {
	SiteID  SiteID `json:"site"`
	Station string `json:"station" validate:"required"`
}

// Scan implements database/sql Scanner interface.
func (site *UniqueSite) Scan(src interface{}) error {
	return ScanJSON(src, site)
}

// Value implements database/sql/driver Valuer interface.
func (site UniqueSite) Value() (driver.Value, error) {
	return json.Marshal(site)
}

func (s StationSites) Contains(target SiteID) bool {
	for _, site := range s {
		if site.SiteID == target {
			return true
		}
	}
	return false
}

// Scan implements database/sql Scanner interface.
func (sites *SitesID) Scan(src interface{}) error {
	return ScanJSON(src, sites)
}

// Value implements database/sql/driver Valuer interface.
func (sites SitesID) Value() (driver.Value, error) {
	return json.Marshal(sites)
}

// StationGroup definition.
type StationGroup struct {
	ID string `gorm:"type:text;primaryKey"`
	// Stations are relative to Station.ID.
	Stations pq.StringArray `gorm:"type:varchar(32)[];default:'{}';not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (StationGroup) TableName() string {
	return "station_group"
}
