package mcom

import "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"

type Orderable interface {
	// WithOrder(...Order)self.
	NeedOrder() bool
	ValidateOrder() error
	GetOrder() []Order
}

// OrderRequest.
// replaces the default order with the specified fields.
//
// field names are according to the comments of the DataManager methods.
type OrderRequest struct {
	OrderBy []Order
}

type Order struct {
	Name       string
	Descending bool
}

func isValidName[T models.Model](name string) bool {
	validNames := staticOrderableFieldsManager.get(*new(T))
	_, ok := validNames[name]
	return ok
}

type OrderableField struct {
	FieldName  string
	ColumnName string
}
