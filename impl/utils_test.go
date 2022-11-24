package impl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/impl/pda"
)

func TestCodeParser(t *testing.T) {
	assert := assert.New(t)
	// not full data.
	{
		codes := codeParser(pda.Code{
			Code0:    "AVAILABLE",
			CodeDsc0: "AVAILABLE",
			Code1:    "HOLD",
			CodeDsc1: "HOLD",
		})
		assert.Equal([]*mcom.Code{
			{
				Code:            "AVAILABLE",
				CodeDescription: "AVAILABLE",
			},
			{
				Code:            "HOLD",
				CodeDescription: "HOLD",
			},
		}, codes)
	}
	// full data.
	{
		codes := codeParser(pda.Code{
			Code0:    "HDAR",
			CodeDsc0: "面積比不符",
			Code1:    "HDCL",
			CodeDsc1: "捲取不符",
			Code2:    "HDDG",
			CodeDsc2: "死膠",
			Code3:    "HDEP",
			CodeDsc3: "超日限",
			Code4:    "HDER",
			CodeDsc4: "外觀不符",
			Code5:    "HDFB",
			CodeDsc5: "異物",
			Code6:    "HDOT",
			CodeDsc6: "其他",
			Code7:    "HDTK",
			CodeDsc7: "厚度不符",
			Code8:    "HDWD",
			CodeDsc8: "寬度不符",
			Code9:    "HDWT",
			CodeDsc9: "重量不符",
		})
		assert.Equal([]*mcom.Code{
			{
				Code:            "HDAR",
				CodeDescription: "面積比不符",
			},
			{
				Code:            "HDCL",
				CodeDescription: "捲取不符",
			},
			{
				Code:            "HDDG",
				CodeDescription: "死膠",
			},
			{
				Code:            "HDEP",
				CodeDescription: "超日限",
			},
			{
				Code:            "HDER",
				CodeDescription: "外觀不符",
			},
			{
				Code:            "HDFB",
				CodeDescription: "異物",
			},
			{
				Code:            "HDOT",
				CodeDescription: "其他",
			},
			{
				Code:            "HDTK",
				CodeDescription: "厚度不符",
			},
			{
				Code:            "HDWD",
				CodeDescription: "寬度不符",
			},
			{
				Code:            "HDWT",
				CodeDescription: "重量不符",
			},
		}, codes)
	}
}

func TestStatusParser(t *testing.T) {
	assert := assert.New(t)
	// not full data.
	{
		codes := statusParser(&pda.Material{
			CodeExt0: "AVAL",
			CodeDsc0: "AVAL",
			CodeExt1: "HOLD",
			CodeDsc1: "HOLD",
		})
		assert.Equal([]*mcom.Code{
			{
				Code:            "HOLD",
				CodeDescription: "HOLD",
			},
			{
				Code:            "AVAL",
				CodeDescription: "AVAL",
			},
		}, codes)
	}
	// full data status: AVAL.
	{
		codes := statusParser(&pda.Material{
			Status:   "AVAL",
			CodeExt0: "ADD",
			CodeDsc0: "ADD",
			CodeExt1: "AVAL",
			CodeDsc1: "AVAL",
			CodeExt2: "HOLD",
			CodeDsc2: "HOLD",
			CodeExt3: "MONT",
			CodeDsc3: "MONT",
			CodeExt4: "NAVL",
			CodeDsc4: "NAVL",
			CodeExt5: "SHIP",
			CodeDsc5: "SHIP",
			CodeExt6: "TEST",
			CodeDsc6: "TEST",
		})
		assert.Equal([]*mcom.Code{
			{
				Code:            "AVAL",
				CodeDescription: "AVAL",
			},
			{
				Code:            "ADD",
				CodeDescription: "ADD",
			},
			{
				Code:            "HOLD",
				CodeDescription: "HOLD",
			},
			{
				Code:            "MONT",
				CodeDescription: "MONT",
			},
			{
				Code:            "NAVL",
				CodeDescription: "NAVL",
			},
			{
				Code:            "SHIP",
				CodeDescription: "SHIP",
			},
			{
				Code:            "TEST",
				CodeDescription: "TEST",
			},
		}, codes)
	}
	// case status: INSP.
	{
		codes := statusParser(&pda.Material{
			Status:   "INSP",
			CodeExt0: "AVAL",
			CodeDsc0: "AVAL",
			CodeExt1: "HOLD",
			CodeDsc1: "HOLD",
		})
		assert.Equal([]*mcom.Code{
			{
				Code:            "AVAL",
				CodeDescription: "AVAL",
			},
			{
				Code:            "HOLD",
				CodeDescription: "HOLD",
			},
		}, codes)
	}
	// case statue: MONT.
	{
		codes := statusParser(&pda.Material{
			Status:   "MONT",
			CodeExt0: "AVAL",
			CodeDsc0: "AVAL",
			CodeExt1: "HOLD",
			CodeDsc1: "HOLD",
		})
		assert.Equal([]*mcom.Code{
			{
				Code:            "AVAL",
				CodeDescription: "AVAL",
			},
			{
				Code:            "HOLD",
				CodeDescription: "HOLD",
			},
		}, codes)
	}
	// case status: HOLD.
	{
		codes := statusParser(&pda.Material{
			Status:   "HOLD",
			CodeExt0: "ADD",
			CodeDsc0: "ADD",
			CodeExt1: "AVAL",
			CodeDsc1: "AVAL",
			CodeExt2: "NAVL",
			CodeDsc2: "NAVL",
			CodeExt3: "REWK",
			CodeDsc3: "REWK",
		})
		assert.Equal([]*mcom.Code{
			{
				Code:            "AVAL",
				CodeDescription: "AVAL",
			},
			{
				Code:            "ADD",
				CodeDescription: "ADD",
			},
			{
				Code:            "NAVL",
				CodeDescription: "NAVL",
			},
			{
				Code:            "REWK",
				CodeDescription: "REWK",
			},
		}, codes)
	}
}
