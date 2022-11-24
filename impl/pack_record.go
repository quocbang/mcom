package impl

import (
	"context"
	"sort"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

// CreatePackRecord implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreatePackRecords(ctx context.Context, req mcom.CreatePackRecordsRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	packs := make([]models.PackRecord, len(req.Packs))
	for i, pack := range req.Packs {
		packs[i] = models.PackRecord{
			Packing:      pack.Packing,
			PackNumber:   pack.PackNumber,
			Quantity:     pack.Quantity,
			ActualWeight: pack.ActualWeight,
			Station:      req.Station,
			CreatedBy:    commonsCtx.UserID(ctx),
			CreatedAt:    int64(pack.Timestamp),
		}
	}

	session := dm.newSession(ctx)
	return session.db.Create(&packs).Error
}

func (dm *DataManager) ListPackRecords(ctx context.Context) (mcom.ListPackRecordsReply, error) {
	session := dm.newSession(ctx)
	var packs []models.PackRecord
	if err := session.db.Find(&packs).Error; err != nil {
		return mcom.ListPackRecordsReply{}, err
	}

	res := make([]mcom.Pack, len(packs))
	for i, pack := range packs {
		res[i] = mcom.Pack(pack)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].SerialNumber < res[j].SerialNumber
	})
	return mcom.ListPackRecordsReply{
		Packs: res,
	}, nil
}
