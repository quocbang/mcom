package impl

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/impl/pda"
)

type mockGetCode struct {
	result string
}

const (
	failed    = "failed"
	txError   = "txError"
	getReason = "getReason"
	getStatus = "getStatus"
	getArea   = "getArea"
)

func (s *mockGetCode) GetCodeFromBRM(codeCate string) (pda.Code, error) {
	switch s.result {
	case failed:
		return pda.Code{}, fmt.Errorf("web service error")
	case txError:
		return pda.Code{ReturnMessage: "tx error"}, nil
	case getReason:
		return pda.Code{
			Code0:    "HDAR",
			CodeDsc0: "面積比不符",
			Code1:    "HDCL",
			CodeDsc1: "捲取不符",
			Code2:    "HDEP",
			CodeDsc2: "超日限",
			Code3:    "HDOT",
			CodeDsc3: "其他",
		}, nil
	case getArea:
		return pda.Code{
			Code0:    "KUBB",
			CodeDsc0: "建大雲林廠-密煉",
			Code1:    "KUSH",
			CodeDsc1: "建大雲林廠-成型",
		}, nil
	}
	return pda.Code{}, nil
}

func (s *mockGetCode) GetMaterialsInfo(materialID string) (*pda.Material, error) {
	switch s.result {
	case failed:
		return &pda.Material{}, fmt.Errorf("web service error")
	case txError:
		return &pda.Material{ReturnMessage: "tx error"}, nil
	case getStatus:
		return &pda.Material{
			CodeExt0:      "AVAL",
			CodeDsc0:      "AVAL",
			CodeExt1:      "HOLD",
			CodeDsc1:      "HOLD",
			CodeExt2:      "TEST",
			CodeDsc2:      "TEST",
			ReturnMessage: "",
		}, nil
	}
	return nil, nil
}

func (s *mockGetCode) UpdateMaterials(reqXML string) (string, error) {
	return "", nil
}

func TestListControlReasons(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	// failed to get code from BRM.
	{
		mockPDA := &mockGetCode{result: failed}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.ListControlReasons(ctx)
		assert.EqualError(err, "web service error")
	}
	// tx error.
	{
		mockPDA := &mockGetCode{result: txError}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.ListControlReasons(ctx)
		assert.EqualError(err, "failed to get control reason code, tx err: tx error")
	}
	// normal.
	{
		mockPDA := &mockGetCode{result: getReason}
		dm := DataManager{
			pdaService: mockPDA,
		}

		codes, err := dm.ListControlReasons(ctx)
		assert.NoError(err)
		assert.Equal(mcom.ListControlReasonsReply{
			Codes: []*mcom.Code{
				{
					Code:            "HDAR",
					CodeDescription: "面積比不符",
				},
				{
					Code:            "HDCL",
					CodeDescription: "捲取不符",
				},
				{
					Code:            "HDEP",
					CodeDescription: "超日限",
				},
				{
					Code:            "HDOT",
					CodeDescription: "其他",
				},
			},
		}, codes)
	}
}

func TestListChangeableStatus(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	// failed to get code from BRM.
	{
		mockPDA := &mockGetCode{result: failed}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.ListChangeableStatus(ctx, mcom.ListChangeableStatusRequest{MaterialID: "test_barcode"})
		assert.EqualError(err, "web service error")
	}
	// tx error.
	{
		mockPDA := &mockGetCode{result: txError}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.ListChangeableStatus(ctx, mcom.ListChangeableStatusRequest{MaterialID: "test_barcode"})
		assert.EqualError(err, "failed to get changeable status code, err: tx error")
	}
	// normal.
	{
		mockPDA := &mockGetCode{result: getStatus}
		dm := DataManager{
			pdaService: mockPDA,
		}

		codes, err := dm.ListChangeableStatus(ctx, mcom.ListChangeableStatusRequest{MaterialID: "test_barcode"})
		assert.NoError(err)
		assert.Equal(mcom.ListChangeableStatusReply{
			Codes: []*mcom.Code{
				{
					Code:            "HOLD",
					CodeDescription: "HOLD",
				},
				{
					Code:            "AVAL",
					CodeDescription: "AVAL",
				},
				{
					Code:            "TEST",
					CodeDescription: "TEST",
				},
			},
		}, codes)
	}
}

func TestListControlAreas(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	// failed to get code from BRM.
	{
		mockPDA := &mockGetCode{result: failed}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.ListControlAreas(ctx)
		assert.EqualError(err, "web service error")
	}
	// tx error.
	{
		mockPDA := &mockGetCode{result: txError}
		dm := DataManager{
			pdaService: mockPDA,
		}

		_, err := dm.ListControlAreas(ctx)
		assert.EqualError(err, "failed to get control area code, err: tx error")
	}
	// normal.
	{
		mockPDA := &mockGetCode{result: getArea}
		dm := DataManager{
			pdaService: mockPDA,
		}

		codes, err := dm.ListControlAreas(ctx)
		assert.NoError(err)
		assert.Equal(mcom.ListControlAreasReply{
			Codes: []*mcom.Code{
				{
					Code:            "KUBB",
					CodeDescription: "建大雲林廠-密煉",
				},
				{
					Code:            "KUSH",
					CodeDescription: "建大雲林廠-成型",
				},
			},
		}, codes)
	}
}
