package impl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

func TestDatamanager_ListCarriers(t *testing.T) {
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	assert := assert.New(t)
	resetCarrierTestStatus(db, assert)
	{ // case 1:normal case (should pass)
		assert.NoError(dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{
			DepartmentOID:   "ABC",
			IDPrefix:        "AA",
			Quantity:        3,
			AllowedMaterial: "AM1",
		}))
		actual, err := dm.ListCarriers(ctx, mcom.ListCarriersRequest{DepartmentOID: "ABC"})
		assert.NoError(err)
		expected := []mcom.CarrierInfo{{
			ID:              "AA0001",
			AllowedMaterial: "AM1",
			Contents:        []string{},
		}, {
			ID:              "AA0002",
			AllowedMaterial: "AM1",
			Contents:        []string{},
		}, {
			ID:              "AA0003",
			AllowedMaterial: "AM1",
			Contents:        []string{},
		}}
		assert.True(assertListCarriersReplies(assert, expected, actual.Info))
	}
	{ // case 2: empty departmentOID (should not pass)
		_, actual := dm.ListCarriers(ctx, mcom.ListCarriersRequest{DepartmentOID: ""})
		assert.Equal(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "department OID is required",
		}, actual)
	}
	{ // case 3: carrier not found (should not pass)
		actual, err := dm.ListCarriers(ctx, mcom.ListCarriersRequest{DepartmentOID: "notFound"})
		assert.NoError(err)
		assert.Equal(mcom.ListCarriersReply{Info: []mcom.CarrierInfo{}}, actual)
	}
	{ // case 4:normal case test order (should pass)
		assert.NoError(dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{
			DepartmentOID:   "XYZ",
			IDPrefix:        "BB",
			Quantity:        3,
			AllowedMaterial: "AM1",
		}))
		// we attempt to make the order messy.
		assert.NoError(dm.UpdateCarrier(ctx, mcom.UpdateCarrierRequest{
			ID: "BB0002",
			Action: mcom.UpdateProperties{
				AllowedMaterial: "update it",
			},
		}))
		actual, err := dm.ListCarriers(ctx, mcom.ListCarriersRequest{DepartmentOID: "XYZ"}.
			WithOrder(
				mcom.Order{
					Name:       "id_prefix",
					Descending: false,
				}, mcom.Order{
					Name:       "serial_number",
					Descending: false,
				},
			),
		)
		assert.NoError(err)
		expected := []mcom.CarrierInfo{{
			ID:              "BB0001",
			AllowedMaterial: "AM1",
			Contents:        []string{},
		}, {
			ID:              "BB0002",
			AllowedMaterial: "update it",
			Contents:        []string{},
		}, {
			ID:              "BB0003",
			AllowedMaterial: "AM1",
			Contents:        []string{},
		}}
		assert.True(assertListCarriersReplies(assert, expected, actual.Info))
	}
	{ // case 4:normal case test pagination and order (should pass)
		actual, err := dm.ListCarriers(ctx, mcom.ListCarriersRequest{DepartmentOID: "XYZ"}.WithOrder(
			[]mcom.Order{{
				Name:       "id_prefix",
				Descending: false,
			}, {
				Name:       "serial_number",
				Descending: false,
			}}...,
		).WithPagination(mcom.PaginationRequest{
			PageCount:      1,
			ObjectsPerPage: 2,
		}))
		assert.NoError(err)
		expected := []mcom.CarrierInfo{{
			ID:              "BB0001",
			AllowedMaterial: "AM1",
			Contents:        []string{},
		}, {
			ID:              "BB0002",
			AllowedMaterial: "update it",
			Contents:        []string{},
		}}
		assert.True(assertListCarriersReplies(assert, expected, actual.Info))
		assert.Equal(actual.AmountOfData, int64(3))
	}
	resetCarrierTestStatus(db, assert)
}

func TestDatamanager_GetCarrier(t *testing.T) {
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	assert := assert.New(t)
	resetCarrierTestStatus(db, assert)
	{ // insufficient request.
		_, err := dm.GetCarrier(ctx, mcom.GetCarrierRequest{})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "carrier ID is required"})
	}
	{ // carrier not found.
		_, err := dm.GetCarrier(ctx, mcom.GetCarrierRequest{ID: "NO1234"})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_CARRIER_NOT_FOUND})
	}
	{ // good case.
		assert.NoError(dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{
			DepartmentOID:   "AA",
			IDPrefix:        "AA",
			Quantity:        5,
			AllowedMaterial: "AM",
		}))
		actual, err := dm.GetCarrier(ctx, mcom.GetCarrierRequest{ID: "AA0002"})
		assert.NoError(err)
		expected := mcom.GetCarrierReply{
			ID:              "AA0002",
			AllowedMaterial: "AM",
			Contents:        []string{},
			UpdateAt:        actual.UpdateAt,
		}
		assert.Equal(expected, actual)
	}
	{ // wrong format.
		_, err := dm.GetCarrier(ctx, mcom.GetCarrierRequest{ID: "aa99999"})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_CARRIER_NOT_FOUND})
	}
	{ // itoa error.
		_, err := dm.GetCarrier(ctx, mcom.GetCarrierRequest{ID: "aa123g"})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_CARRIER_NOT_FOUND})
	}

	resetCarrierTestStatus(db, assert)
}

func TestDatamanager_CreateCarrier(t *testing.T) {
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	assert := assert.New(t)
	resetCarrierTestStatus(db, assert)
	{ // case 1:normal case (Should pass)
		assert.NoError(dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{
			DepartmentOID:   "ABC",
			IDPrefix:        "AA",
			Quantity:        3,
			AllowedMaterial: "AM1",
		}))
		expected := []models.Carrier{{
			IDPrefix:        "AA",
			SerialNumber:    1,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "AA",
			SerialNumber:    2,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "AA",
			SerialNumber:    3,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}}
		actual := getAllCarriers(db)
		assert.True(assertCarriers(t, expected, actual))
	}
	{ // case 2:normal case (Should pass)
		assert.NoError(dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{
			DepartmentOID:   "ABC",
			IDPrefix:        "AA",
			Quantity:        2,
			AllowedMaterial: "AM1",
		}))
		expected := []models.Carrier{{
			IDPrefix:        "AA",
			SerialNumber:    1,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "AA",
			SerialNumber:    2,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "AA",
			SerialNumber:    3,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "AA",
			SerialNumber:    4,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "AA",
			SerialNumber:    5,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}}
		actual := getAllCarriers(db)
		assert.True(assertCarriers(t, expected, actual))
	}
	{ // case 3:normal case (Should pass)
		assert.NoError(dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{
			DepartmentOID:   "ABC",
			IDPrefix:        "BB",
			Quantity:        2,
			AllowedMaterial: "AM1",
		}))
		expected := []models.Carrier{{
			IDPrefix:        "AA",
			SerialNumber:    1,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "AA",
			SerialNumber:    2,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "AA",
			SerialNumber:    3,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "AA",
			SerialNumber:    4,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "AA",
			SerialNumber:    5,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "BB",
			SerialNumber:    1,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "BB",
			SerialNumber:    2,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}}
		actual := getAllCarriers(db)
		assert.True(assertCarriers(t, expected, actual))
	}
	{ // case 4:empty departmentOID (should not pass)
		actual := dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{IDPrefix: "NN", Quantity: 1})
		assert.Equal(mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "illegal CreateCarrierRequest"}, actual)
	}
	{ // case 5:illegal idPrefix (should not pass)
		actual := dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{DepartmentOID: "ABC", IDPrefix: "NNNNNNN", Quantity: 1})
		assert.Equal(mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "illegal CreateCarrierRequest"}, actual)
	}
	{ // case 6:empty quantity (should not pass)
		actual := dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{DepartmentOID: "ABC", IDPrefix: "NN"})
		assert.Equal(mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "illegal CreateCarrierRequest"}, actual)
	}
	{ // case 7:max value reached (should not pass)
		// ! This case takes too much time to be in the automated test.
		// actual := dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{DepartmentOID: "ABC", IDPrefix: "AA", Quantity: 10000})
		// expected := mcomErr.Error{Code: mcomErr.Code_CARRIER_QUANTITY_LIMIT}
		// assert.Equal( expected, actual)
	}
	resetCarrierTestStatus(db, assert)
}

func TestDatamanager_UpdateCarrier(t *testing.T) {
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	assert := assert.New(t)
	resetCarrierTestStatus(db, assert)
	{ // normal case (Should pass)
		assert.NoError(dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{
			DepartmentOID:   "ABC",
			IDPrefix:        "AA",
			Quantity:        3,
			AllowedMaterial: "AM1",
		}))
		assert.NoError(dm.UpdateCarrier(ctx, mcom.UpdateCarrierRequest{
			ID: "AA0002",
			Action: mcom.UpdateProperties{
				AllowedMaterial: "AM2",
				DepartmentOID:   "DEF",
			},
		}))
		expected := []models.Carrier{{
			IDPrefix:        "AA",
			SerialNumber:    1,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}, {
			IDPrefix:        "AA",
			SerialNumber:    2,
			DepartmentOID:   "DEF",
			AllowedMaterial: "AM2",
			Deprecated:      false,
		}, {
			IDPrefix:        "AA",
			SerialNumber:    3,
			DepartmentOID:   "ABC",
			AllowedMaterial: "AM1",
			Deprecated:      false,
		}}
		actual := getAllCarriers(db)
		assert.Equal(true, assertCarriers(t, expected, actual))
	}
	{ // deprecated carrier (Should not pass)
		assert.NoError(dm.DeleteCarrier(ctx, mcom.DeleteCarrierRequest{ID: "AA0001"}))
		actual := dm.UpdateCarrier(ctx, mcom.UpdateCarrierRequest{
			ID: "AA0001",
			Action: mcom.UpdateProperties{
				AllowedMaterial: "AM2",
			},
		})
		assert.Equal(mcomErr.Error{Code: mcomErr.Code_CARRIER_NOT_FOUND}, actual)
	}
	{ // carrier not exists (Should not pass)
		actual := dm.UpdateCarrier(ctx, mcom.UpdateCarrierRequest{
			ID: "BB0001",
			Action: mcom.UpdateProperties{
				AllowedMaterial: "AM2",
			},
		})
		assert.Equal(mcomErr.Error{Code: mcomErr.Code_CARRIER_NOT_FOUND}, actual)
	}
	{ // empty id (should not pass)
		actual := dm.UpdateCarrier(ctx, mcom.UpdateCarrierRequest{
			ID: "",
		})
		assert.Equal(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "id is required",
		}, actual)
	}
	{ // update allowed material and do not update department.
		assert.NoError(dm.UpdateCarrier(ctx, mcom.UpdateCarrierRequest{
			ID: "AA0003",
			Action: mcom.UpdateProperties{
				AllowedMaterial: "",
				DepartmentOID:   "",
			},
		}))

		var carrier models.Carrier
		assert.NoError(db.Where(&models.Carrier{IDPrefix: "AA", SerialNumber: 3}).Take(&carrier).Error)

		assert.Equal("ABC", carrier.DepartmentOID)
		assert.Equal("", carrier.AllowedMaterial)
	}

	// #region contents
	resetCarrierTestStatus(db, assert)
	assert.NoError(dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{
		DepartmentOID:   "DOID",
		IDPrefix:        "AA",
		Quantity:        1,
		AllowedMaterial: "AM1",
	}))
	{ // add resources.
		assert.NoError(dm.UpdateCarrier(ctx, mcom.UpdateCarrierRequest{
			ID: "AA0001",
			Action: mcom.BindResources{
				ResourcesID: []string{"RID1", "RID2", "RID3"},
			},
		}))
		rep, err := dm.GetCarrier(ctx, mcom.GetCarrierRequest{
			ID: "AA0001",
		})
		assert.NoError(err)
		assert.Equal([]string{"RID1", "RID2", "RID3"}, rep.Contents)
	}
	{ // remove resources.
		assert.NoError(dm.UpdateCarrier(ctx, mcom.UpdateCarrierRequest{
			ID: "AA0001",
			Action: mcom.RemoveResources{
				ResourcesID: []string{"RID1", "RID3"},
			},
		}))
		rep, err := dm.GetCarrier(ctx, mcom.GetCarrierRequest{
			ID: "AA0001",
		})
		assert.NoError(err)
		assert.Equal([]string{"RID2"}, rep.Contents)
	}
	{ // clear resources.
		assert.NoError(dm.UpdateCarrier(ctx, mcom.UpdateCarrierRequest{
			ID:     "AA0001",
			Action: mcom.ClearResources{},
		}))
		rep, err := dm.GetCarrier(ctx, mcom.GetCarrierRequest{
			ID: "AA0001",
		})
		assert.NoError(err)
		assert.Equal([]string{}, rep.Contents)
	}
	// #endregion contents
	resetCarrierTestStatus(db, assert)
}

func TestDatamanager_DeleteCarrier(t *testing.T) {
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	assert := assert.New(t)
	resetCarrierTestStatus(db, assert)
	{ // case 1:normal case (should pass)
		assert.NoError(dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{
			DepartmentOID:   "ABC",
			IDPrefix:        "AA",
			Quantity:        3,
			AllowedMaterial: "AM1",
		}))
		assert.NoError(dm.DeleteCarrier(ctx, mcom.DeleteCarrierRequest{ID: "AA0001"}))
		actual, err := dm.ListCarriers(ctx, mcom.ListCarriersRequest{DepartmentOID: "ABC"})
		assert.NoError(err)
		expected := []mcom.CarrierInfo{{
			ID:              "AA0002",
			AllowedMaterial: "AM1",
			Contents:        []string{},
		}, {
			ID:              "AA0003",
			AllowedMaterial: "AM1",
			Contents:        []string{},
		}}
		assert.True(assertListCarriersReplies(assert, expected, actual.Info))
	}
	{ // case 2: empty id (should not pass)
		actual := dm.DeleteCarrier(ctx, mcom.DeleteCarrierRequest{ID: ""})
		assert.Equal(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "id is required",
		}, actual)
	}
	{ // case 1:delete again (should not pass)
		assert.ErrorIs(dm.DeleteCarrier(ctx, mcom.DeleteCarrierRequest{ID: "AA0001"}), mcomErr.Error{Code: mcomErr.Code_CARRIER_NOT_FOUND})
		actual, err := dm.ListCarriers(ctx, mcom.ListCarriersRequest{DepartmentOID: "ABC"})
		assert.NoError(err)
		expected := []mcom.CarrierInfo{{
			ID:              "AA0002",
			AllowedMaterial: "AM1",
			Contents:        []string{},
		}, {
			ID:              "AA0003",
			AllowedMaterial: "AM1",
			Contents:        []string{},
		}}
		assert.True(assertListCarriersReplies(assert, expected, actual.Info))
	}
	resetCarrierTestStatus(db, assert)
}

func getAllCarriers(db *gorm.DB) (res []models.Carrier) {
	if err := db.Find(&res).Error; err != nil {
		return []models.Carrier{}
	}
	return
}

func assertCarriers(t *testing.T, expected, actual []models.Carrier) bool {
	assert := assert.New(t)
	if !assert.Equal(len(expected), len(actual)) {
		return false
	}
	expectedMap := make(map[struct {
		idPrefix string
		sn       int32
	}]models.Carrier, len(expected))
	for _, elem := range expected {
		expectedMap[struct {
			idPrefix string
			sn       int32
		}{elem.IDPrefix, elem.SerialNumber}] = elem
	}
	for _, elem := range actual {
		exp := expectedMap[struct {
			idPrefix string
			sn       int32
		}{elem.IDPrefix, elem.SerialNumber}]
		if !(assert.Equal(exp.IDPrefix, elem.IDPrefix) &&
			assert.Equal(exp.SerialNumber, elem.SerialNumber) &&
			assert.Equal(exp.DepartmentOID, elem.DepartmentOID) &&
			assert.Equal(exp.AllowedMaterial, elem.AllowedMaterial) &&
			assert.Equal(exp.Deprecated, elem.Deprecated)) {
			return false
		}
	}
	return true
}

func assertListCarriersReplies(assert *assert.Assertions, expected, actual []mcom.CarrierInfo) bool {
	parseListCarriersReplies := func(r []mcom.CarrierInfo) []mcom.CarrierInfo {
		s := make([]mcom.CarrierInfo, len(r))
		for i, ele := range r {
			s[i] = mcom.CarrierInfo{
				ID:              ele.ID,
				AllowedMaterial: ele.AllowedMaterial,
				Contents:        ele.Contents,
			}
		}
		return s
	}

	return assert.Equal(parseListCarriersReplies(expected), parseListCarriersReplies(actual))
}

func resetCarrierTestStatus(db *gorm.DB, assert *assert.Assertions) {
	clearCarrier(db, assert)
	clearCarrierSerial(db, assert)
	resetSequence(db, assert)
}

func clearCarrier(db *gorm.DB, assert *assert.Assertions) {
	assert.NoError(newClearMaster(db, &models.Carrier{}).Clear())
}

func clearCarrierSerial(db *gorm.DB, assert *assert.Assertions) {
	assert.NoError(db.Where("1=1").Delete(&models.CarrierSerial{}).Error, "gorm function error")
	assert.NoError(newClearMaster(db, &models.CarrierSerial{}).Clear())
}

func resetSequence(db *gorm.DB, assert *assert.Assertions) {
	// the sequence for the AA prefix.
	assert.NoError(db.Exec("DROP SEQUENCE IF EXISTS carrier_seq_65_65;").Error)
	// the sequence for the BB prefix.
	assert.NoError(db.Exec("DROP SEQUENCE IF EXISTS carrier_seq_66_66;").Error)
}
