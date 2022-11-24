package impl

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

// ListCarriers implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListCarriers(ctx context.Context, req mcom.ListCarriersRequest) (mcom.ListCarriersReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListCarriersReply{}, err
	}

	session := dm.newSession(ctx)

	dataCounts, carriers, err := listHandler[models.Carrier](&session, req,
		func(d *gorm.DB) *gorm.DB {
			return d.Model(&models.Carrier{}).Where(&models.Carrier{DepartmentOID: req.DepartmentOID}).Where("deprecated = ?", false)
		})

	if err != nil {
		return mcom.ListCarriersReply{}, err
	}

	res := make([]mcom.CarrierInfo, len(carriers))
	for i, carrier := range carriers {
		res[i] = mcom.NewCarrierInfo(carrier)
	}
	return mcom.ListCarriersReply{
		Info: res,
		PaginationReply: mcom.PaginationReply{
			AmountOfData: dataCounts,
		},
	}, nil
}

// GetCarrier implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetCarrier(ctx context.Context, req mcom.GetCarrierRequest) (mcom.GetCarrierReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetCarrierReply{}, err
	}

	if len(req.ID) != 6 {
		return mcom.GetCarrierReply{}, mcomErr.Error{Code: mcomErr.Code_CARRIER_NOT_FOUND}
	}

	serialNumber, err := strconv.Atoi(req.ID[2:])
	if err != nil {
		return mcom.GetCarrierReply{}, mcomErr.Error{Code: mcomErr.Code_CARRIER_NOT_FOUND}
	}

	session := dm.newSession(ctx)
	var carrier models.Carrier
	if err := session.db.Where(&models.Carrier{
		IDPrefix:     req.ID[:2],
		SerialNumber: int32(serialNumber),
	}).Take(&carrier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcom.GetCarrierReply{}, mcomErr.Error{Code: mcomErr.Code_CARRIER_NOT_FOUND}
		}
		return mcom.GetCarrierReply{}, err
	}

	return mcom.GetCarrierReply(mcom.NewCarrierInfo(carrier)), nil
}

// CreateCarrier implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreateCarrier(ctx context.Context, req mcom.CreateCarrierRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	if err := session.db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.CarrierSerial{CarrierIDPrefix: req.IDPrefix}).Error; err != nil {
		return err
	}

	creator := commonsCtx.UserID(ctx)
	newCarriers := make([]models.Carrier, req.Quantity)
	for i := 0; i < int(req.Quantity); i++ {
		newCarriers[i] = models.Carrier{
			IDPrefix:        req.IDPrefix,
			DepartmentOID:   req.DepartmentOID,
			AllowedMaterial: req.AllowedMaterial,
			CreatedBy:       creator,
			UpdatedBy:       creator,
		}
	}
	if err := session.db.CreateInBatches(&newCarriers, 100).Error; err != nil {
		if IsPqError(err, MaxValueOfSequence) {
			return mcomErr.Error{Code: mcomErr.Code_CARRIER_QUANTITY_LIMIT}
		}
		return err
	}
	return nil
}

// UpdateCarrier implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) UpdateCarrier(ctx context.Context, req mcom.UpdateCarrierRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	idPrefix, serialNumber, err := splitCarrierID(req.ID)
	if err != nil {
		return err
	}

	session := dm.newSession(ctx)

	sqlCommand := session.db.Model(&models.Carrier{}).
		Where(&models.Carrier{IDPrefix: idPrefix, SerialNumber: serialNumber}).
		Where("deprecated = ?", false)

	var carrier models.Carrier
	if err := sqlCommand.Take(&carrier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcomErr.Error{Code: mcomErr.Code_CARRIER_NOT_FOUND}
		}
		return err
	}

	carrier.UpdatedBy = commonsCtx.UserID(ctx)
	switch req.Action.(type) {
	case mcom.UpdateProperties:
		act := req.Action.(mcom.UpdateProperties)
		newDep := act.DepartmentOID
		newAM := act.AllowedMaterial
		if newDep == "" {
			sqlCommand = sqlCommand.Updates(map[string]interface{}{"allowed_material": newAM, "updated_by": commonsCtx.UserID(ctx)})
		} else {
			sqlCommand = sqlCommand.Updates(map[string]interface{}{"allowed_material": newAM, "updated_by": commonsCtx.UserID(ctx), "department_oid": newDep})
		}
	case mcom.BindResources:
		act := req.Action.(mcom.BindResources)
		carrier.Contents = append(carrier.Contents, act.ResourcesID...)
		sqlCommand = sqlCommand.Updates(&models.Carrier{Contents: carrier.Contents, UpdatedBy: carrier.UpdatedBy})
	case mcom.RemoveResources:
		act := req.Action.(mcom.RemoveResources)
		toRemove := make(map[string]struct{})
		for _, res := range act.ResourcesID {
			toRemove[res] = struct{}{}
		}

		newContents := []string{}
		for _, res := range carrier.Contents {
			if _, ok := toRemove[res]; !ok {
				newContents = append(newContents, res)
			}
		}
		carrier.Contents = newContents
		sqlCommand = sqlCommand.Updates(&models.Carrier{Contents: carrier.Contents, UpdatedBy: carrier.UpdatedBy})
	case mcom.ClearResources:
		carrier.Contents = []string{}
		sqlCommand = sqlCommand.Updates(&models.Carrier{Contents: carrier.Contents, UpdatedBy: carrier.UpdatedBy})
	default:
		return fmt.Errorf("undefined update carrier action")
	}

	return sqlCommand.Error
}

func splitCarrierID(id string) (string, int32, error) {
	sn, err := strconv.Atoi(id[2:])
	if err != nil {
		return "", 0, err
	}
	return id[:2], int32(sn), nil
}

// DeleteCarrier implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) DeleteCarrier(ctx context.Context, req mcom.DeleteCarrierRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}
	serialNumber, err := strconv.Atoi(req.ID[2:])
	if err != nil {
		return err
	}

	session := dm.newSession(ctx)
	command := session.db.
		Model(&models.Carrier{IDPrefix: req.ID[:2], SerialNumber: int32(serialNumber)}).
		Where("deprecated = ?", false).
		Updates(map[string]interface{}{
			"deprecated": true,
			"updated_by": commonsCtx.UserID(ctx),
		})
	if command.RowsAffected == 0 {
		return mcomErr.Error{Code: mcomErr.Code_CARRIER_NOT_FOUND}
	}
	return command.Error
}
