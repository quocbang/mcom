package mcom

import "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"

type ResourceBindRequest interface {
	GetDetails() []BindRequestDetail
}

type BindRequestDetail interface {
	GetUniqueSite() models.UniqueSite
}
