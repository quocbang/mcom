package mcom

import "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"

var staticOrderableFieldsManager *orderableFieldsManager

func init() {
	if err := registerStructures(); err != nil {
		panic(err)
	}
}

func registerStructures() error {
	staticOrderableFieldsManager = newOrderableFieldsManager()

	type toRegister struct {
		model  models.Model
		fields []OrderableField
	}

	defaultOrder := []toRegister{
		{
			model: models.Station{},
			fields: []OrderableField{
				{
					FieldName:  "ID",
					ColumnName: "id",
				}},
		},
		{
			model: models.Carrier{},
			fields: []OrderableField{
				{
					FieldName:  "IDPrefix",
					ColumnName: "id_prefix",
				}, {
					FieldName:  "SerialNumber",
					ColumnName: "serial_number",
				}},
		},
		{
			model: models.MaterialResource{},
			fields: []OrderableField{
				{
					FieldName:  "CreatedAt",
					ColumnName: "created_at",
				}},
		},
		{
			model: models.Recipe{},
			fields: []OrderableField{
				{
					FieldName:  "ID",
					ColumnName: "id",
				},
				{
					FieldName:  "ProductID",
					ColumnName: "product_id",
				},
				{
					FieldName:  "ReleasedAt",
					ColumnName: "released_at",
				},
				{
					FieldName:  "Stage",
					ColumnName: "stage",
				},
				{
					FieldName:  "Major",
					ColumnName: "major",
				},
				{
					FieldName:  "Minor",
					ColumnName: "minor",
				},
			},
		},
	}

	for _, order := range defaultOrder {
		if err := staticOrderableFieldsManager.register(order.model, order.fields...); err != nil {
			return err
		}
	}

	return nil
}
