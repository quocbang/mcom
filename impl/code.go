package impl

import (
	"context"
	"fmt"

	"gitlab.kenda.com.tw/kenda/mcom"
)

const (
	controlAreaCodeCate   = "LNID"
	controlReasonCodeCate = "MTHL"
)

// ListControlReasons implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListControlReasons(context.Context) (mcom.ListControlReasonsReply, error) {
	res, err := dm.pdaService.GetCodeFromBRM(controlReasonCodeCate)
	if err != nil {
		return mcom.ListControlReasonsReply{}, err
	}

	if res.ReturnMessage != "" {
		return mcom.ListControlReasonsReply{}, fmt.Errorf("failed to get control reason code, tx err: %s", res.ReturnMessage)
	}

	return mcom.ListControlReasonsReply{Codes: codeParser(res)}, nil
}

// ListChangeableStatus implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListChangeableStatus(ctx context.Context, req mcom.ListChangeableStatusRequest) (mcom.ListChangeableStatusReply, error) {
	res, err := dm.pdaService.GetMaterialsInfo(req.MaterialID)
	if err != nil {
		return mcom.ListChangeableStatusReply{}, err
	}

	if res.ReturnMessage != "" {
		return mcom.ListChangeableStatusReply{}, fmt.Errorf("failed to get changeable status code, err: %s", res.ReturnMessage)
	}

	return mcom.ListChangeableStatusReply{Codes: statusParser(res)}, nil
}

// ListControlAreas implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListControlAreas(context.Context) (mcom.ListControlAreasReply, error) {
	res, err := dm.pdaService.GetCodeFromBRM(controlAreaCodeCate)
	if err != nil {
		return mcom.ListControlAreasReply{}, err
	}

	if res.ReturnMessage != "" {
		return mcom.ListControlAreasReply{}, fmt.Errorf("failed to get control area code, err: %s", res.ReturnMessage)
	}

	return mcom.ListControlAreasReply{Codes: codeParser(res)}, nil
}
