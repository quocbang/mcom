package impl

import (
	"context"
	"encoding/xml"
	"fmt"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/pda"
)

const (
	barcodeGetMaterialError = "barcodeGetMaterialError"
	barcodeTXError          = "barcodeTXError"
	barcodeWrongExpireDate  = "barcodeWrongExpireDate"
	barcodeNormalTest       = "barcodeNormalTest"
	barcodeNoExpandDate     = "barcodeNoExpandDate"
	barcodeWrongExpandDate  = "barcodeWrongExpandDate"
)

type mockGetMaterial struct{}

func (s *mockGetMaterial) GetCodeFromBRM(codeCate string) (pda.Code, error) {
	return pda.Code{}, nil
}

func (s *mockGetMaterial) GetMaterialsInfo(materialID string) (*pda.Material, error) {
	switch materialID {
	case barcodeGetMaterialError:
		return nil, fmt.Errorf("web service error")
	case barcodeTXError:
		return &pda.Material{ReturnMessage: "tx error"}, nil
	case barcodeWrongExpireDate:
		return &pda.Material{
			ExpireDate:    "wrongDate",
			ReturnMessage: "",
		}, nil
	case barcodeNormalTest:
		return &pda.Material{
			MaterialProductID: "test_product_id",
			MaterialID:        materialID,
			MaterialType:      "test_type",
			Sequence:          "001",
			Status:            "test",
			Quantity:          "1.23",
			Comment:           "test",
			ExpireDate:        "2020-02-20",
			SpreadDate:        "3",
			ReturnMessage:     "",
		}, nil
	case barcodeNoExpandDate:
		return &pda.Material{
			MaterialProductID: "test_product_id",
			MaterialID:        barcodeNoExpandDate,
			MaterialType:      "test",
			SpreadDate:        "",
		}, nil
	case barcodeWrongExpandDate:
		return &pda.Material{
			MaterialProductID: "test_product_id",
			MaterialID:        barcodeWrongExpandDate,
			MaterialType:      "test",
			SpreadDate:        "NaN",
		}, nil
	}
	return nil, nil
}

const (
	xmlStringFailedToUpdate = "xmlStringFailedToUpdate"
	xmlStringTxError        = "xmlStringTxError"
	xmlExceededTimes        = "xmlExceededTimes"
)

func (s *mockGetMaterial) UpdateMaterials(reqXML string) (string, error) {
	switch reqXML {
	case xmlStringFailedToUpdate:
		return "", fmt.Errorf("failed to update materials")
	case xmlStringTxError:
		return "tx error", nil
	case xmlExceededTimes:
		return "此物料已達可展延次數(2)", nil
	}
	return "", nil
}

func TestGetMaterial(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	// GetMaterialsInfo error.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.GetMaterial(ctx, mcom.GetMaterialRequest{MaterialID: barcodeGetMaterialError})
		assert.EqualError(err, "failed to get materials info, err: web service error")
	}
	// tx error.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.GetMaterial(ctx, mcom.GetMaterialRequest{MaterialID: barcodeTXError})
		assert.EqualError(err, "get material info err: tx error")
	}
	// materials not found.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.GetMaterial(ctx, mcom.GetMaterialRequest{MaterialID: ""})
		assert.EqualError(err, "RESOURCE_NOT_FOUND")
	}
	// parse date time error.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.GetMaterial(ctx, mcom.GetMaterialRequest{MaterialID: barcodeWrongExpireDate})
		assert.EqualError(err, "time parse error: parsing time \"wrongDate\" as \"2006-01-02\": cannot parse \"wrongDate\" as \"2006\"")
	}
	// normal.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		result, err := dm.GetMaterial(ctx, mcom.GetMaterialRequest{MaterialID: barcodeNormalTest})
		assert.NoError(err)

		expireDate, err := time.Parse("2006-01-02", "2020-02-20")
		assert.NoError(err)

		assert.Equal(mcom.GetMaterialReply{
			MaterialProductID: "test_product_id",
			MaterialID:        barcodeNormalTest,
			MaterialType:      "test_type",
			Sequence:          "001",
			Status:            "test",
			Quantity:          decimal.NewFromFloat(1.23),
			Comment:           "test",
			ExpireDate:        expireDate,
		}, result)
	}
}

func TestUpdateMaterial(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	// barcode not found.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		err := dm.UpdateMaterial(ctx, mcom.UpdateMaterialRequest{MaterialID: ""})
		assert.EqualError(err, "RESOURCE_NOT_FOUND")
	}
	// marshal error.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		monkey.Patch(xml.Marshal, func(v interface{}) ([]byte, error) {
			return []byte{}, fmt.Errorf("marshal error")
		})

		err := dm.UpdateMaterial(ctx, mcom.UpdateMaterialRequest{
			MaterialID: barcodeNormalTest,
		})
		monkey.Unpatch(xml.Marshal)
		assert.EqualError(err, "marshal error")
	}
	// failed to update materials.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		monkey.Patch(xml.Marshal, func(v interface{}) ([]byte, error) {
			return []byte(xmlStringFailedToUpdate), nil
		})

		err := dm.UpdateMaterial(ctx, mcom.UpdateMaterialRequest{
			MaterialID: barcodeNormalTest,
		})
		monkey.Unpatch(xml.Marshal)
		assert.EqualError(err, "failed to update materials, err: failed to update materials")
	}
	// tx error.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		monkey.Patch(xml.Marshal, func(v interface{}) ([]byte, error) {
			return []byte(xmlStringTxError), nil
		})

		err := dm.UpdateMaterial(ctx, mcom.UpdateMaterialRequest{
			MaterialID: barcodeNormalTest,
		})
		monkey.Unpatch(xml.Marshal)
		assert.EqualError(err, "failed to update materials, err: tx error")
	}
	// exceeded times error.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		monkey.Patch(xml.Marshal, func(v interface{}) ([]byte, error) {
			return []byte(xmlExceededTimes), nil
		})

		err := dm.UpdateMaterial(ctx, mcom.UpdateMaterialRequest{
			MaterialID: barcodeNormalTest,
		})
		monkey.Unpatch(xml.Marshal)
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_RESOURCE_CONTROL_ABOVE_EXTENDED_COUNT,
		}, err)
	}
	// normal.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		err := dm.UpdateMaterial(ctx, mcom.UpdateMaterialRequest{
			MaterialID:       barcodeNormalTest,
			ExtendedDuration: 72 * time.Hour,
			User:             "tester",
			NewStatus:        "TEST",
			Reason:           "test",
			ProductCate:      "test",
			ControlArea:      "test",
		})
		assert.NoError(err)
	}
}

func TestGetMaterialExtendDate(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	// GetMaterialsInfo error.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.GetMaterialExtendDate(ctx, mcom.GetMaterialExtendDateRequest{MaterialID: barcodeGetMaterialError})
		assert.EqualError(err, "web service error")
	}
	// tx error.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.GetMaterialExtendDate(ctx, mcom.GetMaterialExtendDateRequest{MaterialID: barcodeTXError})
		assert.EqualError(err, "failed to get changeable status code, err: tx error")
	}
	// materials not found.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.GetMaterialExtendDate(ctx, mcom.GetMaterialExtendDateRequest{MaterialID: ""})
		assert.EqualError(err, "RESOURCE_NOT_FOUND")
	}
	// unable to get expand date.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.GetMaterialExtendDate(ctx, mcom.GetMaterialExtendDateRequest{MaterialID: barcodeNoExpandDate})
		assert.EqualError(err, "unable to get expand date of type [test]")
	}
	// string convert error: NaN.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.GetMaterialExtendDate(ctx, mcom.GetMaterialExtendDateRequest{MaterialID: barcodeWrongExpandDate})
		assert.EqualError(err, "strconv.Atoi: parsing \"NaN\": invalid syntax")
	}
	// normal.
	{
		mockPDA := &mockGetMaterial{}
		dm := DataManager{
			pdaService: mockPDA,
		}

		extendDate, err := dm.GetMaterialExtendDate(ctx, mcom.GetMaterialExtendDateRequest{MaterialID: barcodeNormalTest})
		assert.NoError(err)
		assert.Equal(mcom.GetMaterialExtendDateReply(time.Duration(3)*24*time.Hour), extendDate)
	}
}
