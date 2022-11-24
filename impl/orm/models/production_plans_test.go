package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/callbacks"
)

func TestProductionPlan_BeforeCreate(t *testing.T) {
	assert := assert.New(t)

	assert.Implements(new(callbacks.BeforeCreateInterface), new(ProductionPlan))

	var pp ProductionPlan
	assert.NoError(pp.BeforeCreate(nil))
	assert.NotEmpty(pp.OID)
}
