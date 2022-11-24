package impl

import (
	"time"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/impl/pda"
)

const (
	dateLayout = "2006-01-02"
)

// parse date-time to date
func parseDate(t time.Time) (time.Time, error) {
	return time.Parse(dateLayout, t.Format(dateLayout))
}

// codeParser pares xml(pda.Code) to []*mcom.Code.
func codeParser(code pda.Code) []*mcom.Code {
	codes := []*mcom.Code{}
	if code.Code0 != "" {
		codes = append(codes, &mcom.Code{
			Code:            code.Code0,
			CodeDescription: code.CodeDsc0,
		})
	}
	if code.Code1 != "" {
		codes = append(codes, &mcom.Code{
			Code:            code.Code1,
			CodeDescription: code.CodeDsc1,
		})
	}
	if code.Code2 != "" {
		codes = append(codes, &mcom.Code{
			Code:            code.Code2,
			CodeDescription: code.CodeDsc2,
		})
	}
	if code.Code3 != "" {
		codes = append(codes, &mcom.Code{
			Code:            code.Code3,
			CodeDescription: code.CodeDsc3,
		})
	}
	if code.Code4 != "" {
		codes = append(codes, &mcom.Code{
			Code:            code.Code4,
			CodeDescription: code.CodeDsc4,
		})
	}
	if code.Code5 != "" {
		codes = append(codes, &mcom.Code{
			Code:            code.Code5,
			CodeDescription: code.CodeDsc5,
		})
	}
	if code.Code6 != "" {
		codes = append(codes, &mcom.Code{
			Code:            code.Code6,
			CodeDescription: code.CodeDsc6,
		})
	}
	if code.Code7 != "" {
		codes = append(codes, &mcom.Code{
			Code:            code.Code7,
			CodeDescription: code.CodeDsc7,
		})
	}
	if code.Code8 != "" {
		codes = append(codes, &mcom.Code{
			Code:            code.Code8,
			CodeDescription: code.CodeDsc8,
		})
	}
	if code.Code9 != "" {
		codes = append(codes, &mcom.Code{
			Code:            code.Code9,
			CodeDescription: code.CodeDsc9,
		})
	}
	return codes
}

func statusParser(material *pda.Material) []*mcom.Code {
	codes := []*mcom.Code{}

	if material.Status == "INSP" ||
		material.Status == "MONT" {
		if material.CodeExt0 != "" {
			codes = append(codes, &mcom.Code{
				Code:            material.CodeExt0,
				CodeDescription: material.CodeDsc0,
			})
		}
		if material.CodeExt1 != "" {
			codes = append(codes, &mcom.Code{
				Code:            material.CodeExt1,
				CodeDescription: material.CodeDsc1,
			})
		}
	} else {
		if material.CodeExt1 != "" {
			codes = append(codes, &mcom.Code{
				Code:            material.CodeExt1,
				CodeDescription: material.CodeDsc1,
			})
		}
		if material.CodeExt0 != "" {
			codes = append(codes, &mcom.Code{
				Code:            material.CodeExt0,
				CodeDescription: material.CodeDsc0,
			})
		}
	}

	if material.CodeExt2 != "" {
		codes = append(codes, &mcom.Code{
			Code:            material.CodeExt2,
			CodeDescription: material.CodeDsc2,
		})
	}
	if material.CodeExt3 != "" {
		codes = append(codes, &mcom.Code{
			Code:            material.CodeExt3,
			CodeDescription: material.CodeDsc3,
		})
	}
	if material.CodeExt4 != "" {
		codes = append(codes, &mcom.Code{
			Code:            material.CodeExt4,
			CodeDescription: material.CodeDsc4,
		})
	}
	if material.CodeExt5 != "" {
		codes = append(codes, &mcom.Code{
			Code:            material.CodeExt5,
			CodeDescription: material.CodeDsc5,
		})
	}
	if material.CodeExt6 != "" {
		codes = append(codes, &mcom.Code{
			Code:            material.CodeExt6,
			CodeDescription: material.CodeDsc6,
		})
	}
	return codes
}
