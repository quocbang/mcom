package mcom

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func Test_NewCarrierInfo(t *testing.T) {
	assert := assert.New(t)

	now := types.ToTimeNano(time.Now())
	actual := NewCarrierInfo(models.Carrier{
		IDPrefix:        "AA",
		SerialNumber:    1,
		DepartmentOID:   "department",
		AllowedMaterial: "am",
		Deprecated:      false,
		Contents:        []string{"Hello", "World"},
		UpdatedAt:       now,
		UpdatedBy:       "Rockefeller",
		CreatedAt:       now,
		CreatedBy:       "Rockefeller",
	})
	expected := CarrierInfo{
		ID:              "AA0001",
		AllowedMaterial: "am",
		Contents:        []string{"Hello", "World"},
		UpdateBy:        "Rockefeller",
		UpdateAt:        now.Time(),
	}

	assert.Equal(expected, actual)
}

// for coverate.
func Test_implementUpdateCarrierAction(t *testing.T) {
	UpdateProperties{}.implementUpdateCarrierAction()
	BindResources{}.implementUpdateCarrierAction()
	RemoveResources{}.implementUpdateCarrierAction()
	ClearResources{}.implementUpdateCarrierAction()
}
