package impl

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

func TestDataManager_CreatePackRecords(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &models.PackRecord{})
	assert.NoError(cm.Clear())
	assert.NoError(db.Exec("ALTER SEQUENCE pack_record_serial_number_seq RESTART;").Error)
	{ // ordinary case (should pass)
		assert.NoError(dm.CreatePackRecords(ctx, mcom.CreatePackRecordsRequest{
			Packs: []mcom.PackReq{{
				Packing:      "Hamlet",
				PackNumber:   5,
				Quantity:     50,
				ActualWeight: decimal.NewFromFloat(250.0),
			}},
			Station: "Elsinore",
		}))

		var actual models.PackRecord
		assert.NoError(db.Where(&models.PackRecord{Packing: "Hamlet"}).Take(&actual).Error)
		expected := models.PackRecord{
			SerialNumber: 1,
			Packing:      "Hamlet",
			PackNumber:   5,
			Quantity:     50,
			ActualWeight: decimal.New(250000000, -6),
			Station:      "Elsinore",
			CreatedAt:    actual.CreatedAt,
			CreatedBy:    actual.CreatedBy,
		}

		assert.Equal(expected, actual)
	}
	{ // empty request (should pass)
		assert.NoError(dm.CreatePackRecords(ctx, mcom.CreatePackRecordsRequest{Packs: []mcom.PackReq{{}}}))

		var actual models.PackRecord
		assert.NoError(db.Where(`packing = ''`).Take(&actual).Error)
		expected := models.PackRecord{
			SerialNumber: 2,
			ActualWeight: decimal.New(0, -6),
			CreatedAt:    actual.CreatedAt,
			CreatedBy:    actual.CreatedBy,
		}

		assert.Equal(expected, actual)
	}
	{ // create multiple records.
		assert.NoError(dm.CreatePackRecords(ctx, mcom.CreatePackRecordsRequest{
			Station: "station",
			Packs: []mcom.PackReq{
				{
					Packing:      "3",
					PackNumber:   3,
					Quantity:     3,
					ActualWeight: decimal.NewFromInt(3),
					Timestamp:    3,
				},
				{
					Packing:      "4",
					PackNumber:   4,
					Quantity:     4,
					ActualWeight: decimal.NewFromInt(4),
					Timestamp:    4,
				},
			},
		}))

		actual, err := dm.ListPackRecords(ctx)
		assert.NoError(err)

		expected := mcom.ListPackRecordsReply{
			Packs: []mcom.Pack{
				{
					SerialNumber: 1,
					Packing:      "Hamlet",
					PackNumber:   5,
					Quantity:     50,
					ActualWeight: decimal.New(250000000, -6),
					Station:      "Elsinore",
					CreatedAt:    actual.Packs[0].CreatedAt,
					CreatedBy:    actual.Packs[0].CreatedBy,
				},
				{
					SerialNumber: 2,
					ActualWeight: decimal.New(0, -6),
					CreatedAt:    actual.Packs[1].CreatedAt,
					CreatedBy:    actual.Packs[1].CreatedBy,
				},
				{
					SerialNumber: 3,
					Packing:      "3",
					PackNumber:   3,
					Quantity:     3,
					ActualWeight: decimal.RequireFromString("3.000000"),
					Station:      "station",
					CreatedAt:    actual.Packs[2].CreatedAt,
					CreatedBy:    actual.Packs[2].CreatedBy,
				},
				{
					SerialNumber: 4,
					Packing:      "4",
					PackNumber:   4,
					Quantity:     4,
					ActualWeight: decimal.RequireFromString("4.000000"),
					Station:      "station",
					CreatedAt:    actual.Packs[3].CreatedAt,
					CreatedBy:    actual.Packs[3].CreatedBy,
				},
			},
		}
		assert.Equal(expected, actual)
	}
	assert.NoError(cm.Clear())
	assert.NoError(db.Exec("ALTER SEQUENCE pack_record_serial_number_seq RESTART;").Error)
}

func TestDataManager_ListPackRecords(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &models.PackRecord{})
	assert.NoError(cm.Clear())
	assert.NoError(db.Exec("ALTER SEQUENCE pack_record_serial_number_seq RESTART;").Error)
	{
		assert.NoError(dm.CreatePackRecords(ctx, mcom.CreatePackRecordsRequest{
			Packs: []mcom.PackReq{
				{
					Packing:      "p1",
					PackNumber:   1,
					Quantity:     1,
					ActualWeight: decimal.NewFromInt(1),
					Timestamp:    1,
				},
			},
			Station: "1",
		}))
		assert.NoError(dm.CreatePackRecords(ctx, mcom.CreatePackRecordsRequest{
			Packs: []mcom.PackReq{
				{
					Packing:      "p2",
					PackNumber:   2,
					Quantity:     2,
					ActualWeight: decimal.NewFromInt(2),
					Timestamp:    2,
				},
			},
			Station: "2",
		}))
		assert.NoError(dm.CreatePackRecords(ctx, mcom.CreatePackRecordsRequest{
			Packs: []mcom.PackReq{
				{
					Packing:      "p3",
					PackNumber:   3,
					Quantity:     3,
					ActualWeight: decimal.NewFromInt(3),
					Timestamp:    3,
				},
			},
			Station: "3",
		}))
		assert.NoError(dm.CreatePackRecords(ctx, mcom.CreatePackRecordsRequest{
			Packs: []mcom.PackReq{
				{
					Packing:      "p4",
					PackNumber:   4,
					Quantity:     4,
					ActualWeight: decimal.NewFromInt(4),
					Timestamp:    4,
				},
			},
			Station: "4",
		}))
		assert.NoError(dm.CreatePackRecords(ctx, mcom.CreatePackRecordsRequest{
			Packs: []mcom.PackReq{
				{
					Packing:      "p5",
					PackNumber:   5,
					Quantity:     5,
					ActualWeight: decimal.NewFromInt(5),
					Timestamp:    5,
				},
			},
			Station: "5",
		}))

		actual, err := dm.ListPackRecords(ctx)
		assert.NoError(err)
		expected := mcom.ListPackRecordsReply{
			Packs: []mcom.Pack{{
				SerialNumber: 1,
				Packing:      "p1",
				PackNumber:   1,
				Quantity:     1,
				ActualWeight: decimal.New(1000000, -6),
				Station:      "1",
				CreatedAt:    1,
			}, {
				SerialNumber: 2,
				Packing:      "p2",
				PackNumber:   2,
				Quantity:     2,
				ActualWeight: decimal.New(2000000, -6),
				Station:      "2",
				CreatedAt:    2,
			}, {
				SerialNumber: 3,
				Packing:      "p3",
				PackNumber:   3,
				Quantity:     3,
				ActualWeight: decimal.New(3000000, -6),
				Station:      "3",
				CreatedAt:    3,
			}, {
				SerialNumber: 4,
				Packing:      "p4",
				PackNumber:   4,
				Quantity:     4,
				ActualWeight: decimal.New(4000000, -6),
				Station:      "4",
				CreatedAt:    4,
			}, {
				SerialNumber: 5,
				Packing:      "p5",
				PackNumber:   5,
				Quantity:     5,
				ActualWeight: decimal.New(5000000, -6),
				Station:      "5",
				CreatedAt:    5,
			}},
		}

		assert.Equal(expected, actual)
	}
	assert.NoError(cm.Clear())
	assert.NoError(db.Exec("ALTER SEQUENCE pack_record_serial_number_seq RESTART;").Error)
}
