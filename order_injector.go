package mcom

import (
	"fmt"
	"reflect"

	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

type orderableFieldsManager struct {
	// the first key : struct name, the second key : orderable field names.
	modelFieldPairs map[ /*struct name*/ string]map[ /*orderable field name*/ string]struct{}
}

func newOrderableFieldsManager() *orderableFieldsManager {
	return &orderableFieldsManager{
		modelFieldPairs: make(map[string]map[string]struct{}),
	}
}

func (ofm orderableFieldsManager) register(model models.Model, fieldColumnPairs ...OrderableField) error {
	m := reflect.TypeOf(model)
	fieldCount := m.NumField()
	validFieldNames := make(map[string]struct{}, fieldCount)

	for i := 0; i < fieldCount; i++ {
		validFieldNames[m.Field(i).Name] = struct{}{}
	}

	orderableFields := make(map[string]struct{})

	for _, of := range fieldColumnPairs {
		if _, ok := validFieldNames[of.FieldName]; !ok {
			return fmt.Errorf("field name not found, table: " + model.TableName() + ", field: " + of.FieldName)
		}

		orderableFields[of.ColumnName] = struct{}{}
	}

	ofm.modelFieldPairs[m.Name()] = orderableFields
	return nil
}

func (ofm orderableFieldsManager) get(model models.Model) map[string]struct{} {
	return ofm.modelFieldPairs[reflect.TypeOf(model).Name()]
}
