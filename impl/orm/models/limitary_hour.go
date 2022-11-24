package models

// Limitary.
//
// A product of {ProductType} can only be used after standing for {Min} hours and it will be expired after {Max} hours.
//
// For Example.
// There is a {LimitaryHour} record in DB.
//
// {
// 		ProductType = "TypeA",
// 		Min         = 24,
// 		Max         = 24,
// }
//
// productA (TypeA) which was created at 2022/7/21 08:00 can only be used during 2022/7/22 08:00 ~ 2022/7/23 08:00.
type LimitaryHour struct {
	ProductType string `gorm:"primaryKey"`
	Min         int32  `gorm:"not null"` // Rest time, time unit is hour.
	Max         int32  `gorm:"not null"` // valid time, time unit is hour.
}

// Model implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (*LimitaryHour) Model() {}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (*LimitaryHour) TableName() string {
	return "limitary_hour"
}
