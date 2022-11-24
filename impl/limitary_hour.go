package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

func (dm *DataManager) CreateLimitaryHour(ctx context.Context, req mcom.CreateLimitaryHourRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	tx := dm.beginTx(ctx)
	defer tx.Rollback() // nolint: errcheck

	limitaryInsertData := make([]models.LimitaryHour, len(req.LimitaryHour))
	for i, v := range req.LimitaryHour {
		limitaryInfo := models.LimitaryHour{
			ProductType: v.ProductType,
			Min:         v.LimitaryHour.Min,
			Max:         v.LimitaryHour.Max,
		}

		limitaryInsertData[i] = limitaryInfo
	}

	if err := tx.db.Create(&limitaryInsertData).Error; err != nil {
		if IsPqError(err, UniqueViolation) {
			return mcomErr.Error{
				Code:    mcomErr.Code_LIMITARY_HOUR_ALREADY_EXISTS,
				Details: "product-type already exist",
			}
		}
		return err
	}
	return tx.Commit()
}

// GetLimitaryHour implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetLimitaryHour(ctx context.Context, req mcom.GetLimitaryHourRequest) (mcom.GetLimitaryHourReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetLimitaryHourReply{}, err
	}

	var limitaryHour models.LimitaryHour
	session := dm.newSession(ctx)
	if err := session.db.
		Where(&models.LimitaryHour{ProductType: req.ProductType}).
		Take(&limitaryHour).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcom.GetLimitaryHourReply{}, mcomErr.Error{Code: mcomErr.Code_LIMITARY_HOUR_NOT_FOUND}
		}
		return mcom.GetLimitaryHourReply{}, err
	}

	return mcom.GetLimitaryHourReply{
		LimitaryHour: mcom.LimitaryHourParameter{
			Min: limitaryHour.Min,
			Max: limitaryHour.Max,
		},
	}, nil

}
