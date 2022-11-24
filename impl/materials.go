package impl

import (
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/pda"
)

const errExceededCount = "此物料已達可展延次數(2)"

// GetMaterial implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetMaterial(ctx context.Context, req mcom.GetMaterialRequest) (mcom.GetMaterialReply, error) {
	res, err := dm.pdaService.GetMaterialsInfo(req.MaterialID)
	if err != nil {
		return mcom.GetMaterialReply{}, fmt.Errorf("failed to get materials info, err: %w", err)
	}

	if res == nil {
		return mcom.GetMaterialReply{}, mcomErr.Error{
			Code: mcomErr.Code_RESOURCE_NOT_FOUND,
		}
	}

	if res.ReturnMessage != "" {
		return mcom.GetMaterialReply{}, fmt.Errorf("get material info err: %s", res.ReturnMessage)
	}

	expireDate, err := time.Parse("2006-01-02", res.ExpireDate)
	if err != nil {
		return mcom.GetMaterialReply{}, fmt.Errorf("time parse error: %w", err)
	}

	quantity, err := decimal.NewFromString(strings.TrimSpace(res.Quantity))
	if err != nil {
		return mcom.GetMaterialReply{}, err
	}

	return mcom.GetMaterialReply{
		MaterialProductID: res.MaterialProductID,
		MaterialID:        res.MaterialID,
		MaterialType:      res.MaterialType,
		Sequence:          res.Sequence,
		Status:            res.Status,
		Quantity:          quantity,
		Comment:           res.Comment,
		ExpireDate:        expireDate,
	}, nil
}

// UpdateMaterial implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) UpdateMaterial(ctx context.Context, req mcom.UpdateMaterialRequest) error {
	materialInfo, err := dm.GetMaterial(ctx, mcom.GetMaterialRequest{
		MaterialID: req.MaterialID,
	})
	if err != nil {
		return err
	}

	xmlString, err := xml.Marshal(&pda.SaveChangeRequest{
		MaterialProductID: materialInfo.MaterialProductID,
		MaterialID:        req.MaterialID,
		MaterialType:      materialInfo.MaterialType,
		Quantity:          materialInfo.Quantity.String(),
		DateAdd:           strconv.Itoa(int(req.ExtendedDuration/time.Hour) / 24),
		ExpireDate:        materialInfo.ExpireDate.Format("2006-01-02"),
		User:              req.User,
		OldStatus:         materialInfo.Status,
		NewStatus:         req.NewStatus,
		Reason:            req.Reason,
		ProduceCate:       req.ProductCate,
		Area:              req.ControlArea,
	})
	if err != nil {
		return err
	}

	result, err := dm.pdaService.UpdateMaterials(string(xmlString))
	if err != nil {
		return fmt.Errorf("failed to update materials, err: %w", err)
	}

	if result == errExceededCount {
		return mcomErr.Error{
			Code: mcomErr.Code_RESOURCE_CONTROL_ABOVE_EXTENDED_COUNT,
		}
	}

	if result != "" {
		return fmt.Errorf("failed to update materials, err: %s", result)
	}
	return nil
}

// GetMaterialExtendDate implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetMaterialExtendDate(ctx context.Context, materialID mcom.GetMaterialExtendDateRequest) (mcom.GetMaterialExtendDateReply, error) {
	res, err := dm.pdaService.GetMaterialsInfo(materialID.MaterialID)
	if err != nil {
		return 0, err
	}

	if res == nil {
		return 0, mcomErr.Error{
			Code: mcomErr.Code_RESOURCE_NOT_FOUND,
		}
	}

	if res.ReturnMessage != "" {
		return 0, fmt.Errorf("failed to get changeable status code, err: %s", res.ReturnMessage)
	}

	if strings.TrimSpace(res.SpreadDate) == "" {
		return 0, fmt.Errorf("unable to get expand date of type [%s]", res.MaterialType)
	}

	expandDate, err := strconv.Atoi(strings.TrimSpace(res.SpreadDate))
	if err != nil {
		return 0, err
	}

	return mcom.GetMaterialExtendDateReply(time.Duration(expandDate) * 24 * time.Hour), nil
}
